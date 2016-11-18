package views

import (
	"payments/models"
	"utils/log"

	"sync"
	"time"
)

const (
	// time to wait processing if it's 2 requests to process at the same time
	recheckDelay = time.Second * 30
	workerNum    = 4
)

// Async checker:
// guarranty that only one pay with given id is checked at this time

type checkerScheduler struct {
	processingNow  map[uint]bool
	processingLock sync.RWMutex
	workerChain    chan *models.Session
	workers        uint
	done           chan bool

	server *paymentServer
}

func createScheduler(server *paymentServer) *checkerScheduler {

	shed := &checkerScheduler{
		processingNow: make(map[uint]bool),
		workerChain:   make(chan *models.Session, 64),
		done:          make(chan bool), // no buffer there so stop() waits until workers are stopped
		server:        server,
	}

	shed.start()

	return shed
}

// process session; run this in a separate routine
func (c *checkerScheduler) process(sess *models.Session) {

	var retries = 2

	for retries > 0 {
		c.processingLock.RLock()
		beingProcessed := c.processingNow[sess.ID]
		c.processingLock.RUnlock()

		if beingProcessed {
			// retry in a some time
			time.Sleep(recheckDelay)
			retries--
			continue
		}

		c.workerChain <- sess
		break
	}
}

// start routines
func (c *checkerScheduler) start() {
	for i := 0; i < workerNum; i++ {
		go c.worker()
	}
}

func (c *checkerScheduler) stop() {
	for i := uint(0); i < c.workers; i++ {
		c.done <- true
	}
	c.workers = 0
}

func (c *checkerScheduler) worker() {
	c.workers++
	for {
		select {
		// stahp first
		case <-c.done:
			return
		case sess := <-c.workerChain:
			c.work(sess)
		}
	}
}

func (c *checkerScheduler) work(sess *models.Session) {

	// set as being processed
	c.processingLock.Lock()
	c.processingNow[sess.ID] = true
	c.processingLock.Unlock()

	// do stuff
	err := c.server.checkStatus(sess)
	if err != nil {
		log.Warn("Error: %v", err)
	}

	// unset being processed
	c.processingLock.Lock()
	delete(c.processingNow, sess.ID)
	c.processingLock.Unlock()
}
