package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	room struct {
		clients map[*client]bool
		join    chan *client
		leave   chan *client
		forward chan []byte
	}
)

func newRoom() *room {
	return &room{
		clients: make(map[*client]bool),
		join:    make(chan *client),
		leave:   make(chan *client),
		forward: make(chan []byte),
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var (
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  socketBufferSize,
		WriteBufferSize: socketBufferSize,
	}
)

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// Joining
			r.clients[client] = true
		case client := <-r.leave:
			// Leaving
			delete(r.clients, client)
			close(client.receive)
		case msg := <-r.forward:
			// Forwarding
			for client := range r.clients {
				client.receive <- msg
			}
		}
	}
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	client := &client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}

	r.join <- client

	defer func() { r.leave <- client }()

	go client.write()

	client.read()
}
