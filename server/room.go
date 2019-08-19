package main

import (
	"encoding/json"
	"fmt"

	"golang.org/x/net/websocket"
)

// Player represents a player with an active websocket connection.
// A player can belong to one room at most.
type Player struct {
	conn   *websocket.Conn
	roomID string
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
