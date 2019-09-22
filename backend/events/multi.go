package events

import (
	"sync"
)

type MultiHandler struct {
	sync.RWMutex
	Routers map[Handler]bool
}

func (em *MultiHandler) Fire(e *Event) {
	em.RLock()
	defer em.RUnlock()
	for r := range em.Routers {
		r.Fire(e)
	}
}

func (em *MultiHandler) AddHandler(er Handler) error {
	em.Lock()
	defer em.Unlock()
	em.Routers[er] = true
	return nil
}

func (em *MultiHandler) RemoveHandler(er Handler) error {
	em.Lock()
	defer em.Unlock()
	delete(em.Routers, er)
	return nil
}

func NewMultiHandler() *MultiHandler {
	return &MultiHandler{
		Routers: make(map[Handler]bool),
	}
}
