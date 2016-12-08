package instagram

import (
	"math/rand"
	"sync"
	"time"
)

// Pool allows to use multiple accounts easily.
// Instagram fastly bans when it comes to repititive and common requests
// Create a pool, add joined APIs to it
// Every time you use GetFree() you will get a Client instance
// Every client is returned delayed at some random time between TimeoutMin and TimeoutMax
// If there is a problem with client (it's banned for instance), it is delayed with ReloginTimeout

// Pool defines thread-safe account pool
type Pool struct {
	storage  chan *Instagram
	settings *PoolSettings

	stoppers map[*Instagram]chan bool
	mutex    sync.Mutex // mutex only for map access
}

// PoolSettings defines settings for pool to be created
// Timeouts-only for now
type PoolSettings struct {
	ReloginTimeout int // Milliseconds
	TimeoutMin     int
	TimeoutMax     int
}

// NewPool creates a new pool
func NewPool(settings *PoolSettings) *Pool {
	return &Pool{
		storage:  make(chan *Instagram),
		stoppers: make(map[*Instagram]chan bool),
		settings: settings,
	}
}

// Add connection to the pool
func (p *Pool) Add(api *Instagram) {
	stop := make(chan bool)

	p.mutex.Lock()
	p.stoppers[api] = stop
	p.mutex.Unlock()

	go p.poolWorker(api, stop)
}

// Remove only one connection from the pool
func (p *Pool) Remove(api *Instagram) {
	p.mutex.Lock()
	p.stoppers[api] <- true
	delete(p.stoppers, api)
	p.mutex.Unlock()
}

// RemoveAll deletes all connections from the pool
func (p *Pool) RemoveAll() {
	for api := range p.stoppers {
		p.Remove(api)
	}
}

// main loop started for each connection
func (p *Pool) poolWorker(api *Instagram, stop <-chan bool) {
	for {
		select {
		case p.storage <- api:
			// each api can be used only after timeout is passed
			p.randomTimeout()

			if !api.LoggedIn {
				// lower the risks if there are some problems with and account
				time.Sleep(time.Millisecond * time.Duration(p.settings.ReloginTimeout))
			}

		case <-stop:
			return
		}
	}
}

// GetFree returns first freed Connection
// First available client is waited
func (p *Pool) GetFree() *Instagram {
	api := <-p.storage
	return api
}

func (p *Pool) randomTimeout() {

	min := p.settings.TimeoutMin
	max := p.settings.TimeoutMax

	// random timeout
	rndTimeout := min + rand.Intn(max-min)
	time.Sleep(time.Duration(rndTimeout) * time.Millisecond)
}
