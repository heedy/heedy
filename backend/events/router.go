package events

import (
	"errors"
	"sync"
)

var ErrNotSubscribed = errors.New("Not subscribed")

type eventListElement struct {
	e Event
	h Handler
}

type eventList struct {
	list []eventListElement
}

func (el eventList) Fire(e *Event) {
	for i := range el.list {
		e2 := el.list[i].e
		if e2.App != "" && e2.App != "*" && e2.App != e.App {

		} else if e2.Key != "" && e2.Key != "*" && e2.Key != e.Key {

		} else if e2.Object != "" && e2.Object != "*" && e2.Object != e.Object {

		} else if e2.Plugin != nil && *e2.Plugin != "*" && (e.Plugin == nil || *e2.Plugin != *e.Plugin) {

		} else if e2.Type != "" && e2.Type != "*" && e2.Type != e.Type {

		} else if e2.User != "" && e2.User != "*" && e2.User != e.User {

		} else {
			el.list[i].h.Fire(e)
		}
	}
}

func (el *eventList) Subscribe(event Event, h Handler) error {
	el.Unsubscribe(event, h) // Make sure we don't duplicate handlers
	el.list = append(el.list, eventListElement{
		e: event,
		h: h,
	})
	return nil
}

func (el *eventList) Unsubscribe(e Event, h Handler) error {
	for i := range el.list {
		e.Event = el.list[i].e.Event // Make the events match
		if el.list[i].h == h && el.list[i].e == e {
			if len(el.list)-i > 1 {
				el.list[i] = el.list[len(el.list)-1]
			}
			el.list = el.list[:len(el.list)-1]
			return nil
		}
	}
	return ErrNotSubscribed
}

func newEventList() *eventList {
	return &eventList{
		list: make([]eventListElement, 0),
	}
}

// Map: NOT THREADSAFE
type Map map[string]*MultiHandler

func (em Map) Fire(e *Event) {
	m, ok := em[e.Event]
	if ok {
		m.Fire(e)
	}
	m, ok = em["*"]
	if ok {
		m.Fire(e)
	}
}

func (em Map) Subscribe(event string, h Handler) error {
	mh, ok := em[event]
	if !ok {
		mh = NewMultiHandler()
		em[event] = mh
	}
	return mh.AddHandler(h)
}

func (em Map) Unsubscribe(event string, h Handler) error {
	mh, ok := em[event]
	if !ok {
		return ErrNotSubscribed
	}
	return mh.RemoveHandler(h)
}

func NewMap() Map {
	return make(Map)
}

type Router struct {
	sync.RWMutex

	EventMap map[string]*eventList

	NoEvent *eventList
}

func NewRouter() *Router {
	return &Router{
		EventMap: make(map[string]*eventList),
		NoEvent:  newEventList(),
	}
}

func (r *Router) Subscribe(e Event, h Handler) error {
	r.Lock()
	defer r.Unlock()
	if e.Event == "" || e.Event == "*" {
		return r.NoEvent.Subscribe(e, h)
	}
	em, ok := r.EventMap[e.Event]
	if !ok {
		em = newEventList()
		r.EventMap[e.Event] = em
	}
	return em.Subscribe(e, h)
}

func (r *Router) Unsubscribe(e Event, h Handler) error {
	r.Lock()
	defer r.Unlock()
	if e.Event == "" || e.Event == "*" {
		return r.NoEvent.Unsubscribe(e, h)
	}
	em, ok := r.EventMap[e.Event]
	if !ok {
		return ErrNotSubscribed
	}
	return em.Unsubscribe(e, h)
}

func (r *Router) Fire(e *Event) {
	r.RLock()
	defer r.RUnlock()
	em, ok := r.EventMap[e.Event]
	if ok {
		em.Fire(e)
	}
	r.NoEvent.Fire(e)
}
