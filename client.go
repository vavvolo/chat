package main

import (
	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type client struct {
	// socket holds a reference to the web socket that allows us to communicate with the client.
	socket *websocket.Conn
	// send is a buffered channel where received messages are queued to be forwarded to the user's browser via the socket.
	send chan []byte
	// room is the room this client is chatting in.
	room *room
}

// read allows the client to read from the web socket
// and sending any received message to the room forward channel.
func (c *client) read() {
	defer c.socket.Close() // closes the underlying network connection

	for {
		_, p, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		c.room.forward <- p
	}
}

// write continually accepts messages from the send channel
// and writes everything to the web socket.
func (c *client) write() {
	defer c.socket.Close()

	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
