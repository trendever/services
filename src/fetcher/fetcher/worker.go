package fetcher

import (
	"instagram"
	"sync"
	"time"
)

type worker struct {
	api     *instagram.Instagram
	timeout time.Duration
}

type workerPool struct {
	sync.Mutex
	values map[string]*worker
}

var pool = workerPool{values: make(map[string]*worker)}

// delay for next processing loop
func (w *worker) next() {
	time.Sleep(w.timeout)
}
