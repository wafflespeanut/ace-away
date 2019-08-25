package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/websocket"
)

// Player represents a player with an active websocket connection.
// A player can belong to one room at most.
type Player struct {
	conn *websocket.Conn
	// ID of the room to which this player belongs.
	roomID string
	// Player's hand containing cards.
	hand []Card
	// Whether this player is the dealer for some round.
	dealer bool
	// Index of this player (i.e., for turns).
	index uint8
	// Whether this player has left this room and has been disabled.
	// If they have, then another player can take their place.
	left bool
}

// debugString for `Player`
func (p *Player) debugString() string {
	return spew.Sprintf("Room ID: %+v\nHand: %+v\nisDealer: %+v\nindex: %+v\nhasLeft: %+v\n",
		p.roomID, p.hand, p.dealer, p.index, p.left)
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

// containsSuite checks whether the player has a card matching the suite
// of the given card.
func (p Player) containsSuite(card Card) *Card {
	for _, c := range p.hand {
		if c.Suite == card.Suite {
			return &c
		}
	}

	return nil
}

// Room containing some players.
type Room struct {
	// Lock so that only one connection can persist stuff at a time.
	lock sync.Mutex
	// Map of player IDs to their meta info.
	players map[string]*Player
	// Index of the player taking the current turn.
	currentTurn uint8
	// Max number of players allowed in this room.
	limit uint8
	// Table containing player IDs and their cards for this round.
	table []PlayerCard
}

// `debugString` for `Room`.
func (r *Room) debugString() string {
	s := "Players:\n"
	for id, p := range r.players {
		s += spew.Sprintf("%#+v: %s\n", id, p.debugString())
	}

	s += spew.Sprintf("Current turn: %+v\nLimit: %+v\nTable: %+v\n",
		r.currentTurn, r.limit, r.table)
	return s
}

// tableReachedLimit returns whether the table has cards from all players
// with cards in their hands, indicating the end of a round.
func (r *Room) tableReachedLimit() bool {
	l := len(r.table)
	for _, p := range r.players {
		if len(p.hand) > 0 {
			l--
		}
	}

	return l == 0
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

// setDealerForNextRound resets previous dealers, gets the player
// who has submitted the highest ranked card and marks them as dealer.
// Also updates the room's `currentTurn` with that player's index.
func (r *Room) setDealerForNextRound() (string, *Player) {
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
	// log.Printf("New dealer: %s (index: %d, old index: %d)\n",
	// 	player, newDealer.index, r.currentTurn)
	r.currentTurn = newDealer.index
	return player, newDealer
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
			next = 0 // wrap around
		}

		if next == index {
			return nil // no one's there
		}

		player := r.players[players[next]]
		if len(player.hand) > 0 {
			// log.Printf("Next player with hand: %s (index: %d)\n", players[next], player.index)
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

	// If table has reached its limit, then we can set the dealer and
	// begin the next round.
	if r.tableReachedLimit() {
		// log.Println("Table reached limit. Setting dealer for next round.")
		r.setDealerForNextRound()
		r.table = make([]PlayerCard, 0)
		return true
	}

	nextPlayer := r.nextPlayerWithHand(player.index)
	if nextPlayer == nil {
		return false
	}

	// log.Printf("Switching to player %d\n", nextPlayer.index+1)
	r.currentTurn = nextPlayer.index
	return true
}

// isFull checks whether this room is full.
func (r *Room) isFull() bool {
	return len(r.players) == int(r.limit)
}

// playerIDs returns the IDs of players in this room
// in the order they'd joined.
func (r *Room) playerIDs() []string {
	players := make([]string, len(r.players))
	for k, p := range r.players {
		players[p.index] = k
	}

	return players
}

// winnerIDs indicate players who have successfully gotten rid
// of all their cards.
func (r *Room) winnerIDs() []string {
	players := make([]string, 0)
	for k, p := range r.players {
		if len(p.hand) == 0 {
			players = append(players, k)
		}
	}

	return players
}

// matchesSuite checks whether all cards in the table matches
// the given card's suite.
func (r *Room) matchesSuite(card Card) bool {
	matches := true
	for _, c := range r.table {
		matches = matches && c.Card.Suite == card.Suite
	}

	return matches
}

// dealConnectedPlayers through the given WS connection.
// This requires that `room.currentTurn` is set for the next player.
func (r *Room) dealConnectedPlayers(ws *websocket.Conn) {
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

// validateAndApplyTurn from the given player in the given room.
func (hub *Hub) validateAndApplyTurn(ws *websocket.Conn, roomID string, playerID string, data *json.RawMessage) *HandlerError {
	room, exists := hub.getRoom(roomID)
	if !exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s doesn't exist. Restart the game by creating a new room.", roomID),
			Event: eventRoomMissing,
		}
	}

	room.lock.Lock()
	defer room.lock.Unlock()

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

	gameEnds, e := hub.applyPlayerTurn(room, playerID, req.Card)
	if e != nil {
		return e
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

// applyPlayerTurn (after validation) in the given room using the player and their card.
//
// **NOTE:** The caller is responsible for synchronizing access to room pointer.
func (hub *Hub) applyPlayerTurn(room *Room, playerID string, card Card) (bool, *HandlerError) {
	// log.Printf("Before update: %s", room.debugString())
	player := room.players[playerID]
	// Check whether the player has that card and remove it.
	if !player.removeCard(card) {
		return false, &HandlerError{
			Msg: "You don't have that card.",
		}
	}

	var gameEnds bool

	// If player has that card, then it's automatically valid. Let's rank stuff.
	if len(room.table) == 0 {
		// Table is empty. If the player isn't the dealer, reject the request.
		if !player.dealer {
			player.hand = append(player.hand, card)
			return false, &HandlerError{
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
		matchedCard := player.containsSuite(room.table[0].Card)
		if matchedCard != nil {
			player.hand = append(player.hand, card)
			return false, &HandlerError{
				Msg: fmt.Sprintf("Illegal move. You have %s%s which matches the suite in table.",
					matchedCard.Label, prettyMap[matchedCard.Suite]),
			}
		}

		// Player who had the highest rank gets all the junk
		// and becomes the dealer.
		_, newDealer := room.setDealerForNextRound()
		// log.Printf("New dealer: %s -> %s", dealerID, newDealer.debugString())
		newDealer.hand = append(newDealer.hand, card)
		for _, c := range room.table {
			newDealer.hand = append(newDealer.hand, c.Card)
		}

		room.table = make([]PlayerCard, 0)
		gameEnds = room.nextPlayerWithHand(newDealer.index) == nil
	}

	// log.Printf("After update: %s", room.debugString())
	return gameEnds, nil
}

// Adds player to a room. The room must exist at this point. Also does some sanity
// checks to ensure that some player cannot override someone else's stuff.
func (hub *Hub) addPlayer(ws *websocket.Conn, roomID string, playerID string) *HandlerError {
	swapPlayer := ""
	room, exists := hub.getRoom(roomID)
	if !exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s doesn't exist. Feel free to create one!", roomID),
			Event: eventRoomMissing,
		}
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	if room.isFull() {
		oldID, oldPlayer := room.forgottenPlayer()
		if oldPlayer == nil {
			return &HandlerError{
				Msg:   fmt.Sprintf("Room %s is full. Pick a different room.", roomID),
				Event: eventRoomExists,
			}
		}

		// This player can take the place of an old player.
		swapPlayer = oldID
	}

	hub.setConnection(ws, roomID)

	_, exists = room.players[playerID]
	if exists && swapPlayer == "" {
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
				Escaped: room.winnerIDs(),
				Max:     room.limit,
				TurnIdx: room.currentTurn,
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

// startGame begins a new game and deals all players.
//
// **NOTE:** The caller is responsible for synchronizing access to room pointer.
func (hub *Hub) startGame(ws *websocket.Conn, room *Room) *HandlerError {
	hands := randomDeckChunks(room.limit)
	playerIdx := 0

	for _, p := range room.players {
		p.hand = hands[playerIdx]
		// If player has a spade ace, then they're the dealer.
		for _, card := range p.hand {
			if card.Label == aceSpade.Label && card.Suite == aceSpade.Suite {
				p.dealer = true
				room.currentTurn = p.index
				break
			}
		}

		playerIdx++
	}

	room.dealConnectedPlayers(ws)
	return nil
}

// Creates a room with the given data and adds the player to that room.
func (hub *Hub) createRoomWithPlayer(ws *websocket.Conn, roomID string, playerID string, data *json.RawMessage) *HandlerError {
	for {
		room, exists := hub.getRoom(roomID)
		if roomID == "" {
			roomID = randSeq(16)
			continue
		} else if exists {
			room.lock.Lock()
			if room.isFull() {
				_, oldPlayer := room.forgottenPlayer()
				if oldPlayer == nil {
					return &HandlerError{
						Msg:   fmt.Sprintf("Room %s already exists and is full. Choose a different name.", roomID),
						Event: eventRoomExists,
					}
				}
			}

			room.lock.Unlock()
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

	hub.setRoom(roomID, room)
	return hub.addPlayer(ws, roomID, playerID)
}
