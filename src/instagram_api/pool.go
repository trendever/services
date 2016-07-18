package instagram_api

import (
	"math/rand"
	"sync"
	"time"
)

type Pool struct {
	storage  chan *Instagram
	stoppers map[*Instagram]chan bool
	settings *PoolSettings
	sync.Mutex
}

type PoolSettings struct {
	ReloginTimeout int // Milliseconds
	TimeoutMin     int
	TimeoutMax     int
}

func NewPool(settings *PoolSettings) *Pool {
	return &Pool{
		storage:  make(chan *Instagram),
		stoppers: make(map[*Instagram]chan bool),
		settings: settings,
	}
}

func (p *Pool) Add(api *Instagram) {
	stop := make(chan bool)

	p.Lock()
	p.stoppers[api] = stop
	p.Unlock()

	go p.poolWorker(api, stop)
}

func (p *Pool) Remove(api *Instagram) {
	p.Lock()
	p.stoppers[api] <- true
	delete(p.stoppers, api)
	p.Unlock()
}

func (p *Pool) RemoveAll() {
	for api := range p.stoppers {
		p.Remove(api)
	}
}

func (p *Pool) poolWorker(api *Instagram, stop <-chan bool) {
	for {
		select {
		case p.storage <- api:
			// each api can be used only after timeout is passed
			p.randomTimeout()

			if !api.isLoggedIn {
				// lower the risks if there are some problems with and account
				time.Sleep(time.Millisecond * time.Duration(p.settings.ReloginTimeout))
			}

		case <-stop:
			return
		}
	}
}

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
