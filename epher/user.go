package epher

import (
	"io"
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

// TextBroadcaster helps mocking the SendText call
// This is an alternate solution to the Roomer (in the handler.go) which define how a Room should work.
// Here we just define a single requirement to have the User struct a specific method and tie it to an interface.
// Now in the room_test.go we could simply overwrite how this function should work
type TextSender interface {
	SendText(b []byte) error
}

type wsConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	NextWriter(messageType int) (io.WriteCloser, error)
}

// User handler to wrap read/write and other data
type User struct {
	ID int64

	connLock *sync.Mutex
	conn     wsConn

	TextSender // Embed the interface
}

//NewUser creates a new user
func NewUser(ws wsConn) *User {
	u := &User{
		ID:       rand.Int63(),
		connLock: &sync.Mutex{},
		conn:     ws,
	}

	u.TextSender = u // We should wire in the original method into the interface (by default it will be a nil pointer!)
	return u
}

//SendText message to the user
func (u *User) SendText(b []byte) error {
	u.connLock.Lock()
	defer u.connLock.Unlock()

	w, err := u.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }() // Don't forget to close it

	_, err = w.Write(b)
	return err
}

//ReadLoop start
func (u *User) ReadLoop() error {
	var err error
	for {
		_, _, err = u.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			log.Println("Read error", err)
			return err // Stop the loop
		}
	}
}
