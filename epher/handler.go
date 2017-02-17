package epher

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Epher is the main struct to store the global state
type Epher struct {
	roomLock *sync.RWMutex
	Rooms    map[string]*Room
}

// NewEpher creates a new global state
func NewEpher() *Epher {
	return &Epher{
		roomLock: &sync.RWMutex{},
		Rooms:    make(map[string]*Room),
	}
}

func (e *Epher) addConnection(room string, u *User) {
	e.roomLock.Lock()
	if _, ok := e.Rooms[u.Room]; ok { // Room exists
		e.Rooms[room].AddUser(u)
		//log.Println("User added to", room)
	} else {
		r := NewRoom(room)
		r.AddUser(u)
		e.Rooms[room] = r
		//log.Println("Room created", room)
	}
	e.roomLock.Unlock()
}

func (e *Epher) delConnection(room string, u *User) {
	e.roomLock.Lock()
	e.Rooms[room].DelUser(u)
	//log.Println("User removed from", room)
	if e.Rooms[room].UserCount() == 0 { // Nobody left
		delete(e.Rooms, room)
		//log.Println("Room", room, "destroyed")
	}
	e.roomLock.Unlock()
}

// WebsocketHandler is the public websocket interface
func (e *Epher) WebsocketHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ws, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	room := vars["room"]
	u := NewUser(room, ws)
	e.addConnection(room, u)
	defer e.delConnection(room, u)

	// Keep the loop
	u.ReadLoop()
}

//PushHandler sends the HTTP post to the websocket listeners
func (e *Epher) PushHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	room := vars["room"]

	e.roomLock.RLock()
	if _, ok := e.Rooms[room]; ok {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			return
		}
		e.Rooms[room].BroadcastText(b)
	} else {
		log.Println("No listener in", room)
	}

	e.roomLock.RUnlock()
}
