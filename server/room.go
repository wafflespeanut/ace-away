package main

import (
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

// Player represents a player with an active websocket connection.
// A player can belong to one room at most.
type Player struct {
	conn   *websocket.Conn
	roomID string
	hand   []Card
	dealer bool
	index  uint8
}

// Room containing some players.
type Room struct {
	players map[string]*Player
	limit   uint8
}

// Checks whether this room is full.
func (r Room) isFull() bool {
	return len(r.players) == int(r.limit)
}

// Returns the IDs of players in this room.
func (r Room) playerIDs() []string {
	players := make([]string, 0, len(r.players))
	for k := range r.players {
		players = append(players, k)
	}

	return players
}

// We're beginning a new game. Deal all players.
func (hub *Hub) startGame(ws *websocket.Conn, room *Room) *HandlerError {
	hands := randomDeckChunks(room.limit)
	i := 0
	for playerID, p := range room.players {
		p.hand = hands[i]
		// If player has a spade ace, then they're the dealer.
		for _, card := range p.hand {
			if card.Label == "A" && card.Suite == "s" {
				p.dealer = true
				break
			}
		}

		i++

		// Send dealt hands to all players.
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: playerID,
			Room:   p.roomID,
			Event:  eventGameBegins,
			Response: &DealResponse{
				Hand:     p.hand,
				IsDealer: p.dealer,
			},
		})
	}

	return nil
}

// Adds player to a room. The room must exist at this point. Also does some sanity
// checks to ensure that some player cannot override someone else's stuff.
func (hub *Hub) addPlayer(ws *websocket.Conn, roomID string, playerID string) *HandlerError {
	room, exists := hub.rooms[roomID]
	if !exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s doesn't exist. Feel free to create one!", roomID),
			Event: eventRoomMissing,
		}
	}

	_, exists = hub.connRooms[ws]
	if !exists {
		hub.connRooms[ws] = roomID
	}

	_, exists = room.players[playerID]
	if exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Player %s already exists in room %s. Choose a different name.", playerID, roomID),
			Event: eventPlayerExists,
		}
	}

	player := &Player{
		conn:   ws,
		roomID: roomID,
		hand:   make([]Card, 0),
		index:  uint8(len(room.players)),
	}

	room.players[playerID] = player
	for _, p := range room.players {
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: playerID,
			Room:   roomID,
			Event:  eventPlayerJoin,
			Response: &RoomResponse{
				Players: room.playerIDs(),
				Max:     room.limit,
			},
		})
	}

	if room.isFull() {
		log.Printf("Room %s is full. Starting a new game.\n", roomID)
		return hub.startGame(ws, room)
	}

	return nil
}

// Creates a room with the given data and adds the player to that room.
func (hub *Hub) createRoomWithPlayer(ws *websocket.Conn, roomID string, playerID string, data *json.RawMessage) *HandlerError {
	for {
		room, exists := hub.rooms[roomID]
		if roomID == "" {
			roomID = randSeq(16)
			continue
		} else if exists && room.isFull() {
			return &HandlerError{
				Msg:   fmt.Sprintf("Room %s already exists and is full. Choose a different name.", roomID),
				Event: eventRoomExists,
			}
		} else if exists {
			return hub.addPlayer(ws, roomID, playerID)
		} else {
			break
		}
	}

	var req RoomCreationRequest
	err := json.Unmarshal(*data, &req)
	if err != nil {
		return &HandlerError{
			Msg: "Invalid request for creating room.",
		}
	}

	if req.Players < minPlayers || req.Players > maxPlayers {
		return &HandlerError{
			Msg: fmt.Sprintf("Only %d-%d players are allowed.", minPlayers, maxPlayers),
		}
	}

	room := &Room{
		players: make(map[string]*Player),
		limit:   req.Players,
	}

	hub.rooms[roomID] = room
	return hub.addPlayer(ws, roomID, playerID)
}
