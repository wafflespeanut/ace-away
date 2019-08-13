package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

const (
	eventRoomCreate = "RoomCreate"
	eventUserJoin   = "UserJoin"
	eventUserMsg    = "UserMsg"
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

	http.ListenAndServe(fmt.Sprintf(":%d", *intPtr), nil)
}
