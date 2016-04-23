package main

import (
	"fmt"
	"github.com/tabone/websocket"
	"log"
	"time"
)

type manager struct {
	/*
		seq is a sequence which will be used to assign a unique id to each
		socket added to the list of online users.
	*/
	seq int

	/*
		users will contain a reference to all the online sockets.
	*/
	users map[int]*websocket.Socket
}

/*
	addSocket is used to add a socket to the online list of users.
*/
func (m *manager) addSocket(s *websocket.Socket) {
	m.seq++
	log.Println("user", m.seq, "has logged in")
	m.users[m.seq] = s
	m.config(m.seq)

	j := fmt.Sprintf(`{"type":"login","data":{"user": %d, "count":%d}}`, m.seq, len(m.users))
	m.broadcast([]byte(j))

	go m.ping(s)

	// Start listening for new data.
	s.Listen()
}

func (m *manager) ping(s *websocket.Socket) {
	t := time.NewTicker(time.Second * 5)

	for {
		<-t.C
		if err := s.Write(websocket.OpcodePing, nil); err != nil {
			log.Println(err)
			break
		}
	}
	t.Stop()
}

/*
	removeSocket is used to remove a socket from the online list of users using
	its id.
*/
func (m *manager) removeSocket(i int) {
	log.Println("user", i, "has logged out")
	delete(m.users, i)
}

/*
	config is used to configure the socket instance.
*/
func (m *manager) config(i int) {
	s := m.users[i]

	s.ReadHandler = func(o int, p []byte) {
		log.Println("user", i, "sent a message:", string(p))
		j := fmt.Sprintf(`{"type":"message","data":"%s"}`, p)
		m.broadcast([]byte(j))
	}

	s.CloseHandler = func(err error) {
		log.Println("user", i, "disconnected:", err)
		m.removeSocket(i)
		j := fmt.Sprintf(`{"type":"logout","data":{"user": %d, "count":%d}}`, i, len(m.users))
		m.broadcast([]byte(j))
	}

	s.PongHandler = func(p []byte) {
		log.Println("user", i, "pong recieved")
	}
}

/*
	broadcast is used to send a message to all the connected users.
*/
func (m *manager) broadcast(p []byte) {
	for _, s := range m.users {
		s.Write(websocket.OpcodeText, p)
	}
}
