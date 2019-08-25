package main

import (
	"encoding/json"
	"log"
	"strings"

	"golang.org/x/net/websocket"
)

// Hub contains a map of websocket connections and the associated player IDs.
type Hub struct {
	// Map of room IDs to actual room objects.
	rooms map[string]*Room
	// Map of WS connections to room IDs.
	connRooms map[*websocket.Conn]string
	// Hub command channel.
	cmdChan chan hubCommand
	// Channel for room pointers from hub.
	roomChan chan *Room
	// Channel for IDs from hub.
	connChan chan string
	// Ack channel for other operations.
	ackChan chan bool
}

type hubCmdType int

const (
	cmdSetRoom = iota
	cmdGetRoom
	cmdDeleteRoom
	cmdSetConnection
	cmdDeleteConnection
)

// Command sent for requesting/updating stuff in the hub.
type hubCommand struct {
	// Type of the command (mandatory).
	ty     hubCmdType
	roomID string
	room   *Room
	ws     *websocket.Conn
}

/* Map-like methods specific to our types. */

// getRoom corresponding to the given room ID.
func (hub *Hub) getRoom(roomID string) (*Room, bool) {
	hub.cmdChan <- hubCommand{
		ty:     cmdGetRoom,
		roomID: roomID,
	}

	room := <-hub.roomChan
	return room, room != nil
}

// setRoom for the given ID.
func (hub *Hub) setRoom(roomID string, room *Room) {
	hub.cmdChan <- hubCommand{
		ty:     cmdSetRoom,
		roomID: roomID,
		room:   room,
	}
	_ = <-hub.ackChan
}

// deleteRoom corresponding to the given room ID.
func (hub *Hub) deleteRoom(roomID string) {
	hub.cmdChan <- hubCommand{
		ty:     cmdDeleteRoom,
		roomID: roomID,
	}
	_ = <-hub.ackChan
}

// setConnection to the given room ID.
func (hub *Hub) setConnection(ws *websocket.Conn, roomID string) {
	hub.cmdChan <- hubCommand{
		ty:     cmdSetConnection,
		roomID: roomID,
		ws:     ws,
	}
	_ = <-hub.ackChan
}

// deleteConnection and return its room ID (if any).
func (hub *Hub) deleteConnection(ws *websocket.Conn) (string, bool) {
	hub.cmdChan <- hubCommand{
		ty: cmdDeleteConnection,
		ws: ws,
	}
	id := <-hub.connChan
	return id, id != ""
}

// handleCommands for this hub.
//
// **NOTE:** This must be launched into a separate goroutine.
func (hub *Hub) handleCommands() {
	for {
		cmd := <-hub.cmdChan
		if cmd.ty == cmdGetRoom {
			room := hub.rooms[cmd.roomID]
			hub.roomChan <- room
		} else if cmd.ty == cmdSetRoom {
			hub.rooms[cmd.roomID] = cmd.room
			hub.ackChan <- true
		} else if cmd.ty == cmdDeleteRoom {
			_, exists := hub.rooms[cmd.roomID]
			delete(hub.rooms, cmd.roomID)
			hub.ackChan <- exists
		} else if cmd.ty == cmdSetConnection {
			hub.connRooms[cmd.ws] = cmd.roomID
			hub.ackChan <- true
		} else if cmd.ty == cmdDeleteConnection {
			roomID, exists := hub.connRooms[cmd.ws]
			if !exists {
				roomID = ""
			}

			delete(hub.connRooms, cmd.ws)
			hub.connChan <- roomID
		}
	}
}

// Serve an incoming websocket connection.
func (hub *Hub) serve(ws *websocket.Conn) {
	var playerID string
	for {
		var msg GameMessage
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			hub.dropPlayer(ws, playerID)
			break
		}

		playerID = strings.ToLower(strings.TrimSpace(msg.Player))
		if playerID == "" {
			log.Println("Ignoring message from anonymous player.")
			continue
		}

		roomID := strings.TrimSpace(msg.Room)
		log.Printf("Event %s from player %s for room %s\n", msg.Event, playerID, roomID)

		var responseErr *HandlerError

		if msg.Event == eventRoomCreate {
			responseErr = hub.createRoomWithPlayer(ws, roomID, playerID, msg.Data)
		} else if msg.Event == eventPlayerJoin {
			responseErr = hub.addPlayer(ws, roomID, playerID)
		} else if msg.Event == eventPlayerTurn {
			responseErr = hub.validateAndApplyTurn(ws, roomID, playerID, msg.Data)
		}

		if responseErr != nil {
			websocket.JSON.Send(ws, &GameMessage{
				Event: responseErr.Event,
				Msg:   responseErr.Msg,
			})
		}
	}
}

// Cleanup and drop a connection.
func (hub *Hub) dropPlayer(ws *websocket.Conn, playerID string) {
	log.Printf("Dropping connection for player %s\n", playerID)
	roomID, exists := hub.deleteConnection(ws)
	if !exists {
		return
	}

	room, _ := hub.getRoom(roomID)
	log.Printf("Disabling player %s in room %s\n", playerID, roomID)
	room.lock.Lock()
	defer room.lock.Unlock()

	room.players[playerID].left = true

	allLeft := true
	for _, p := range room.players {
		allLeft = allLeft && p.left
	}

	if allLeft {
		log.Printf("All players have left. Removing room %s\n", roomID)
		hub.deleteRoom(roomID)
	}
}

// Models

// HandlerError is the interface for all server reactions.
type HandlerError struct {
	Event string `json:"event"`
	Msg   string `json:"msg"`
}

// GameMessage represents a message through the websocket.
type GameMessage struct {
	Player   string           `json:"player"`
	Room     string           `json:"room"`
	Event    string           `json:"event"`
	Data     *json.RawMessage `json:"data"`
	Response interface{}      `json:"response"`
	Msg      string           `json:"msg"`
}

// RoomCreationRequest from the client for creating a room.
type RoomCreationRequest struct {
	// Number of players to be allowed in that room.
	Players uint8 `json:"players"`
}

// TurnRequest for a player's attempt at submitting a card.
type TurnRequest struct {
	// Card submitted by the player in some round.
	Card Card `json:"card"`
}

// RoomResponse from the server.
type RoomResponse struct {
	// IDs of players in the room.
	Players []string `json:"players"`
	// When a player takes place of another player who has left the room, it should
	// be possible to show the winner(s) in the room (if any).
	Escaped []string `json:"escaped"`
	// Max number of players allowed for this room.
	Max uint8 `json:"max"`
	// Index of the player taking the current turn.
	TurnIdx uint8 `json:"turnIdx"`
}

// DealResponse from the server when the game begins.
type DealResponse struct {
	// Table containing IDs of players and the cards submitted by them for some round.
	Table []PlayerCard `json:"table"`
	// Hand of the player getting this response.
	Hand []Card `json:"hand"`
	// Whether this player is the dealer for this round.
	IsDealer bool `json:"isDealer"`
	// Whether this is the receiving player's turn.
	OurTurn bool `json:"ourTurn"`
	// Whose turn is this?
	TurnPlayer string `json:"turnPlayer"`
}

// Card from a deck.
type Card struct {
	Label string `json:"label"`
	Suite string `json:"suite"`
}

// PlayerCard containing card with player ID.
type PlayerCard struct {
	ID   string `json:"id"`
	Card Card   `json:"card"`
}
