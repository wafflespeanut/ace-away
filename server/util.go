package main

import (
	"fmt"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

var (
	aceSpade = Card{
		Label: "A",
		Suite: "s",
	}

	prettyMap = map[string]string{
		"d": "♦",
		"c": "♣",
		"h": "♥",
		"s": "♠",
	}

	labelRanks = map[string]uint8{
		"2":  2,
		"3":  3,
		"4":  4,
		"5":  5,
		"6":  6,
		"7":  7,
		"8":  8,
		"9":  9,
		"10": 10,
		"J":  11,
		"Q":  12,
		"K":  13,
		"A":  14,
	}

	suites = [...]string{"s", "c", "h", "d"}
)

// getNextAceCard relative to the given card.
// This is tossed to the lost player's hand.
func getNextAceCard(card Card) *Card {
	for i, suite := range suites {
		if i == len(suites)-1 {
			break
		} else if card.Suite == suite {
			return &Card{
				Label: card.Label,
				Suite: suites[i+1],
			}
		}
	}

	rank := labelRanks[card.Label] - 1
	for label, r := range labelRanks {
		if rank == r {
			return &Card{
				Label: label,
				Suite: "s",
			}
		}
	}

	return nil
}

// cardDeck takes a bunch of cards, adds them to the deck and then adds
// the remaining cards to that deck.
func cardDeck(skipCards []Card) []Card {
	deck := make([]Card, len(labelRanks)*len(suites))
	toSkip := make(map[string]struct{})
	for i, c := range skipCards {
		deck[i] = c
		toSkip[fmt.Sprintf("%s%s", c.Label, c.Suite)] = struct{}{}
	}

	i := len(skipCards)
	for _, s := range suites {
		for l := range labelRanks {
			_, exists := toSkip[fmt.Sprintf("%s%s", l, s)]
			if exists {
				continue
			}

			deck[i] = Card{
				Label: l,
				Suite: s,
			}
			i++
		}
	}

	return deck
}

// randomDeckChunks shuffles a deck, distributes the cards for the
// given number of players and returns the collection. It also takes
// a bunch of cards which are added to the first chunk.
func randomDeckChunks(numHands uint8, aceCards []Card) [][]Card {
	n := int(numHands)
	deck := cardDeck(aceCards)
	perHand := len(deck) / n
	extra := len(deck) % n

	// The first chunk is the one which gets the high rank cards.
	offset := perHand
	if extra > 0 {
		offset++
	}

	if len(aceCards) < offset {
		offset = len(aceCards)
	}

	// Shuffle everything other than the high rank cards.
	shuffleLen := len(deck) - offset
	rand.Shuffle(shuffleLen, func(i, j int) {
		deck[i+offset], deck[j+offset] = deck[j+offset], deck[i+offset]
	})

	start := 0
	hands := make([][]Card, n)
	for i := 0; i < n; i++ {
		size := perHand
		if i < extra {
			size++
		}

		end := start + size
		deckSlice := deck[start:end]
		hands[i] = make([]Card, len(deckSlice))
		copy(hands[i], deckSlice) // copy so that we don't affect the i+1'th slice on appending.
		start += size
	}

	return hands
}
