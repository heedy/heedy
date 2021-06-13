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
// for any queries that are run from object implementations
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
		host:      pdb.host,
		client:    pdb.client,
	}
	return pdb2
}

// UnmarshalObjectMeta extracts the meta portion from the object, unmarshalling it into the given object.
// This is because the meta portion of the object is base64 encoded in the X-Heedy-Meta header
// to avoid unnecessary read queries to the database.
func UnmarshalObjectMeta(r *http.Request, obj interface{}) error {
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

// ObjectInfo holds the information sent from heedy as http headers about a object.
// These headers are only present in requests for object API
type ObjectInfo struct {
	Type         string
	ID           string
	Owner        string
	App          string
	ModifiedDate *string
	Meta         map[string]interface{}
	Access       database.ScopeArray
}

// AsObject returns the "As" of the object owner, be it a user or an app
func (o *ObjectInfo) AsObject() string {
	as := o.Owner
	if o.App != "" {
		as += "/" + o.App
	}
	return as
}

// GetObjectInfo prepares all object details that come in as part of a object request
func GetObjectInfo(r *http.Request) (*ObjectInfo, error) {
	si := ObjectInfo{
		Type:  r.Header.Get("X-Heedy-Type"),
		ID:    r.Header.Get("X-Heedy-Object"),
		Owner: r.Header.Get("X-Heedy-Owner"),
	}
	if si.Type == "" || si.ID == "" || si.Owner == "" {
		return nil, ErrPlugin("No type or ID or Owner headers were present in object request")
	}
	ne, ok := r.Header["X-Heedy-Modified-Date"]
	if !ok || len(ne) != 1 {
		return nil, ErrPlugin("No Modified-Date in object request")
	}
	if ne[0] != "null" {
		si.ModifiedDate = &ne[0]
	}
	ne, ok = r.Header["X-Heedy-App"]
	if ok && len(ne) > 0 {
		if ne[0] != "null" {
			si.App = ne[0]
		}
	}

	a, ok := r.Header["X-Heedy-Access"]
	if !ok {
		return nil, ErrPlugin("No access scope was present in object request")
	}
	si.Access = database.ScopeArray{Scope: a}

	m := r.Header.Get("X-Heedy-Meta")
	if m == "" {
		return nil, ErrPlugin("No meta in object request")
	}

	b, err := base64.StdEncoding.DecodeString(m)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &(si.Meta))
	return &si, err

}
