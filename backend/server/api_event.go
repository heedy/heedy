package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
)

func FireEvent(w http.ResponseWriter, r *http.Request) {
	var err error
	c := rest.CTX(r)
	if c.DB.Type() != database.AdminType {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("access_denied: Only plugins may fire events"))
		return
	}
	var e events.Event
	if err = rest.UnmarshalRequest(r, &e); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	if err = database.FillEvent(c.DB.AdminDB(), &e); err == nil {
		events.Fire(&e)

	}
	rest.WriteResult(w, r, err)
}

type WebsocketEventHandler struct {
	Ws           *websocket.Conn
	R            *http.Request
	Heartbeat    time.Duration
	WriteTimeout time.Duration
	Haderror     chan error
	timer        *time.Timer
	sync.Mutex
}

func (eh *WebsocketEventHandler) Close() {
	eh.Lock()
	defer eh.Unlock()
	if eh.timer != nil {
		eh.timer.Stop()
	}
	eh.Heartbeat = 0
}

func (eh *WebsocketEventHandler) ResetHeartbeat() {
	eh.Lock()
	defer eh.Unlock()
	if eh.Heartbeat == 0 {
		return
	}

	if eh.timer == nil {
		eh.timer = time.AfterFunc(eh.Heartbeat, func() {
			c := rest.CTX(eh.R)
			if c.DB.AdminDB().Assets().Config.Verbose {
				c.Log.Debug("<- PING")
			}
			ctx, cancel := context.WithTimeout(eh.R.Context(), eh.WriteTimeout)
			defer cancel()
			err := eh.Ws.Ping(ctx)
			if err != nil {
				select {
				case eh.Haderror <- err:
					// Nothing, let's end
				default:
					// These errors happen once another error already fired, so don't actually warn on them
					rest.CTX(eh.R).Log.Debug("Websocket write secondary error", err)
				}
			} else {
				if c.DB.AdminDB().Assets().Config.Verbose {
					c.Log.Debug("-> PONG")
				}
				eh.ResetHeartbeat()
			}
		})
	} else {
		eh.timer.Stop()
		eh.timer.Reset(eh.Heartbeat)
	}
}

func (eh *WebsocketEventHandler) Fire(e *events.Event) {
	go func() {
		c := rest.CTX(eh.R)

		// It looks like writing an event to a dead connection (i.e. sleeping macbook)
		// succeeds...  We can therefore not count on writing to the connection to give information
		// on whether the connection is alive, so it can't reset the heartbeat handler.
		// https://groups.google.com/g/golang-nuts/c/IDnJDdM5Ek8
		// https://stackoverflow.com/questions/28830549/golang-write-net-conn-without-returning-error-but-the-other-side-of-the-socket-c
		// https://stackoverflow.com/questions/5227520/how-many-times-will-tcp-retransmit
		//
		// It does look like the connection eventually times out even without a heartbeat,
		// I am assuming this is due to the tcp retransmit limit being reached.
		//eh.ResetHeartbeat()

		ctx, cancel := context.WithTimeout(eh.R.Context(), eh.WriteTimeout)
		defer cancel()

		if c.DB.AdminDB().Assets().Config.Verbose {
			c.Log.Debugf("<- %s", e.String())
		}
		err := wsjson.Write(ctx, eh.Ws, e)
		if err != nil {
			select {
			case eh.Haderror <- err:
				// Nothing, let's end
			default:
				// These errors happen once another error already fired, so don't actually warn on them
				c.Log.Debug("Websocket write secondary error", err)
			}

		}
	}()

}

// A wsMessage is a message that is sent to the websocket
type wsMessage struct {
	events.Event
	Cmd string `json:"cmd"`
}

func EventWebsocket(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	cfg := c.DB.AdminDB().Assets().Config
	if c.DB.ID() == "public" && cfg.AllowPublicWebsocket != nil && !*cfg.AllowPublicWebsocket {
		rest.WriteJSONError(w, r, http.StatusForbidden, errors.New("The public is not allowed to access event websockets"))
		return
	}
	ws, err := websocket.Accept(w, r, nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	// The heartbeat value was already validated
	hb, _ := time.ParseDuration(*cfg.WebsocketHeartbeat)
	wt, _ := time.ParseDuration(*cfg.WebsocketWriteTimeout)

	haderror := make(chan error, 1)

	eventHandler := &WebsocketEventHandler{
		Ws:           ws,
		R:            r,
		Haderror:     haderror,
		Heartbeat:    hb,
		WriteTimeout: wt,
	}
	eventHandler.ResetHeartbeat()

	eventRouter := events.NewRouter()

	events.AddHandler(eventRouter)

	c.Log.Debug("Started websocket")

	go func() {
		// This goroutine reads messages, and performs the corresponding subscribe/unsubscribe
		var msg wsMessage
		for {
			var err error
			var b []byte
			if c.DB.AdminDB().Assets().Config.Verbose {
				var mst websocket.MessageType
				mst, b, err = ws.Read(r.Context())
				if err == nil {
					if mst != websocket.MessageText {
						err = errors.New("Websocket message is not text")
					} else {
						c.Log.Debugf("-> %s", string(b))
						err = json.Unmarshal(b, &msg)
					}
				}
			} else {
				err = wsjson.Read(r.Context(), ws, &msg)

			}
			if err != nil {
				select {
				case haderror <- err:
					// Nothing, let's end
				default:
					// These errors happen once another error already fired, so don't actually warn on them
					c.Log.Debug("Websocket read secondary error", err)
				}
				break
			}

			// Since a message was received, reset the heartbeat timer
			eventHandler.ResetHeartbeat()

			// msg now holds the command
			switch msg.Cmd {
			case "subscribe":
				err = database.CanSubscribe(c.DB, &msg.Event)
				if err == nil {
					err = eventRouter.Subscribe(msg.Event, eventHandler)
				}

			case "unsubscribe":
				err = eventRouter.Unsubscribe(msg.Event, eventHandler)
			case "ping":
				// Allow client to send its own heartbeat messages with "ping" command
				eventHandler.Fire(&events.Event{
					Event: "pong",
				})
			default:
				err = fmt.Errorf("Unrecognized command '%s'", msg.Cmd)
			}

			if err != nil {
				select {
				case haderror <- err:
					// Nothing, let's end
				default:
					// These errors happen once another error already fired, so don't actually warn on them
					c.Log.Debug("Websocket invalid message", err)
				}
				break
			}

		}
		if c.DB.AdminDB().Assets().Config.Verbose {
			c.Log.Debug("Closing websocket reader")
		}
	}()

	err = <-haderror
	events.RemoveHandler(eventRouter)
	c.Log.Debug("Closing websocket")

	eventHandler.Close()

	var cerr websocket.CloseError
	if !errors.As(err, &cerr) {
		c.Log.Warn("Websocket abnormal closure:", err)
		ws.Close(websocket.StatusInternalError, err.Error())
		return
	}
	ws.Close(websocket.StatusNormalClosure, "")
}
