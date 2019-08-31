package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

func TestPlayerGettingDumped(t *testing.T) {
	assert := assert.New(t)
	hands := []string{
		"[{\"label\":\"4\",\"suite\":\"d\"},{\"label\":\"5\",\"suite\":\"d\"},{\"label\":\"7\",\"suite\":\"d\"},{\"label\":\"4\",\"suite\":\"c\"},{\"label\":\"7\",\"suite\":\"c\"},{\"label\":\"K\",\"suite\":\"c\"},{\"label\":\"A\",\"suite\":\"c\"},{\"label\":\"3\",\"suite\":\"h\"},{\"label\":\"6\",\"suite\":\"h\"},{\"label\":\"10\",\"suite\":\"h\"},{\"label\":\"J\",\"suite\":\"h\"},{\"label\":\"K\",\"suite\":\"h\"},{\"label\":\"3\",\"suite\":\"s\"},{\"label\":\"4\",\"suite\":\"s\"},{\"label\":\"6\",\"suite\":\"s\"},{\"label\":\"7\",\"suite\":\"s\"},{\"label\":\"10\",\"suite\":\"s\"},{\"label\":\"J\",\"suite\":\"s\"}]",
		"[{\"label\":\"2\",\"suite\":\"d\"},{\"label\":\"3\",\"suite\":\"d\"},{\"label\":\"8\",\"suite\":\"d\"},{\"label\":\"10\",\"suite\":\"d\"},{\"label\":\"2\",\"suite\":\"c\"},{\"label\":\"6\",\"suite\":\"c\"},{\"label\":\"9\",\"suite\":\"c\"},{\"label\":\"4\",\"suite\":\"h\"},{\"label\":\"5\",\"suite\":\"h\"},{\"label\":\"7\",\"suite\":\"h\"},{\"label\":\"Q\",\"suite\":\"h\"}]",
		"[{\"label\":\"3\",\"suite\":\"c\"},{\"label\":\"5\",\"suite\":\"c\"},{\"label\":\"8\",\"suite\":\"c\"},{\"label\":\"10\",\"suite\":\"c\"},{\"label\":\"J\",\"suite\":\"c\"},{\"label\":\"Q\",\"suite\":\"c\"},{\"label\":\"2\",\"suite\":\"h\"},{\"label\":\"8\",\"suite\":\"h\"},{\"label\":\"9\",\"suite\":\"h\"},{\"label\":\"A\",\"suite\":\"h\"},{\"label\":\"2\",\"suite\":\"s\"}]",
	}

	room, h := setup3PlayerRoom(hands)
	room.players["player1"].dealer = true

	p0 := room.players["player1"]
	firstCard := Card{
		Label: "6",
		Suite: "s",
	}
	assert.Contains(p0.hand, firstCard)
	assert.Len(p0.hand, 18)

	_, err := h.applyPlayerTurn(room, "player1", firstCard)
	assert.Nil(err)
	assert.Equal(room.table[0].Card, firstCard)
	assert.NotContains(p0.hand, firstCard)
	assert.Len(p0.hand, 17)
	assert.EqualValues(room.currentTurn, 1)

	p1 := room.players["player2"]
	secondCard := Card{ // player doesn't have a spade.
		Label: "Q",
		Suite: "h",
	}
	assert.Contains(p1.hand, secondCard)
	assert.Len(p1.hand, 11)

	turnEffect, err := h.applyPlayerTurn(room, "player2", secondCard)
	assert.Nil(err)
	assert.EqualValues(turnEffect, tableFull)
	assert.Len(p0.hand, 19)
	assert.Contains(p0.hand, firstCard) // both cards are dumped to first player.
	assert.Contains(p0.hand, secondCard)
	assert.NotContains(p1.hand, firstCard)
	assert.NotContains(p1.hand, secondCard)
	assert.Len(p1.hand, 10)
	assert.EqualValues(room.currentTurn, 0) // first player becomes dealer.
}

func TestRepetitiveDumps(t *testing.T) {
	assert := assert.New(t)
	hands := []string{
		"[{\"label\":\"9\",\"suite\":\"c\"},{\"label\":\"9\",\"suite\":\"h\"},{\"label\":\"8\",\"suite\":\"s\"},{\"label\":\"4\",\"suite\":\"h\"},{\"label\":\"K\",\"suite\":\"h\"},{\"label\":\"7\",\"suite\":\"s\"},{\"label\":\"2\",\"suite\":\"d\"},{\"label\":\"6\",\"suite\":\"s\"},{\"label\":\"10\",\"suite\":\"c\"},{\"label\":\"K\",\"suite\":\"d\"},{\"label\":\"K\",\"suite\":\"c\"},{\"label\":\"5\",\"suite\":\"h\"},{\"label\":\"7\",\"suite\":\"d\"},{\"label\":\"5\",\"suite\":\"c\"},{\"label\":\"10\",\"suite\":\"d\"},{\"label\":\"4\",\"suite\":\"d\"},{\"label\":\"10\",\"suite\":\"s\"},{\"label\":\"7\",\"suite\":\"c\"}]",
		"[{\"label\":\"3\",\"suite\":\"d\"},{\"label\":\"J\",\"suite\":\"d\"},{\"label\":\"A\",\"suite\":\"d\"},{\"label\":\"7\",\"suite\":\"h\"},{\"label\":\"6\",\"suite\":\"c\"},{\"label\":\"9\",\"suite\":\"s\"},{\"label\":\"K\",\"suite\":\"s\"},{\"label\":\"A\",\"suite\":\"c\"},{\"label\":\"Q\",\"suite\":\"h\"},{\"label\":\"5\",\"suite\":\"s\"},{\"label\":\"J\",\"suite\":\"s\"},{\"label\":\"9\",\"suite\":\"d\"},{\"label\":\"Q\",\"suite\":\"c\"},{\"label\":\"6\",\"suite\":\"h\"},{\"label\":\"3\",\"suite\":\"c\"},{\"label\":\"8\",\"suite\":\"d\"},{\"label\":\"2\",\"suite\":\"c\"}]",
		"[{\"label\":\"2\",\"suite\":\"h\"},{\"label\":\"3\",\"suite\":\"s\"},{\"label\":\"Q\",\"suite\":\"d\"},{\"label\":\"8\",\"suite\":\"c\"},{\"label\":\"A\",\"suite\":\"h\"},{\"label\":\"8\",\"suite\":\"h\"},{\"label\":\"3\",\"suite\":\"h\"},{\"label\":\"4\",\"suite\":\"c\"},{\"label\":\"A\",\"suite\":\"s\"},{\"label\":\"6\",\"suite\":\"d\"},{\"label\":\"2\",\"suite\":\"s\"},{\"label\":\"5\",\"suite\":\"d\"},{\"label\":\"4\",\"suite\":\"s\"},{\"label\":\"Q\",\"suite\":\"s\"},{\"label\":\"10\",\"suite\":\"h\"},{\"label\":\"J\",\"suite\":\"h\"},{\"label\":\"J\",\"suite\":\"c\"}]",
	}

	room, h := setup3PlayerRoom(hands)
	room.currentTurn = 2
	room.players["player3"].dealer = true

	turns := [...]PlayerCard{
		// player3 has ace spade
		PlayerCard{"player3", Card{Label: "Q", Suite: "d"}},
		PlayerCard{"player1", Card{Label: "K", Suite: "d"}},
		PlayerCard{"player2", Card{Label: "A", Suite: "d"}},
		// high rank card in table from player2
		PlayerCard{"player2", Card{Label: "J", Suite: "d"}},
		PlayerCard{"player3", Card{Label: "6", Suite: "d"}},
		PlayerCard{"player1", Card{Label: "10", Suite: "d"}},
		// again from player2
		PlayerCard{"player2", Card{Label: "3", Suite: "d"}},
		PlayerCard{"player3", Card{Label: "5", Suite: "d"}},
		PlayerCard{"player1", Card{Label: "7", Suite: "d"}},
		// now from player1
		PlayerCard{"player1", Card{Label: "4", Suite: "d"}},
		PlayerCard{"player2", Card{Label: "9", Suite: "d"}},
		PlayerCard{"player3", Card{Label: "A", Suite: "h"}},
		// back to player2
		PlayerCard{"player2", Card{Label: "9", Suite: "d"}},
		PlayerCard{"player3", Card{Label: "3", Suite: "h"}},
		// player2 getting smacked
		PlayerCard{"player2", Card{Label: "9", Suite: "d"}},
		PlayerCard{"player3", Card{Label: "A", Suite: "s"}},
	}

	p2 := room.players["player2"]
	assert.Contains(p2.hand, Card{Label: "9", Suite: "d"})
	p3 := room.players["player3"]
	assert.Contains(p3.hand, Card{Label: "A", Suite: "h"})
	assert.Contains(p3.hand, Card{Label: "3", Suite: "h"})
	assert.Contains(p3.hand, Card{Label: "A", Suite: "s"})

	for i, c := range turns {
		effect, err := h.applyPlayerTurn(room, c.ID, c.Card)
		assert.Nil(err)
		if i == 2 || i == 5 || i == 8 || i == 11 || i == 13 || i == 15 {
			assert.EqualValues(effect, tableFull)
		}
		if effect == tableFull {
			room.table = make([]PlayerCard, 0)
		}
	}

	assert.EqualValues(room.currentTurn, 1)
	assert.Contains(p2.hand, Card{Label: "9", Suite: "d"})
	assert.Contains(p2.hand, Card{Label: "A", Suite: "h"})
	assert.Contains(p2.hand, Card{Label: "3", Suite: "h"})
	assert.Contains(p2.hand, Card{Label: "A", Suite: "s"})
	assert.NotContains(p3.hand, Card{Label: "9", Suite: "d"})
	assert.NotContains(p3.hand, Card{Label: "A", Suite: "h"})
	assert.NotContains(p3.hand, Card{Label: "3", Suite: "h"})
	assert.NotContains(p3.hand, Card{Label: "A", Suite: "s"})
}

func TestOnePlayerExiting(t *testing.T) {
	assert := assert.New(t)
	hands := []string{
		"[{\"label\":\"3\",\"suite\":\"c\"},{\"label\":\"2\",\"suite\":\"h\"},{\"label\":\"9\",\"suite\":\"s\"},{\"label\":\"5\",\"suite\":\"h\"},{\"label\":\"2\",\"suite\":\"d\"},{\"label\":\"8\",\"suite\":\"s\"},{\"label\":\"3\",\"suite\":\"h\"},{\"label\":\"8\",\"suite\":\"h\"}]",
		"[{\"label\":\"3\",\"suite\":\"s\"},{\"label\":\"4\",\"suite\":\"s\"},{\"label\":\"6\",\"suite\":\"s\"},{\"label\":\"2\",\"suite\":\"s\"},{\"label\":\"9\",\"suite\":\"h\"},{\"label\":\"7\",\"suite\":\"h\"},{\"label\":\"7\",\"suite\":\"s\"}]",
		"[{\"label\":\"6\",\"suite\":\"h\"}]",
	}

	room, h := setup3PlayerRoom(hands)
	room.currentTurn = 1
	room.players["player2"].dealer = true

	turns := [...]PlayerCard{
		PlayerCard{"player2", Card{Label: "9", Suite: "h"}},
		PlayerCard{"player3", Card{Label: "6", Suite: "h"}},
		PlayerCard{"player1", Card{Label: "8", Suite: "h"}},
	}

	for i, c := range turns {
		effect, err := h.applyPlayerTurn(room, c.ID, c.Card)
		assert.Nil(err)
		if i == 2 {
			assert.EqualValues(effect, tableFull)
			winner := room.endRound()
			assert.Equal(winner, []string{"player3"})
		} else {
			assert.EqualValues(effect, turnApplied)
		}
	}

	p3 := room.players["player3"]
	assert.Empty(p3.hand)
	assert.Empty(room.table)
	assert.True(p3.exited)
}

func TestOnePlayerCaughtWhileExiting(t *testing.T) {
	assert := assert.New(t)
	hands := []string{
		"[{\"label\":\"7\",\"suite\":\"d\"},{\"label\":\"4\",\"suite\":\"d\"},{\"label\":\"8\",\"suite\":\"d\"},{\"label\":\"6\",\"suite\":\"d\"},{\"label\":\"3\",\"suite\":\"s\"},{\"label\":\"6\",\"suite\":\"s\"}]",
		"[{\"label\":\"6\",\"suite\":\"c\"},{\"label\":\"5\",\"suite\":\"d\"},{\"label\":\"4\",\"suite\":\"s\"},{\"label\":\"5\",\"suite\":\"s\"},{\"label\":\"3\",\"suite\":\"d\"},{\"label\":\"5\",\"suite\":\"c\"},{\"label\":\"3\",\"suite\":\"c\"},{\"label\":\"7\",\"suite\":\"c\"},{\"label\":\"2\",\"suite\":\"h\"},{\"label\":\"9\",\"suite\":\"d\"}]",
		"[{\"label\":\"Q\",\"suite\":\"h\"}]",
	}

	room, h := setup3PlayerRoom(hands)
	room.currentTurn = 1
	room.players["player2"].dealer = true

	turns := [...]PlayerCard{
		PlayerCard{"player2", Card{Label: "2", Suite: "h"}},
		PlayerCard{"player3", Card{Label: "Q", Suite: "h"}},
		PlayerCard{"player1", Card{Label: "7", Suite: "d"}},
	}

	p3 := room.players["player3"]
	assert.EqualValues(len(p3.hand), 1)

	for i, c := range turns {
		effect, err := h.applyPlayerTurn(room, c.ID, c.Card)
		assert.Nil(err)
		if i == 2 {
			winnerID := room.endRound()
			assert.Empty(winnerID)
			assert.EqualValues(effect, tableFull)
		} else {
			assert.EqualValues(effect, turnApplied)
		}
	}

	assert.EqualValues(len(p3.hand), 3)
}

func TestOnePlayerExited(t *testing.T) {
	assert := assert.New(t)
	hands := []string{
		"[{\"label\":\"5\",\"suite\":\"s\"},{\"label\":\"5\",\"suite\":\"h\"},{\"label\":\"7\",\"suite\":\"s\"},{\"label\":\"2\",\"suite\":\"h\"},{\"label\":\"9\",\"suite\":\"s\"},{\"label\":\"10\",\"suite\":\"d\"},{\"label\":\"7\",\"suite\":\"d\"},{\"label\":\"4\",\"suite\":\"h\"},{\"label\":\"4\",\"suite\":\"d\"},{\"label\":\"6\",\"suite\":\"s\"},{\"label\":\"3\",\"suite\":\"h\"},{\"label\":\"6\",\"suite\":\"c\"},{\"label\":\"10\",\"suite\":\"c\"},{\"label\":\"8\",\"suite\":\"c\"}]",
		"[{\"label\":\"Q\",\"suite\":\"c\"},{\"label\":\"2\",\"suite\":\"c\"},{\"label\":\"8\",\"suite\":\"h\"},{\"label\":\"K\",\"suite\":\"s\"},{\"label\":\"2\",\"suite\":\"s\"},{\"label\":\"3\",\"suite\":\"c\"},{\"label\":\"9\",\"suite\":\"c\"},{\"label\":\"A\",\"suite\":\"c\"},{\"label\":\"4\",\"suite\":\"s\"},{\"label\":\"10\",\"suite\":\"s\"},{\"label\":\"K\",\"suite\":\"d\"},{\"label\":\"5\",\"suite\":\"d\"},{\"label\":\"8\",\"suite\":\"s\"},{\"label\":\"A\",\"suite\":\"s\"},{\"label\":\"6\",\"suite\":\"h\"},{\"label\":\"A\",\"suite\":\"d\"},{\"label\":\"K\",\"suite\":\"c\"},{\"label\":\"J\",\"suite\":\"c\"},{\"label\":\"Q\",\"suite\":\"s\"},{\"label\":\"9\",\"suite\":\"d\"},{\"label\":\"7\",\"suite\":\"c\"},{\"label\":\"8\",\"suite\":\"d\"},{\"label\":\"5\",\"suite\":\"c\"},{\"label\":\"3\",\"suite\":\"d\"},{\"label\":\"J\",\"suite\":\"s\"},{\"label\":\"3\",\"suite\":\"s\"},{\"label\":\"4\",\"suite\":\"c\"},{\"label\":\"2\",\"suite\":\"d\"},{\"label\":\"10\",\"suite\":\"h\"}]",
		"[]",
	}

	room, h := setup3PlayerRoom(hands)
	room.currentTurn = 1
	room.players["player2"].dealer = true
	room.players["player3"].exited = true

	turns := [...]PlayerCard{
		PlayerCard{"player2", Card{Label: "A", Suite: "s"}},
		PlayerCard{"player1", Card{Label: "9", Suite: "s"}},
		PlayerCard{"player2", Card{Label: "K", Suite: "s"}},
		PlayerCard{"player1", Card{Label: "7", Suite: "s"}},
	}

	for i, c := range turns {
		effect, err := h.applyPlayerTurn(room, c.ID, c.Card)
		if i == 1 {
			assert.EqualValues(effect, tableFull)
			room.endRound()
		}

		assert.Nil(err)
	}

	p2 := room.players["player2"] // player2 is still the dealer
	assert.True(p2.dealer)
	assert.EqualValues(room.currentTurn, 1)
}

func TestGameAceCards(t *testing.T) {
	assert := assert.New(t)
	room, _ := setup3PlayerRoom([]string{"[]", "[]", "[{\"label\":\"5\",\"suite\":\"s\"}]"})

	assert.Nil(room.previousAcePlayer)
	assert.Empty(room.acePlayerCollection)
	p3 := room.players["player3"]
	room.endRound()
	room.startGame()
	assert.Equal(room.acePlayerCollection, []Card{aceSpade})
	assert.EqualValues(2, room.previousAcePlayer.index)
	assert.EqualValues(aceSpade, p3.hand[0])
	assert.EqualValues(2, room.currentTurn)
	assert.True(p3.dealer)

	p1, p2 := room.players["player1"], room.players["player2"]
	for r := 0; r < 5; r++ {
		p1.hand, p2.hand, p3.hand = []Card{}, []Card{}, []Card{Card{"10", "s"}}
		room.endRound()
		room.startGame()
		assert.True(p3.dealer)
	}

	assert.Equal([]Card{
		aceSpade, Card{"A", "c"}, Card{"A", "h"}, Card{"A", "d"}, Card{"K", "s"}, Card{"K", "c"},
	}, room.acePlayerCollection)

	// New player gets ace. Ace cards get reset.
	p1.hand, p2.hand, p3.hand = []Card{Card{"2", "h"}}, []Card{}, []Card{}
	p1.exited = false
	p2.exited = true
	p3.exited = true
	room.startGame()
	assert.Equal([]Card{aceSpade}, room.acePlayerCollection)
	assert.EqualValues(0, room.previousAcePlayer.index)
	assert.EqualValues(aceSpade, p1.hand[0])
	assert.EqualValues(0, room.currentTurn)
	assert.True(p1.dealer)

	// Multiple players have cards. Don't mark player as ace, but
	// leave previous state undisturbed.
	p1.hand, p2.hand, p3.hand = []Card{}, []Card{Card{"2", "h"}}, []Card{Card{"3", "d"}}
	p1.exited = true
	p2.exited = false
	p3.exited = false
	room.startGame()
	assert.Equal([]Card{aceSpade}, room.acePlayerCollection)
	assert.EqualValues(0, room.previousAcePlayer.index)
	assert.EqualValues(aceSpade, p1.hand[0])
	assert.EqualValues(0, room.currentTurn)
	assert.True(p1.dealer)
}

func setup3PlayerRoom(hands []string) (*Room, *Hub) {
	room := &Room{
		players: map[string]*Player{},
		limit:   3,
		table:   make([]PlayerCard, 0),
	}

	for i, h := range hands {
		player := &Player{
			roomID: "test",
			hand:   make([]Card, 0),
			index:  uint8(i),
		}

		json.Unmarshal([]byte(h), &player.hand)
		room.players[fmt.Sprintf("player%d", i+1)] = player
	}

	h := &Hub{
		rooms: map[string]*Room{
			"test": room,
		},
		connRooms: map[*websocket.Conn]string{},
	}

	return room, h
}
