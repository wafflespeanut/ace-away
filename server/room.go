package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/websocket"
)

type turnEffect int

const (
	turnApplied = iota
	turnFailed
	tableFull
	gameEnds
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
	// Whether this player has exited this room after getting rid
	// of all of their cards.
	exited bool
	// Whether this player has requested a restart.
	requestedRestart bool
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
		// Swap remove.
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
	// The player who lost in the previous round.
	previousAcePlayer *Player
	// If the player who lost in the previous round loses again,
	// then the start accumulating high rank cards. This is reset
	// when another player loses.
	acePlayerCollection []Card
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

// endRound in this room by clearing the table. If a player
// doesn't have any card in their hand, then the player is marked
// "exited" and their ID is returned.
func (r *Room) endRound() []string {
	playerIDs := make([]string, 0)

	r.table = make([]PlayerCard, 0)
	for id, p := range r.players {
		if len(p.hand) == 0 && !p.exited {
			p.exited = true
			playerIDs = append(playerIDs, id)
		}
	}

	return playerIDs
}

// tableReachedLimit returns whether the table has cards from all players
// with at least one card in their hands, indicating the end of a round.
func (r *Room) tableReachedLimit() bool {
	l := len(r.table)
	for _, p := range r.players {
		if !p.exited {
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
		if p.exited {
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

// Number of players who have issued a request for restarting the game.
// If majority have, then a restart is issued.
func (r *Room) restartRequests() uint8 {
	count := uint8(0)
	for _, p := range r.players {
		if p.requestedRestart {
			count++
		}
	}

	return count
}

// startGame clears the table, begins a new game and deals all players.
// If this room has an ongoing game, then it finds the player
// who hasn't "exited", and hands them high rank card(s) depending
// on how many times they've lost.
func (r *Room) startGame() {
	r.table = make([]PlayerCard, 0)
	aceCount := 0
	var acePlayer *Player
	for _, p := range r.players {
		if !p.exited {
			aceCount++
			acePlayer = p
		}
	}

	aceExistedBefore := r.previousAcePlayer != nil
	aceExistsNow := acePlayer != nil && aceCount == 1
	isAcePlayerNew := aceExistedBefore && aceExistsNow && r.previousAcePlayer.index != acePlayer.index

	if aceExistsNow {
		if !aceExistedBefore || isAcePlayerNew {
			// If we have an ace and if it's either first time or it's for a different player,
			// then reset with an ace spade.
			r.acePlayerCollection = []Card{aceSpade}
		} else if aceExistedBefore && !isAcePlayerNew {
			// If it's the same player getting an ace, then whack them with another high card.
			nextCard := getNextAceCard(r.acePlayerCollection[len(r.acePlayerCollection)-1])
			if nextCard != nil {
				r.acePlayerCollection = append(r.acePlayerCollection, *nextCard)
			}
		}

		r.previousAcePlayer = acePlayer
	}

	hands := randomDeckChunks(r.limit, r.acePlayerCollection)
	if aceExistsNow || aceExistedBefore {
		idx := r.previousAcePlayer.index
		// This ensures that the lost player gets the high rank card(s) again.
		hands[0], hands[idx] = hands[idx], hands[0]
	}

	for _, p := range r.players {
		p.exited = false
		p.hand = hands[p.index]
		// If player has a spade ace, then they're the dealer.
		for _, card := range p.hand {
			if card.Label == aceSpade.Label && card.Suite == aceSpade.Suite {
				p.dealer = true
				r.currentTurn = p.index
			}
		}
	}
}

// dealConnectedPlayers through the given WS connection.
// This requires that `room.currentTurn` is set for the next player.
func (r *Room) dealConnectedPlayers(ws *websocket.Conn) {
	var turnPlayerID string
	for playerID, p := range r.players {
		// Reset restart request for players.
		p.requestedRestart = false
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
func (hub *Hub) validateAndApplyTurn(ws *websocket.Conn, roomID, playerID string, data *json.RawMessage) *HandlerError {
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

	turnEffect, e := hub.applyPlayerTurn(room, playerID, req.Card)
	if e != nil {
		return e
	}

	if turnEffect == tableFull {
		// Notify players before clearing the table.
		room.dealConnectedPlayers(ws)
		// log.Println("Table reached limit. Setting dealer for next round.")
		winnerIDs := room.endRound()
		// Broadcast winning message to all players at the end of a round.
		if len(winnerIDs) > 0 {
			for _, winnerID := range winnerIDs {
				for _, p := range room.players {
					websocket.JSON.Send(p.conn, &GameMessage{
						Player: winnerID,
						Room:   p.roomID,
						Event:  eventPlayerWins,
					})
				}
			}
		}

		// By the end of each round, check if the game has ended.
		if room.nextPlayerWithHand(room.currentTurn) == nil {
			turnEffect = gameEnds
		}
	}

	room.dealConnectedPlayers(ws)

	// If game has ended, broadcast victim's losing to all players.
	if turnEffect == gameEnds {
		// There could be multiple winners, in which case, `victimID` would be an empty string.
		var victimID string
		for id, p := range room.players {
			if len(p.hand) > 0 {
				victimID = id
			} else {
				// Set the exit status of players without any cards. This is an off-by-one
				// case which happens when the last turn involves a player dumping their
				// last card to their opponent and winning the game.
				p.exited = true
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
func (hub *Hub) applyPlayerTurn(room *Room, playerID string, card Card) (turnEffect, *HandlerError) {
	// log.Printf("Before update: %s", room.debugString())
	player := room.players[playerID]
	// Check whether the player has that card and remove it.
	if !player.removeCard(card) {
		return turnFailed, &HandlerError{
			Msg: "You don't have that card.",
		}
	}

	// If player has that card, then it's automatically valid. Let's rank stuff.
	if len(room.table) == 0 {
		// Table is empty. If the player isn't the dealer, reject the request.
		if !player.dealer {
			player.hand = append(player.hand, card)
			return turnFailed, &HandlerError{
				Msg: "Only dealers are allowed to start a round.",
			}
		}

		if !room.addCardToTable(playerID, player, card) {
			return gameEnds, nil
		}
	} else if room.matchesSuite(card) {
		// Card matches the suites in table.
		if !room.addCardToTable(playerID, player, card) {
			return gameEnds, nil
		}

		// If table has reached its limit, then we can set the dealer and
		// begin the next round.
		if room.tableReachedLimit() {
			room.setDealerForNextRound()
			return tableFull, nil
		}
	} else {
		// No match! If the player has that suite and is making an illegal move,
		// reject that request.
		matchedCard := player.containsSuite(room.table[0].Card)
		if matchedCard != nil {
			player.hand = append(player.hand, card)
			return turnFailed, &HandlerError{
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

		// Temporarily add the card to table. Table will be cleared by the caller later.
		room.table = append(room.table, PlayerCard{
			ID:   playerID,
			Card: card,
		})

		if room.nextPlayerWithHand(newDealer.index) == nil {
			return gameEnds, nil
		}

		return tableFull, nil
	}

	// log.Printf("After update: %s", room.debugString())
	return turnApplied, nil
}

// Adds player to a room. The room must exist at this point. Also does some sanity
// checks to ensure that some player cannot override someone else's stuff.
func (hub *Hub) addPlayer(ws *websocket.Conn, roomID, playerID string) *HandlerError {
	room, exists := hub.getRoom(roomID)
	if !exists {
		return &HandlerError{
			Msg:   fmt.Sprintf("Room %s doesn't exist. Feel free to create one!", roomID),
			Event: eventRoomMissing,
		}
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	return hub.addPlayerToUnlockedRoom(ws, room, roomID, playerID)
}

// addPlayerToUnlockedRoom accepts an unlocked room and does whatever `addPlayer` method says.
// The method has been split so as to avoid a possible race condition.
func (hub *Hub) addPlayerToUnlockedRoom(ws *websocket.Conn, room *Room, roomID, playerID string) *HandlerError {
	swapPlayer := ""

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

	_, exists := room.players[playerID]
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
		room.startGame()
		room.dealConnectedPlayers(ws)
	}

	return nil
}

// Creates a room with the given data and adds the player to that room.
func (hub *Hub) createRoomWithPlayer(ws *websocket.Conn, roomID, playerID string, data *json.RawMessage) *HandlerError {
	for {
		room, exists := hub.getRoom(roomID)
		if roomID == "" {
			roomID = randSeq(16)
			continue
		} else if exists {
			room.lock.Lock()
			defer room.lock.Unlock()

			if room.isFull() {
				_, oldPlayer := room.forgottenPlayer()
				if oldPlayer == nil {
					return &HandlerError{
						Msg:   fmt.Sprintf("Room %s already exists and is full. Choose a different name.", roomID),
						Event: eventRoomExists,
					}
				}
			}

			return hub.addPlayerToUnlockedRoom(ws, room, roomID, playerID)
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
		players:             make(map[string]*Player),
		limit:               req.Players,
		table:               make([]PlayerCard, 0),
		acePlayerCollection: make([]Card, 0),
	}

	hub.setRoom(roomID, room)

	room.lock.Lock()
	defer room.lock.Unlock()

	return hub.addPlayerToUnlockedRoom(ws, room, roomID, playerID)
}

// shareMessage from one player to everyone in the room (including the player).
func (hub *Hub) shareMessage(ws *websocket.Conn, roomID, playerID, msg string) {
	if msg == "" {
		return
	}

	room, exists := hub.getRoom(roomID)
	if !exists {
		return
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	for _, p := range room.players {
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: playerID,
			Room:   roomID,
			Event:  eventPlayerMsg,
			Msg:    msg,
		})
	}
}

// playerRequestedNewGame broadcasts the request to all players and starts
// a new game if majority have agreed.
func (hub *Hub) playerRequestedNewGame(ws *websocket.Conn, roomID, playerID string) *HandlerError {
	room, exists := hub.getRoom(roomID)
	if !exists {
		return &HandlerError{
			Msg: fmt.Sprintf("Invalid room specified."),
		}
	}

	room.lock.Lock()
	defer room.lock.Unlock()

	player, exists := room.players[playerID]
	if !exists || player.requestedRestart {
		return &HandlerError{
			Msg: fmt.Sprintf("You're not allowed to perform this action."),
		}
	}

	player.requestedRestart = true
	for _, p := range room.players {
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: playerID,
			Room:   roomID,
			Event:  eventNewGameRequest,
		})
	}

	if room.restartRequests() <= uint8(len(room.players)/2) {
		return nil
	}

	log.Printf("Majority of the players in room %s have requested for a restart.", roomID)
	for id, p := range room.players {
		websocket.JSON.Send(p.conn, &GameMessage{
			Player: id,
			Room:   roomID,
			Event:  eventGameRestart,
		})
	}

	room.startGame()
	room.dealConnectedPlayers(ws)
	return nil
}
