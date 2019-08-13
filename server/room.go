package main

import (
	"encoding/json"

	"golang.org/x/net/websocket"
)

// Player represents a player with an active websocket connection.
type Player struct {
	conn *websocket.Conn
}

// Room containing some players.
type Room struct {
	users map[string]*Player
	limit uint8
}

// AddUser to a room. The room must exist at this point. Also does some sanity
// checks to ensure that some user cannot override someone else's stuff.
func (hub *Hub) addUser(ws *websocket.Conn, roomID string, userID string) *Player {
	_, exists := hub.connRooms[ws]
	if !exists {
		hub.connRooms[ws] = roomID
	}

	room := hub.rooms[roomID]
	_, exists = room.users[userID]
	if exists {
		// User already exists in this room. Don't create another.
		return nil
	}

	user := &Player{
		conn: ws,
	}

	room.users[userID] = user
	return user
}

func (hub *Hub) createRoom(ws *websocket.Conn, roomID string, userID string, data *json.RawMessage) {
	// Create a unique room ID.
	if roomID == "" {
		for {
			roomID = randSeq(16)
			_, exists := hub.rooms[roomID]
			if exists {
				return
			}
		}
	}

	var req RoomCreationRequest
	err := json.Unmarshal(*data, &req)
	if err != nil || req.Players < 3 {
		return
	}

	room := &Room{
		users: make(map[string]*Player),
		limit: req.Players,
	}

	hub.rooms[roomID] = room
}
