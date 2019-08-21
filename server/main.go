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
	eventPlayerJoin   = "PlayerJoin"
	eventPlayerMsg    = "PlayerMsg"
	eventRoomCreate   = "RoomCreate"
	eventRoomExists   = "RoomExists"
	eventRoomMissing  = "RoomMissing"
	eventPlayerExists = "PlayerExists"
	eventPlayerTurn   = "PlayerTurn"
	eventGameOver     = "GameOver"
	eventPlayerWins   = "PlayerWin"

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
