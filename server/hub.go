package main

import (
	"encoding/json"
	"log"

	"golang.org/x/net/websocket"
)

// Hub contains a map of websocket connections and the associated user IDs.
type Hub struct {
	rooms     map[string]*Room
	connRooms map[*websocket.Conn]string
}

// Serve an incoming websocket connection.
func (hub *Hub) serve(ws *websocket.Conn) {
	var userID string
	for {
		var msg GameMessage
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			hub.dropConn(ws, userID)
			break
		}

		userID = msg.User
		if userID == "" {
			continue
		}

		roomID := msg.Room
		_, roomExists := hub.rooms[roomID]
		if msg.Event == eventRoomCreate && !roomExists {
			hub.createRoom(ws, roomID, userID, msg.Data)
		} else if msg.Event == eventUserJoin && roomExists {
			user := hub.addUser(ws, roomID, userID)
			if user == nil {
				return
			}

			//
		}
	}
}

// Cleanup and drop a connection.
func (hub *Hub) dropConn(ws *websocket.Conn, userID string) {
	log.Printf("Dropping connection for user %s\n", userID)
}

// Models

// GameMessage represents a message through the websocket.
type GameMessage struct {
	User  string           `json:"user"`
	Room  string           `json:"room"`
	Event string           `json:"event"`
	Data  *json.RawMessage `json:"data"`
}

// RoomCreationRequest for creating a room.
type RoomCreationRequest struct {
	Players uint8 `json:"players"`
}
