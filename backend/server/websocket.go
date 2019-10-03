package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/events"
)

type WebsocketEventHandler struct {
	Ws       *websocket.Conn
	R        *http.Request
	haderror chan error
}

func (eh *WebsocketEventHandler) Fire(e *events.Event) {
	go func() {
		ctx, cancel := context.WithTimeout(eh.R.Context(), time.Second*10)
		defer cancel()
		c := rest.CTX(eh.R)
		if c.DB.AdminDB().Assets().Config.Verbose {
			c.Log.Debugf("<- %s", e.String())
		}
		err := wsjson.Write(ctx, eh.Ws, e)
		if err != nil {
			select {
			case eh.haderror <- err:
				// Nothing, let's end
			default:
				// These errors happen once another error already fired, so don't actually warn on them
				rest.CTX(eh.R).Log.Debug("Websocket write secondary error", err)
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

	haderror := make(chan error, 1)

	eventHandler := &WebsocketEventHandler{
		Ws: ws,
		R:  r,
	}

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
				_, b, err = ws.Read(r.Context())
				if err == nil {
					c.Log.Debugf("-> %s", string(b))
					err = json.Unmarshal(b, &msg)
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

			// msg now holds the command
			switch msg.Cmd {
			case "subscribe":
				err = events.CanSubscribe(c.DB, &msg.Event)
				if err == nil {
					err = eventRouter.Subscribe(msg.Event, eventHandler)
				}

			case "unsubscribe":
				err = eventRouter.Unsubscribe(msg.Event, eventHandler)
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

	var cerr websocket.CloseError
	if !errors.As(err, &cerr) {
		c.Log.Warn("Websocket abnormal closure:", err)
		ws.Close(websocket.StatusInternalError, err.Error())
		return
	}
	ws.Close(websocket.StatusNormalClosure, "")
}
