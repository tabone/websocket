package main

import (
	"github.com/tabone/websocket"
	"log"
	"net/http"
)

var m *manager

func main() {
	m = &manager{
		users: make(map[int]*websocket.Socket),
	}
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/", http.FileServer(http.Dir("public/")))

	log.Println("listening on localhost:8080.")
	http.ListenAndServe("localhost:8080", nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")

	// Create a new websocket request
	q := &websocket.Request{
		CheckOrigin: func(r *http.Request) bool {
			// Accept all requests.
			return true
		},
	}

	// Try to upgrade the http request.
	s, err := q.Upgrade(w, r)

	if err != nil {
		log.Println("upgrade failed:", err)
	}

	// If upgrade has been successfull, include the socket with the other online
	// sockets
	m.addSocket(s)
}
