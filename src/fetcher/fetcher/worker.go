package fetcher

import (
	"fmt"
	"instagram"
	"sync"
	"time"
)

// Worker defines fetcher worker
type Worker struct {
	api     *instagram.Instagram
	timeout time.Duration
}

type workerPool struct {
	sync.RWMutex
	values map[string]*Worker
}

var pool = workerPool{values: make(map[string]*Worker)}

// delay for next processing loop
func (w *Worker) next() {
	time.Sleep(w.timeout)
}

func (w *Worker) start() {
	go w.getActivity()
	go w.directActivity()

	pool.Lock()
	pool.values[w.api.GetUserName()] = w
	pool.Unlock()
}

// GetWorker returns worker with given instagram username
func GetWorker(username string) (*Worker, error) {
	pool.RLock()
	worker, found := pool.values[username]
	pool.RUnlock()
	if !found {
		return nil, fmt.Errorf("Worker for username=%v not found", username)
	}
	return worker, nil
}
