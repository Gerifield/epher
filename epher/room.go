package epher

import "sync"

// Room is the global state for each room
type Room struct {
	Name      string
	userMutex *sync.RWMutex
	Users     map[int64]*User
}

// NewRoom creates a new room
func NewRoom(name string) *Room {
	return &Room{
		Name:      name,
		userMutex: &sync.RWMutex{},
		Users:     make(map[int64]*User),
	}
}

//AddUser to room
func (r *Room) AddUser(u *User) {
	r.userMutex.Lock()
	r.Users[u.ID] = u
	r.userMutex.Unlock()
}

//DelUser from room
func (r *Room) DelUser(u *User) {
	r.userMutex.Lock()
	delete(r.Users, u.ID)
	r.userMutex.Unlock()
}

//UserCount returns the user's count
func (r *Room) UserCount() int {
	r.userMutex.RLock()
	defer r.userMutex.RUnlock()
	return len(r.Users)
}

//BroadcastText broadcast a text message in the room
func (r *Room) BroadcastText(b []byte) error {
	r.userMutex.RLock()
	var err error
	for _, u := range r.Users {
		err = u.TextSender.SendText(b)
		if err != nil {
			break
		}
	}
	r.userMutex.RUnlock()
	return err
}
