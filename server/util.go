package main

import (
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
	labels = [...]string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	suites = [...]string{"d", "c", "h", "s"}
)

func randomlySortedDeck() []Card {
	deck := make([]Card, len(labels)*len(suites))
	i := 0
	for _, s := range suites {
		for _, l := range labels {
			deck[i] = Card{
				Label: l,
				Suite: s,
			}
			i++
		}
	}

	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

func randomDeckChunks(numHands uint8) [][]Card {
	n := int(numHands)
	deck := randomlySortedDeck()
	perHand := len(deck) / n
	extra := len(deck) % n

	start := 0
	hands := make([][]Card, n)
	for i := 0; i < n; i++ {
		size := perHand
		if i < extra {
			size++
		}

		end := start + size
		hands[i] = deck[start:end]
		start += size
	}

	return hands
}
