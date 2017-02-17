package epher

import (
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

// User handler to wrap read/write and other data
type User struct {
	ID   int64
	Room string

	connLock *sync.Mutex
	conn     *websocket.Conn
}

//NewUser creates a new user
func NewUser(room string, ws *websocket.Conn) *User {
	return &User{
		ID:       rand.Int63(),
		Room:     room,
		connLock: &sync.Mutex{},
		conn:     ws,
	}
}

//SendText message to the user
func (u *User) SendText(b []byte) error {
	u.connLock.Lock()
	defer u.connLock.Unlock() // Defer is not so optimal...
	w, err := u.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close() // Don't forget to close it

	_, err = w.Write(b)
	return err
}

//ReadLoop start
func (u *User) ReadLoop() error {
	var err error
	for {
		_, _, err = u.conn.ReadMessage() // TODO: Move this somewhere in the user, but keep this thread "busy"
		if err != nil {
			if e, ok := err.(*websocket.CloseError); ok {
				if websocket.IsCloseError(err, e.Code) { // Just for fun an testing
					return nil
				}
			}
			log.Println("Read error", err)
			return err // Stop the loop
		}
	}
}
