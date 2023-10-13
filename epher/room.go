package epher

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

const broadcastChPattern = "channel:%s"

// Room is the global state for each room
type Room struct {
	Name      string
	userMutex *sync.RWMutex
	Users     map[int64]*User

	redisC *redis.Client
	pubsub *redis.PubSub
}

// NewRoom creates a new room
func NewRoom(name string, redisC *redis.Client) *Room {
	var ps *redis.PubSub
	if redisC != nil {
		ps = redisC.Subscribe(context.Background(), fmt.Sprintf(broadcastChPattern, name))
	}

	r := &Room{
		Name:      name,
		userMutex: &sync.RWMutex{},
		Users:     make(map[int64]*User),

		redisC: redisC,
		pubsub: ps,
	}

	go r.start()

	return r
}

// AddUser to room
func (r *Room) AddUser(u *User) {
	r.userMutex.Lock()
	r.Users[u.ID] = u
	r.userMutex.Unlock()
}

// DelUser from room
func (r *Room) DelUser(u *User) {
	r.userMutex.Lock()
	delete(r.Users, u.ID)
	r.userMutex.Unlock()
}

// UserCount returns the user's count
func (r *Room) UserCount() int {
	r.userMutex.RLock()
	defer r.userMutex.RUnlock()
	return len(r.Users)
}

// BroadcastText broadcast a text message in the room
func (r *Room) BroadcastText(b []byte) error {
	var err error

	// If we have a redis, publish it aswell
	if r.pubsub != nil {
		res := r.redisC.Publish(context.Background(), fmt.Sprintf(broadcastChPattern, r.Name), string(b))
		err = res.Err()
		if err != nil {
			log.Println("error redis publishing message:", err, "broadcasting locally")

			// If we have no redis or it failed, broadcast locally
			return r.broadcastLocal(b)
		}

		return nil
	}

	// If we have no redis, broadcast locally
	return r.broadcastLocal(b)
}

func (r *Room) broadcastLocal(b []byte) error {
	r.userMutex.RLock()
	defer r.userMutex.RUnlock()

	for _, u := range r.Users {
		err := u.TextSender.SendText(b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Room) start() {
	if r.pubsub == nil {
		return
	}

	ch := r.pubsub.Channel()
	var err error

	for msg := range ch {
		err = r.broadcastLocal([]byte(msg.Payload))
		if err != nil {
			log.Println("error local broadcasting message:", err)
		}
	}
}

// Stop the pubsub if available
func (r *Room) Stop() {
	if r.pubsub == nil {
		return
	}

	if err := r.pubsub.Close(); err != nil {
		log.Println("error closing pubsub:", err)
	}
}
