package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/vavvolo/chat/trace"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// In order to use web sockets, we need to upgrade the HTTP connection using the websocket.Upgrader type.
// This is reusable, so we need to create only one
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type room struct {
	// forward is a channel that holds messages that should be forwarded to all clients in the room.
	forward chan []byte
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all the clients that joined the room.
	clients map[*client]bool
	// tracer will receive trace information about activity in the room
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	// This loop will run forever - this is not an issue because
	// this will run in a goroutine in background
	// so it won't block the rest of the application.
	// This loop will watch the 3 channels (join, leave, and forward)
	// and act accordingly based on which channel receives a message.
	// IMPORTANT: the select statement will run only one case at the time,
	// even if there are multiple messages in the join and leave channels,
	// so modifying the clients map won't result in concurrent access errors.
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
			r.tracer.Trace("New client joined.")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left.")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", string(msg))
			// forward message to all clients
			for client := range r.clients {
				client.send <- msg
				r.tracer.Trace(" -- sent to client.")
			}
		}
	}
}

func (r *room) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// A room can now act as a http.Handler
	// because we implement ServeHTTP(http.ResponseWriter, *http.Request)

	// when we get a HTTP request, we upgrade the connection to a web socket
	socket, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	// create a client object
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	// send a message to the room join channel
	// and defer a message to the leave channel
	r.join <- client
	defer func() { r.leave <- client }()

	// run write in a different goroutine
	go client.write()

	// call read on the current goroutine
	// this will block operations, keeping the connection alive
	// until it's time to close it
	client.read()
}
