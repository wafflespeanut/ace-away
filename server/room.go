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

// removeCard from the player's hand (returns `true` if the card gets removed).
func (p Player) removeCard(card Card) bool {
	cardIdx := -1
	for i, c := range p.hand {
		if c.Label == card.Label && c.Suite == card.Suite {
			cardIdx = i
			break
		}
	}

	if cardIdx >= 0 {
		p.hand[cardIdx] = p.hand[len(p.hand)-1]
		p.hand = p.hand[:len(p.hand)-1]
		return true
	}

	return false
}

// Room containing some players.
type Room struct {
	players     map[string]*Player
	currentTurn uint8
	limit       uint8
	table       []PlayerCard
}

// nextPlayer returns the player following the given player index.
func (r Room) nextPlayer(index uint8) (string, *Player) {
	next := index + 1
	if next == uint8(len(r.players)) {
		next = 0
	}

	for id, p := range r.players {
		if p.index == next {
			return id, p
		}
	}

	return "", nil
}

// isFull checks whether this room is full.
func (r Room) isFull() bool {
	return len(r.players) == int(r.limit)
}

// playerIDs returns the IDs of players in this room.
func (r Room) playerIDs() []string {
	players := make([]string, 0, len(r.players))
	for k := range r.players {
		players = append(players, k)
	}

	return players
}

// matchesSuite checks whether all cards in the table matches
// the given card's suite.
func (r Room) matchesSuite(card Card) bool {
	matches := true
	for _, c := range r.table {
		matches = matches && c.Card.Suite == card.Suite
	}

	return matches
}

// startGame begins a new game and deals all players.
func (hub *Hub) startGame(ws *websocket.Conn, room *Room) *HandlerError {
	hands := randomDeckChunks(room.limit)
	playerIdx := 0
	var turnPlayerID string
	var turnPlayer *Player

	for playerID, p := range room.players {
		p.hand = hands[playerIdx]
		aceIndex := -1
		// If player has a spade ace, then they're the dealer.
		for cardIdx, card := range p.hand {
			if card.Label == "A" && card.Suite == "s" {
				p.dealer = true
				aceIndex = cardIdx
				break
			}
		}

		if p.dealer {
			// set current turn
			turnPlayerID, turnPlayer = room.nextPlayer(p.index)
			room.currentTurn = turnPlayer.index
			// Add ace spade to table
			room.table = append(room.table, PlayerCard{
				ID:   playerID,
				Card: p.hand[aceIndex],
			})
			// swap remove ace spade.
			p.hand[aceIndex] = p.hand[len(p.hand)-1]
			p.hand = p.hand[:len(p.hand)-1]
		}

		playerIdx++
	}

	// Send dealt hands to all players after setting up.
	for playerID, p := range room.players {
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: playerID,
			Room:   p.roomID,
			Event:  eventPlayerTurn,
			Response: &DealResponse{
				Hand:       p.hand,
				IsDealer:   p.dealer,
				OurTurn:    room.currentTurn == p.index,
				TurnPlayer: turnPlayerID,
				Table:      room.table,
			},
		})
	}

	return nil
}

func (hub *Hub) applyPlayerTurn(ws *websocket.Conn, roomID string, playerID string, data *json.RawMessage) *HandlerError {
	room, exists := hub.rooms[roomID]
	if !exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s doesn't exist. Restart the game by creating a new room.", roomID),
			Event: eventRoomMissing,
		}
	}

	player, exists := room.players[playerID]
	if !exists {
		return &HandlerError{
			Msg: fmt.Sprintf("You don't belong in room %s. Please join the room first.", roomID),
		}
	}

	// Check whether this is player's turn.
	if player.index != room.currentTurn {
		return &HandlerError{
			Msg: "It's not your turn yet.",
		}
	}

	var req TurnRequest
	err := json.Unmarshal(*data, &req)
	if err != nil {
		return &HandlerError{
			Msg: "Invalid request for player's turn.",
		}
	}

	// Check whether the player has that card and remove it.
	if !player.removeCard(req.Card) {
		return &HandlerError{
			Msg: "You don't have that card.",
		}
	}

	// If player has that card, then it's automatically valid. Let's rank stuff.
	if len(room.table) == 0 {
		// Table is empty.
	} else if room.matchesSuite(req.Card) {
		// Card matches the suites in table.
	} else {
		// No match! Dealer gets all the junk.
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
	} else if room.isFull() {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s is full. Pick a different room.", roomID),
			Event: eventRoomExists,
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
		table:   make([]PlayerCard, 0),
	}

	hub.rooms[roomID] = room
	return hub.addPlayer(ws, roomID, playerID)
}
