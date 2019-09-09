package plugin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/heedy/heedy/backend/database"
)

func ErrPlugin(err string, args ...interface{}) error {
	s := fmt.Sprintf(err, args...)
	return fmt.Errorf("internal_plugin_error: %s", s)
}

// NoOverlay returns a copy of the database with all overlays removed if it is a PluginDB. This is needed
// for any queries that are run from source implementations
func NoOverlay(db database.DB) database.DB {
	pdb, ok := db.(*PluginDB)
	if !ok {
		return db
	}
	pdb2 := &PluginDB{
		P:         pdb.P,
		Entity:    pdb.Entity,
		Overlay:   -1,
		RequestID: pdb.RequestID,
		client:    pdb.client,
	}
	return pdb2
}

// UnmarshalSourceMeta extracts the meta portion from the source, unmarshalling it into the given object.
// This is because the meta portion of the source is base64 encoded in the X-Heedy-Meta header
// to avoid unnecessary read queries to the database.
func UnmarshalSourceMeta(r *http.Request, obj interface{}) error {
	m := r.Header.Get("X-Heedy-Meta")
	if m == "" {
		return errors.New("server_error: could not find X-Heedy-Meta header")
	}
	b, err := base64.StdEncoding.DecodeString(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)
}

// SourceInfo holds the information sent from heedy as http headers about a source.
// These headers are only present in requests for source API
type SourceInfo struct {
	Type   string
	ID     string
	Meta   map[string]interface{}
	Access database.ScopeArray
}

// GetSourceInfo prepares all source details that come in as part of a source request
func GetSourceInfo(r *http.Request) (*SourceInfo, error) {
	si := SourceInfo{
		Type: r.Header.Get("X-Heedy-Type"),
		ID:   r.Header.Get("X-Heedy-Source"),
	}
	if si.Type == "" || si.ID == "" {
		return nil, ErrPlugin("No type or ID headers were present in source request")
	}
	a, ok := r.Header["X-Heedy-Access"]
	if !ok {
		return nil, ErrPlugin("No access scopes were present in source request")
	}
	si.Access = database.ScopeArray{Scopes: a}

	m := r.Header.Get("X-Heedy-Meta")
	if m == "" {
		return nil, ErrPlugin("No meta in source request")
	}

	b, err := base64.StdEncoding.DecodeString(m)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &(si.Meta))
	return &si, err

}
