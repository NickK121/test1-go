package model

import (
	"log"
	"sync"
)

var singl sync.Mutex
var messenger *Messenger = nil

type Messenger struct {
	mutex sync.Mutex
}

func MessengerInstance() *Messenger {
	if messenger == nil {
		singl.Lock()
		defer singl.Unlock()
		if messenger == nil {
			messenger = &Messenger{}
		}
	}
	return messenger
}

func (c *Messenger) SendAll(msg []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, u := range users {
		go func(u User) { u.in <- msg }(u)
	}
}

func (c *Messenger) Send(id int, msg []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if id > 0 && id <= len(users) {
		go func() { users[id-1].in <- msg }()
	} else {
		log.Printf("ERROR: user ID is out of range [id=%d]", id)
	}
}
