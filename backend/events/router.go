package events

import (
	"errors"
	"fmt"
	"sync"
)

var ErrNotSubscribed = errors.New("Not subscribed")

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

// Permitted queries:
//	user
//	connection
//	source
//	plugin
//	connection-key
// 	plugin-key
// 	user-plugin
//	user-plugin-key

type idKey2 struct {
	id  string
	id2 string
}
type idKey3 struct {
	id  string
	id2 string
	id3 string
}

type Router struct {
	sync.RWMutex

	UserEvents       map[string]Map
	ConnectionEvents map[string]Map
	PluginEvents     map[string]Map
	SourceEvents     map[string]Map

	UserPlugin    map[idKey2]Map
	ConnectionKey map[idKey2]Map
	PluginKey     map[idKey2]Map

	UserPluginKey map[idKey3]Map

	// If SourceType is set, we send it over the full thing again,
	// but this time with the given source type
	SourceType map[string]*Router
}

func NewRouter() *Router {
	// None of the maps are initialized at the beginning, they get set up with subscribe
	return &Router{}
}

func (er *Router) Subscribe(e Event, h Handler) error {
	er.Lock()
	defer er.Unlock()

	if e.Type != "" {
		if er.SourceType == nil {
			er.SourceType = make(map[string]*Router)
		}
		em, ok := er.SourceType[e.Type]
		if !ok {
			em = NewRouter()
			er.SourceType[e.Type] = em
		}
		// Set the type to empty string
		e.Type = ""
		return em.Subscribe(e, h)
	}
	if e.User != "" && e.Plugin != nil && *e.Plugin != "" && e.Key != "" {
		if er.UserPluginKey == nil {
			er.UserPluginKey = make(map[idKey3]Map)
		}
		em, ok := er.UserPluginKey[idKey3{e.User, *e.Plugin, e.Key}]
		if !ok {
			em = NewMap()
			er.UserPluginKey[idKey3{e.User, *e.Plugin, e.Key}] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.Plugin != nil && *e.Plugin != "" && e.Key != "" {
		if er.PluginKey == nil {
			er.PluginKey = make(map[idKey2]Map)
		}
		em, ok := er.PluginKey[idKey2{*e.Plugin, e.Key}]
		if !ok {
			em = NewMap()
			er.PluginKey[idKey2{*e.Plugin, e.Key}] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.Connection != "" && e.Key != "" {
		if er.ConnectionKey == nil {
			er.ConnectionKey = make(map[idKey2]Map)
		}
		em, ok := er.ConnectionKey[idKey2{e.Connection, e.Key}]
		if !ok {
			em = NewMap()
			er.ConnectionKey[idKey2{e.Connection, e.Key}] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.Plugin != nil && *e.Plugin != "" && e.User != "" {
		if er.UserPlugin == nil {
			er.UserPlugin = make(map[idKey2]Map)
		}
		em, ok := er.UserPlugin[idKey2{e.User, *e.Plugin}]
		if !ok {
			em = NewMap()
			er.UserPlugin[idKey2{e.User, *e.Plugin}] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.Source != "" {
		if er.SourceEvents == nil {
			er.SourceEvents = make(map[string]Map)
		}
		em, ok := er.SourceEvents[e.Source]
		if !ok {
			em = NewMap()
			er.SourceEvents[e.Source] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.Plugin != nil && *e.Plugin != "" {
		if er.PluginEvents == nil {
			er.PluginEvents = make(map[string]Map)
		}
		em, ok := er.PluginEvents[*e.Plugin]
		if !ok {
			em = NewMap()
			er.PluginEvents[*e.Plugin] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.Connection != "" {
		if er.ConnectionEvents == nil {
			er.ConnectionEvents = make(map[string]Map)
		}
		em, ok := er.ConnectionEvents[e.Connection]
		if !ok {
			em = NewMap()
			er.ConnectionEvents[e.Connection] = em
		}
		return em.Subscribe(e.Event, h)
	}
	if e.User != "" {
		if er.UserEvents == nil {
			er.UserEvents = make(map[string]Map)
		}
		em, ok := er.UserEvents[e.User]
		if !ok {
			em = NewMap()
			er.UserEvents[e.User] = em
		}
		return em.Subscribe(e.Event, h)
	}

	return fmt.Errorf("Could not subscribe to %s", e.String())
}

func (er *Router) Unsubscribe(e Event, h Handler) error {
	er.Lock()
	defer er.Unlock()
	if e.Type != "" {
		if er.SourceType == nil {
			return ErrNotSubscribed
		}
		em, ok := er.SourceType[e.Type]
		if !ok {
			return ErrNotSubscribed
		}
		// Set the type to empty string
		e.Type = ""
		return em.Unsubscribe(e, h)
	}
	if e.User != "" && e.Plugin != nil && *e.Plugin != "" && e.Key != "" {
		if er.UserPluginKey == nil {
			return ErrNotSubscribed
		}
		em, ok := er.UserPluginKey[idKey3{e.User, *e.Plugin, e.Key}]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.Plugin != nil && *e.Plugin != "" && e.Key != "" {
		if er.PluginKey == nil {
			return ErrNotSubscribed
		}
		em, ok := er.PluginKey[idKey2{*e.Plugin, e.Key}]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.Connection != "" && e.Key != "" {
		if er.ConnectionKey == nil {
			return ErrNotSubscribed
		}
		em, ok := er.ConnectionKey[idKey2{e.Connection, e.Key}]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.Plugin != nil && *e.Plugin != "" && e.User != "" {
		if er.UserPlugin == nil {
			return ErrNotSubscribed
		}
		em, ok := er.UserPlugin[idKey2{e.User, *e.Plugin}]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.Source != "" {
		if er.SourceEvents == nil {
			return ErrNotSubscribed
		}
		em, ok := er.SourceEvents[e.Source]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.Plugin != nil && *e.Plugin != "" {
		if er.PluginEvents == nil {
			return ErrNotSubscribed
		}
		em, ok := er.PluginEvents[*e.Plugin]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.Connection != "" {
		if er.ConnectionEvents == nil {
			return ErrNotSubscribed
		}
		em, ok := er.ConnectionEvents[e.Connection]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}
	if e.User != "" {
		if er.UserEvents == nil {
			return ErrNotSubscribed
		}
		em, ok := er.UserEvents[e.User]
		if !ok {
			return ErrNotSubscribed
		}
		return em.Unsubscribe(e.Event, h)
	}

	return fmt.Errorf("Could not unsubscribe from %s", e.String())
}

func (er *Router) Fire(e *Event) {
	er.RLock()
	defer er.RUnlock()

	// User Subscriptions
	if er.UserEvents != nil {
		h, ok := er.UserEvents[e.User]
		if ok {
			h.Fire(e)
		}
		h, ok = er.UserEvents["*"]
		if ok {
			h.Fire(e)
		}
	}

	// Connection Subscriptions
	if e.Connection == "" {
		return
	}
	if er.ConnectionEvents != nil {
		h, ok := er.ConnectionEvents[e.Connection]
		if ok {
			h.Fire(e)
		}
		h, ok = er.ConnectionEvents["*"]
		if ok {
			h.Fire(e)
		}
	}
	if e.Plugin != nil && *e.Plugin != "" {
		if er.PluginEvents != nil {
			h, ok := er.PluginEvents[*e.Plugin]
			if ok {
				h.Fire(e)
			}
		}
		if er.UserPlugin != nil {
			h, ok := er.UserPlugin[idKey2{e.User, *e.Plugin}]
			if ok {
				h.Fire(e)
			}
		}
	}

	// Source Subscriptions
	if e.Source == "" {
		return
	}
	if er.SourceEvents != nil {
		h, ok := er.SourceEvents[e.Source]
		if ok {
			h.Fire(e)
		}
		h, ok = er.SourceEvents["*"]
		if ok {
			h.Fire(e)
		}
	}

	// This will always be nil in the Type router
	if er.SourceType != nil {
		h, ok := er.SourceType[e.Type]
		if ok {
			h.Fire(e)
		}
	}

	if e.Key == "" {
		return
	}
	if er.ConnectionKey != nil {
		h, ok := er.ConnectionKey[idKey2{e.Connection, e.Key}]
		if ok {
			h.Fire(e)
		}
	}
	if e.Plugin == nil || *e.Plugin == "" {
		return
	}
	if er.PluginKey != nil {
		h, ok := er.PluginKey[idKey2{*e.Plugin, e.Key}]
		if ok {
			h.Fire(e)
		}
	}
	if er.UserPluginKey != nil {
		h, ok := er.UserPluginKey[idKey3{e.User, *e.Plugin, e.Key}]
		if ok {
			h.Fire(e)
		}
	}
}
