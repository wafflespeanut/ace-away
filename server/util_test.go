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
