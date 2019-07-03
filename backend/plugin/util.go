package plugin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/heedy/heedy/backend/database"
)

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
