package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAceCards(t *testing.T) {
	assert := assert.New(t)
	cards := []Card{
		Card{"A", "s"},
		Card{"10", "h"},
		Card{"Q", "c"},
		Card{"5", "d"},

		Card{"A", "c"},
		Card{"10", "d"},
		Card{"Q", "h"},
		Card{"4", "s"},
	}

	total := len(cards) / 2
	for i := range cards {
		if i == total {
			break
		}

		c := getNextAceCard(cards[i])
		assert.Equal(cards[total+i].Label, c.Label)
		assert.Equal(cards[total+i].Suite, c.Suite)
	}
}

func TestRandomDeck(t *testing.T) {
	assert := assert.New(t)
	testCases := [][]Card{
		[]Card{Card{"A", "s"}},
		[]Card{Card{"A", "s"}, Card{"A", "c"}, Card{"A", "h"}, Card{"A", "d"}},
		[]Card{
			Card{"A", "s"}, Card{"A", "c"}, Card{"A", "h"}, Card{"A", "d"},
			Card{"K", "s"}, Card{"K", "c"}, Card{"K", "h"}, Card{"K", "d"},
			Card{"J", "s"}, Card{"J", "c"}, Card{"J", "h"}, Card{"J", "d"},
		},
	}

	for idx, cards := range testCases {
		prevChunks := make([][]Card, 0)

		for r := 0; r < 10; r++ { // multiple runs for randomness
			chunks := randomDeckChunks(6, cards)
			if len(prevChunks) > 0 {
				for i := range chunks {
					if idx == 2 && i == 0 {
						// In the final test case, the first chunk will always be the
						// same, as the number of losses is more than the number of cards
						// held in hand.
						assert.Equal(prevChunks[i], chunks[i])
					} else {
						assert.NotEqual(prevChunks[i], chunks[i])
					}
				}
			}

			assert.Len(chunks[0], 9)
			for i, c := range cards {
				if i == len(chunks[0]) {
					break
				}

				// Check the order of cards in first chunk
				assert.Equal(c.Label, chunks[0][i].Label)
				assert.Equal(c.Suite, chunks[0][i].Suite)
			}

			prevChunks = chunks
		}
	}
}
