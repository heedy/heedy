package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugin"
	"github.com/heedy/heedy/backend/server"
)

var queryDecoder = schema.NewDecoder()

type StreamInfo struct {
	plugin.SourceInfo
	Schema map[string]interface{}
	Actor  bool
}

var ErrNotActor = errors.New("not_actor: The given stream does not accept actions")

func GetStreamInfo(r *http.Request) (*StreamInfo, error) {
	si, err := plugin.GetSourceInfo(r)
	if err != nil {
		return nil, err
	}
	schemaInterface, ok := si.Meta["schema"]
	if !ok {
		return nil, plugin.ErrPlugin("Stream metadata does not include schema")
	}
	schemaMap, ok := schemaInterface.(map[string]interface{})
	if !ok {
		return nil, plugin.ErrPlugin("Stream schema invalid")
	}
	actorInterface, ok := si.Meta["actor"]
	if !ok {
		return nil, plugin.ErrPlugin("Stream has incomplete metadata")
	}
	actor, ok := actorInterface.(bool)
	if !ok {
		return nil, plugin.ErrPlugin("Stream actor info invalid")
	}
	return &StreamInfo{
		SourceInfo: *si,
		Schema:     schemaMap,
		Actor:      actor,
	}, nil
}

func validateRequest(w http.ResponseWriter, r *http.Request, scope string) (*StreamInfo, bool) {
	si, err := GetStreamInfo(r)
	if err != nil {
		server.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return nil, false
	}
	if !si.SourceInfo.Access.HasScope(scope) {
		server.WriteJSONError(w, r, http.StatusInternalServerError, database.ErrAccessDenied("Insufficient permissions"))
		return nil, false
	}
	return si, true
}

func ReadData(w http.ResponseWriter, r *http.Request, action bool) {
	c := server.CTX(r)
	si, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	if action && !si.Actor {
		server.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var q Query

	err := queryDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		server.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	q.Actions = &action

	di, err := OpenSQLData(c.DB.AdminDB().DB).ReadStreamData(si.SourceInfo.ID, &q)
	if err != nil {
		server.WriteJSONError(w, r, 400, err)
		return
	}
	ai, err := NewJsonArrayReader(di)
	if err != nil {
		server.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	_, err = io.Copy(w, ai)
	if err != nil {
		c.Log.Warnf("Read failed: %s", err.Error())
	}
}

func DeleteData(w http.ResponseWriter, r *http.Request, action bool) {
	c := server.CTX(r)
	si, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	if action && !si.Actor {
		server.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var q Query

	err := queryDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		server.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	q.Actions = &action

	server.WriteResult(w, r, OpenSQLData(c.DB.AdminDB().DB).RemoveStreamData(si.SourceInfo.ID, &q))
}

func WriteData(w http.ResponseWriter, r *http.Request, action bool) {
	c := server.CTX(r)
	si, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	if action && !si.Actor {
		server.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var iq InsertQuery
	err := queryDecoder.Decode(&iq, r.URL.Query())
	if err != nil {
		server.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	iq.Actions = &action

	if action && !si.Actor {
		server.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}

	var datapoints DatapointArray

	err = server.UnmarshalRequest(r, &datapoints)
	if err != nil {
		server.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	dv, err := NewDataValidator(NewDatapointArrayIterator(datapoints), si.Schema, c.DB.ID())
	if err != nil {
		server.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	server.WriteResult(w, r, OpenSQLData(c.DB.AdminDB().DB).WriteStreamData(si.SourceInfo.ID, dv, &iq))
}

func DataLength(w http.ResponseWriter, r *http.Request, action bool) {
	c := server.CTX(r)
	si, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	if action && !si.Actor {
		server.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	l, err := OpenSQLData(c.DB.AdminDB().DB).StreamDataLength(si.SourceInfo.ID, action)
	server.WriteJSON(w, r, l, err)
}

// Act is given just the data portion of a datapoint, and it is inserted at the current timestamp
func Act(w http.ResponseWriter, r *http.Request) {
	c := server.CTX(r)
	si, ok := validateRequest(w, r, "act")
	if !ok {
		return
	}
	if !si.Actor {
		server.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var i interface{}
	err := server.UnmarshalRequest(r, &i)
	if err != nil {
		server.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	dv, err := NewDataValidator(NewDatapointArrayIterator(DatapointArray{NewDatapoint(i)}), si.Schema, c.DB.ID())
	if err != nil {
		server.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	t := "append"
	a := true
	server.WriteResult(w, r, OpenSQLData(c.DB.AdminDB().DB).WriteStreamData(si.SourceInfo.ID, dv, &InsertQuery{
		Type:    &t,
		Actions: &a,
	}))
}

func DataMux() *chi.Mux {
	m := chi.NewMux()

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ReadData(w, r, false)
	})
	m.Delete("/", func(w http.ResponseWriter, r *http.Request) {
		DeleteData(w, r, false)
	})
	m.Post("/", func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, r, false)
	})
	m.Get("/length", func(w http.ResponseWriter, r *http.Request) {
		DataLength(w, r, false)
	})

	return m
}

func ActionMux() *chi.Mux {
	m := chi.NewMux()

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ReadData(w, r, true)
	})
	m.Delete("/", func(w http.ResponseWriter, r *http.Request) {
		DeleteData(w, r, true)
	})
	m.Post("/", func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, r, true)
	})
	m.Get("/length", func(w http.ResponseWriter, r *http.Request) {
		DataLength(w, r, true)
	})

	return m
}

// Handler is the global router for the stream API
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	m.Mount("/data", DataMux())
	m.Mount("/actions", ActionMux())
	m.Post("/act", Act)

	return m
}()
