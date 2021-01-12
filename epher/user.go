package epher

import (
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

// User handler to wrap read/write and other data
type User struct {
	ID int64

	connLock *sync.Mutex
	conn     *websocket.Conn
}

//NewUser creates a new user
func NewUser(ws *websocket.Conn) *User {
	return &User{
		ID:       rand.Int63(),
		connLock: &sync.Mutex{},
		conn:     ws,
	}
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
			if e, ok := err.(*websocket.CloseError); ok {
				if websocket.IsCloseError(err, e.Code) { // Just for fun and testing
					return nil
				}
			}
			log.Println("Read error", err)
			return err // Stop the loop
		}
	}
}
