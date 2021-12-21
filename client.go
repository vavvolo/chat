package main

import (
	"time"

	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type client struct {
	// socket holds a reference to the web socket that allows us to communicate with the client.
	socket *websocket.Conn
	// send is a buffered channel where received messages are queued to be forwarded to the user's browser via the socket.
	send chan *message
	// room is the room this client is chatting in.
	room *room
	// userData holds information about the user
	userData map[string]interface{}
}

// read allows the client to read from the web socket
// and sending any received message to the room forward channel.
func (c *client) read() {
	defer c.socket.Close() // closes the underlying network connection

	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err != nil {
			return
		}

		msg.UserID = c.userData[userIDKey].(string)
		msg.FullName = c.userData[fullNameKey].(string)
		msg.AvatarURL, _ = c.room.avatar.GetAvatarURL(c)
		msg.When = time.Now()

		c.room.forward <- msg
	}
}

// write continually accepts messages from the send channel
// and writes everything to the web socket.
func (c *client) write() {
	defer c.socket.Close()

	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			return
		}
	}
}
