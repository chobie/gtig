package irc

import (
	"sync"
	"time"
	"fmt"
)

type Room struct {
	sync.RWMutex
	Name []byte
	Users map[string]*User

	Tick *time.Ticker
	Last int64
}

func (self *Room) AddUser(user *User) {
	self.Lock()
	self.Users[string(user.Name)] = user
	self.Unlock()
}

func (self *Room) PartUser(user *User) {
	self.Lock()
	delete(self.Users, string(user.Name))
	self.Unlock()
}

func (self *Room) CloseRoom() {
	fmt.Printf("[close room]\n")
	if self.Tick != nil {
		self.Tick.Stop()
	}
}

func NewRoom(name []byte) *Room {
	r := &Room{
		Name: make([]byte, len(name)),
		Users: make(map[string]*User),
	}
	copy(r.Name, name)
	return r
}

func NewRoomWithCallback(name []byte, callback func(*Room)) *Room {
	r := &Room{
		Name: make([]byte, len(name)),
		Users: make(map[string]*User),
		Tick: time.NewTicker(3 * time.Minute),
	}
	copy(r.Name, name)

	go callback(r)
	go func() {
		for now := range r.Tick.C {
			_ = now
			callback(r)
		}
	}()
	return r
}
