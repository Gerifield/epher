package epher

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var upgrader wsUpgrader = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsUpgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
}

// Roomer help testing the room related here
type Roomer interface {
	AddUser(u *User)
	DelUser(u *User)
	UserCount() int
	BroadcastText(b []byte) error
}

// Epher is the main struct to store the global state
type Epher struct {
	roomLock *sync.RWMutex
	Rooms    map[string]Roomer

	// Metrics related
	allPublishedCounter        prometheus.Counter
	noListenerPublishedCounter prometheus.Counter
	listenerNum                prometheus.Gauge
	roomNum                    prometheus.Gauge
}

// New creates a new global state
func New() *Epher {
	return &Epher{
		roomLock: &sync.RWMutex{},
		Rooms:    make(map[string]Roomer),
	}
}

func (e *Epher) RegisterMetrics() {
	e.allPublishedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "all_publish_ops_total",
		Help: "The total number of published events",
	})

	e.noListenerPublishedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "no_listener_publish_ops_total",
		Help: "The total number of published events where there's no listeners",
	})

	e.listenerNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "listener_num",
		Help: "Actual number of listener connections",
	})

	e.roomNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_num",
		Help: "Actual number of active rooms",
	})
}

func (e *Epher) addConnection(room string, u *User) {
	if e.listenerNum != nil {
		e.listenerNum.Inc()
	}

	e.roomLock.Lock()
	if _, ok := e.Rooms[room]; ok { // Room exists
		e.Rooms[room].AddUser(u)
		//log.Println("User added to", room)
	} else {
		r := NewRoom(room)
		r.AddUser(u)
		e.Rooms[room] = r
		//log.Println("Room created", room)
	}

	if e.roomNum != nil {
		e.roomNum.Set(float64(len(e.Rooms)))
	}
	e.roomLock.Unlock()
}

func (e *Epher) delConnection(room string, u *User) {
	if e.listenerNum != nil {
		e.listenerNum.Dec()
	}

	e.roomLock.Lock()
	e.Rooms[room].DelUser(u)
	//log.Println("User removed from", room)
	if e.Rooms[room].UserCount() == 0 { // Nobody left
		delete(e.Rooms, room)
		//log.Println("Room", room, "destroyed")
	}

	if e.roomNum != nil {
		e.roomNum.Set(float64(len(e.Rooms)))
	}
	e.roomLock.Unlock()
}

// WebsocketHandler is the public websocket interface
// For every new websocket connection we'll keep a new handler open
func (e *Epher) WebsocketHandler(rw http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")

	ws, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Println(err)
		http.Error(rw, "websocket_upgrade_failed", http.StatusInternalServerError)
		return
	}

	u := NewUser(ws)
	e.addConnection(room, u)
	defer e.delConnection(room, u)

	// Keep the loop running until an error or close
	_ = u.ReadLoop()
}

//PushHandler sends the HTTP post to the websocket listeners
func (e *Epher) PushHandler(rw http.ResponseWriter, r *http.Request) {
	room := chi.URLParam(r, "room")

	if e.allPublishedCounter != nil {
		e.allPublishedCounter.Inc()
	}

	e.roomLock.RLock()
	defer e.roomLock.RUnlock()

	if _, ok := e.Rooms[room]; ok {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}

		_ = e.Rooms[room].BroadcastText(b)
	} else {
		log.Println("No listener in", room)
		rw.WriteHeader(http.StatusServiceUnavailable)
		_, _ = rw.Write([]byte("no_room"))

		if e.noListenerPublishedCounter != nil {
			e.noListenerPublishedCounter.Inc()
		}
	}
}
