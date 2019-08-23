package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

const (
	// Event for player requesting to join a room and for server
	// notifying of a player joining some room.
	eventPlayerJoin = "PlayerJoin"
	// Event for player creating a new room.
	eventRoomCreate = "RoomCreate"
	// Room already exists and is full.
	eventRoomExists = "RoomExists"
	// Cannot find room for joining.
	eventRoomMissing = "RoomMissing"
	// Player already exists with that name in the room.
	eventPlayerExists = "PlayerExists"
	// Event for player submitting a card for their turn and for server
	// notifying about a player's turn.
	eventPlayerTurn = "PlayerTurn"
	// One player has lost.
	eventGameOver = "GameOver"
	// Some player has successfully gotten rid of all cards in their hand.
	eventPlayerWins = "PlayerWin"

	minPlayers = 3
	maxPlayers = 6
)

func main() {
	rand.Seed(time.Now().UnixNano())
	pathPtr := flag.String("path", "", "Path to serve directory (required).")
	intPtr := flag.Uint("port", 3000, "Listening port")
	flag.Parse()

	hub := &Hub{
		rooms:     make(map[string]*Room),
		connRooms: make(map[*websocket.Conn]string),
	}

	if *pathPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fs := http.FileServer(http.Dir(*pathPtr))
	http.Handle("/", fs)
	http.Handle("/ws", websocket.Handler(hub.serve))

	log.Printf("Listening on port %d\n", *intPtr)
	http.ListenAndServe(fmt.Sprintf(":%d", *intPtr), nil)
}
