package config

import (
	"errors"
	"time"
)

// Websocket pertains to all config options of a websocket
type Websocket struct {
	MessageLimitBytes int64 `json:"message_limit_bytes"`

	// The time to wait on a socket write
	WriteWait time.Duration `json:"write_wait"`

	// Websockets ping each other to keep the connection alive
	// This sets the number od seconds between pings
	PongWait   time.Duration `json:"pong_wait"`
	PingPeriod time.Duration `json:"ping_period"`

	// The websocket read/write buffer for socket upgrader
	ReadBufferSize  int `json:"read_buffer"`
	WriteBufferSize int `json:"write_buffer"`

	// The number of messages to buffer
	MessageBuffer int64 `json:"message_buffer"`
}

// Validate ensires all websocket options are OK
func (w *Websocket) Validate() error {
	if w.MessageLimitBytes < 100 {
		return errors.New("The limit of a websocket message has to be at least 100 bytes.")
	}

	if w.WriteWait < 1 {
		return errors.New("The websocket write wait time must be at least 1 second")
	}

	if w.PongWait < 1 {
		return errors.New("The pong wait time for websocket must be at least 1s.")
	}

	if w.PingPeriod < 1 {
		return errors.New("Websocket ping period must be at least 1 second")
	}

	if w.MessageBuffer < 1 {
		return errors.New("The websocket message buffer must have at least one message")
	}

	if w.WriteBufferSize < 10 {
		return errors.New("Websocket write buffer must be at least 10 bytes")
	}

	if w.ReadBufferSize < 10 {
		return errors.New("The websocket read buffer must be at least 10 bytes")
	}
	return nil
}
