package events

import (
	"encoding/json"

	"github.com/heedy/heedy/backend/database"
)

type Event struct {
	Event      string `json:"event"`
	User       string `json:"user,omitempty"`
	Connection string `json:"connection,omitempty"`
	Source     string `json:"source,omitempty"`
	Plugin     string `json:"plugin,omitempty"`
	Key        string `json:"key,omitempty"`
	Type       string `json:"type,omitempty"`

	Meta database.JSONObject `json:"meta,omitempty"`
}

func (e *Event) String() string {
	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Handler interface {
	Fire(e *Event)
}
