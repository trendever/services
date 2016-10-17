package fetcher

import (
	"fmt"
	"instagram"
	"sync"
)

// Worker defines fetcher worker
type Worker struct {
	pool     *instagram.Pool
	username string
}

type workerPool struct {
	sync.RWMutex
	values map[string]*Worker
}

var pool = workerPool{values: make(map[string]*Worker)}

// delay for next processing loop
// not needed now, actually?
// @CHECK
func (w *Worker) next() {
}

func (w *Worker) start() {
	go w.getActivity()
	go w.directActivity()

	pool.Lock()
	pool.values[w.username] = w
	pool.Unlock()
}

func (w *Worker) api() *instagram.Instagram {
	return w.pool.GetFree()
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
