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
}

// Serve an incoming websocket connection.
func (hub *Hub) serve(ws *websocket.Conn) {
	var playerID string
	for {
		var msg GameMessage
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			hub.dropConn(ws, playerID)
			break
		}

		playerID = strings.ToLower(strings.TrimSpace(msg.Player))
		if playerID == "" {
			log.Println("Ignoring message from anonymous player.")
			continue
		}

		roomID := strings.TrimSpace(msg.Room)
		log.Printf("Incoming message from player %s for room %s\n", playerID, roomID)
		_, roomExists := hub.rooms[roomID]

		var responseErr *HandlerError

		if msg.Event == eventRoomCreate && !roomExists {
			responseErr = hub.createRoomWithPlayer(ws, roomID, playerID, msg.Data)
		} else if msg.Event == eventPlayerJoin && roomExists {
			responseErr = hub.addPlayer(ws, roomID, playerID)
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
func (hub *Hub) dropConn(ws *websocket.Conn, playerID string) {
	log.Printf("Dropping connection for player %s\n", playerID)
	roomID, exists := hub.connRooms[ws]
	if !exists {
		return
	}

	room := hub.rooms[roomID]
	log.Printf("Removing player %s from room %s\n", playerID, roomID)
	delete(room.players, playerID)
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
	Players uint8 `json:"players"`
}

// RoomResponse from the server.
type RoomResponse struct {
	Players []string `json:"players"`
}
