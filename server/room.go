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
	left   bool
}

// removeCard from the player's hand (returns `true` if the card gets removed).
func (p *Player) removeCard(card Card) bool {
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

// hasSuite checks whether the player has a card matching the suite
// of the given card.
func (p Player) hasSuite(card Card) bool {
	for _, c := range p.hand {
		if c.Suite == card.Suite {
			return true
		}
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

// forgottenPlayer who has left this room at some point.
func (r *Room) forgottenPlayer() (string, *Player) {
	for id, p := range r.players {
		if p.left {
			return id, p
		}
	}

	return "", nil
}

func (r *Room) setDealerForNextRound() *Player {
	highRank := uint8(0)
	player := ""
	for _, c := range r.table {
		r := labelRanks[c.Card.Label]
		if r > highRank {
			highRank = r
			player = c.ID
		}
	}

	// Reset previous dealer.
	for _, p := range r.players {
		p.dealer = false
	}

	newDealer := r.players[player]
	newDealer.dealer = true
	r.currentTurn = newDealer.index
	return newDealer
}

// nextPlayerWithHand returns the player following the given player index
// with cards in their hand.
func (r *Room) nextPlayerWithHand(index uint8) *Player {
	players := make([]string, len(r.players))
	for id, p := range r.players {
		players[p.index] = id
	}

	next := index + 1
	for {
		if next == uint8(len(r.players)) {
			next = 0
		}

		if next == index {
			return nil
		}

		player := r.players[players[next]]
		if len(player.hand) > 0 {
			return player
		}

		next++
	}
}

// addCardToTable for the given player and increment the turn.
// Also dispose the pile if necessary.
func (r *Room) addCardToTable(playerID string, player *Player, card Card) bool {
	r.table = append(r.table, PlayerCard{
		ID:   playerID,
		Card: card,
	})

	if len(r.table) == int(r.limit) {
		r.setDealerForNextRound()
		r.table = make([]PlayerCard, 0)
		return true
	}

	nextPlayer := r.nextPlayerWithHand(player.index)
	if nextPlayer == nil {
		return false
	}

	r.currentTurn = nextPlayer.index
	return true
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

// dealConnectedPlayers through the given WS connection.
// This requires that `room.currentTurn` is set for the next player.
func (r Room) dealConnectedPlayers(ws *websocket.Conn) {
	var turnPlayerID string
	for playerID, p := range r.players {
		if p.index == r.currentTurn {
			turnPlayerID = playerID
		}
	}

	// Send dealt hands to all players after setting up.
	for playerID, p := range r.players {
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: playerID,
			Room:   p.roomID,
			Event:  eventPlayerTurn,
			Response: &DealResponse{
				Hand:       p.hand,
				IsDealer:   p.dealer,
				OurTurn:    r.currentTurn == p.index,
				TurnPlayer: turnPlayerID,
				Table:      r.table,
			},
		})
	}
}

// startGame begins a new game and deals all players.
func (hub *Hub) startGame(ws *websocket.Conn, room *Room) *HandlerError {
	hands := randomDeckChunks(room.limit)
	playerIdx := 0

	for _, p := range room.players {
		p.hand = hands[playerIdx]
		// If player has a spade ace, then they're the dealer.
		for _, card := range p.hand {
			if card.Label == aceSpade.Label && card.Suite == aceSpade.Suite {
				p.dealer = true
				break
			}
		}

		playerIdx++
	}

	for id, p := range room.players {
		if p.dealer {
			return hub.applyPlayerTurn(ws, room, id, p, aceSpade)
		}
	}

	return nil
}

// validateAndApplyTurn from the given player in the given room.
func (hub *Hub) validateAndApplyTurn(ws *websocket.Conn, roomID string, playerID string, data *json.RawMessage) *HandlerError {
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

	// Check whether this is the player's turn.
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

	return hub.applyPlayerTurn(ws, room, playerID, player, req.Card)
}

// FIXME: Cleanup before this spreads.
// applyPlayerTurn (after validation) in the given room using the player and their card.
func (hub *Hub) applyPlayerTurn(ws *websocket.Conn, room *Room, playerID string, player *Player, card Card) *HandlerError {
	// Check whether the player has that card and remove it.
	if !player.removeCard(card) {
		return &HandlerError{
			Msg: "You don't have that card.",
		}
	}

	var gameEnds bool

	// If player has that card, then it's automatically valid. Let's rank stuff.
	if len(room.table) == 0 {
		// Table is empty. If the player isn't the dealer, reject the request.
		if !player.dealer {
			player.hand = append(player.hand, card)
			return &HandlerError{
				Msg: "Only dealers are allowed to start a round.",
			}
		}

		gameEnds = !room.addCardToTable(playerID, player, card)
	} else if room.matchesSuite(card) {
		// Card matches the suites in table.
		gameEnds = !room.addCardToTable(playerID, player, card)
	} else {
		// No match! If the player has that suite and is making an illegal move,
		// reject that request.
		if player.hasSuite(card) {
			player.hand = append(player.hand, card)
			return &HandlerError{
				Msg: "Illegal move. You have a card matching the suite in table.",
			}
		}

		// Player who had the highest rank gets all the junk
		// and becomes the dealer.
		newDealer := room.setDealerForNextRound()
		for _, c := range room.table {
			newDealer.hand = append(newDealer.hand, c.Card)
		}

		room.table = make([]PlayerCard, 0)
		gameEnds = room.nextPlayerWithHand(newDealer.index) == nil
	}

	room.dealConnectedPlayers(ws)
	// Broadcast winning message to all players.
	if len(player.hand) == 0 {
		for _, p := range room.players {
			websocket.JSON.Send(p.conn, &GameMessage{
				Player: playerID,
				Room:   p.roomID,
				Event:  eventPlayerWins,
			})
		}
	}

	// If game has ended, broadcast victim's losing to all players.
	if gameEnds {
		var victimID string
		for id, p := range room.players {
			if len(p.hand) > 0 {
				victimID = id
			}
		}

		for _, p := range room.players {
			websocket.JSON.Send(p.conn, &GameMessage{
				Player: victimID,
				Room:   p.roomID,
				Event:  eventGameOver,
			})
		}
	}

	return nil
}

// Adds player to a room. The room must exist at this point. Also does some sanity
// checks to ensure that some player cannot override someone else's stuff.
func (hub *Hub) addPlayer(ws *websocket.Conn, roomID string, playerID string) *HandlerError {
	swapPlayer := ""
	room, exists := hub.rooms[roomID]
	if !exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s doesn't exist. Feel free to create one!", roomID),
			Event: eventRoomMissing,
		}
	} else if room.isFull() {
		oldID, oldPlayer := room.forgottenPlayer()
		if oldPlayer == nil {
			return &HandlerError{
				Msg:   fmt.Sprintf("Room %s is full. Pick a different room.", roomID),
				Event: eventRoomExists,
			}
		}

		swapPlayer = oldID
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

	if swapPlayer != "" {
		player = room.players[swapPlayer]
		player.conn = ws
		player.left = false
		delete(room.players, swapPlayer)
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

	if swapPlayer != "" {
		room.dealConnectedPlayers(ws)
	} else if room.isFull() {
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
			_, oldPlayer := room.forgottenPlayer()
			if oldPlayer == nil {
				return &HandlerError{
					Msg:   fmt.Sprintf("Room %s already exists and is full. Choose a different name.", roomID),
					Event: eventRoomExists,
				}
			}

			return hub.addPlayer(ws, roomID, playerID)
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
