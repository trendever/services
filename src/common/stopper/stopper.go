package stopper

import "sync"

type Stopper struct {
	once sync.Once
	ch   chan struct{}
}

func NewStopper() *Stopper {
	return &Stopper{ch: make(chan struct{})}
}

func (s *Stopper) Chan() <-chan struct{} {
	return s.ch
}

func (s *Stopper) Stop() {
	s.once.Do(func() { close(s.ch) })
}
