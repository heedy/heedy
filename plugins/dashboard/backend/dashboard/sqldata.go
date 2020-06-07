package dashboard

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
)

var SQLVersion = 1

const sqlSchema = `

CREATE TABLE dashboard_elements (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,
	object_id VARCHAR(36) NOT NULL,
	idx INT NOT NULL,

	-- The element type specifies the API call to make for backend data
	type VARCHAR NOT NULL,

	-- To save on computation, dashboards are updated on-demand
	outdated BOOL NOT NULL DEFAULT TRUE,

	-- The query to run on the backend to update data
	backend_query BLOB NOT NULL,
	-- Saved output of backend_query
	data BLOB NOT NULL,
	-- Settings for displaying the data on the frontend
	frontend BLOB NOT NULL,

	CONSTRAINT all_valid CHECK (json_valid(backend_query) AND json_valid(frontend) AND json_valid(data)),

	CONSTRAINT object_updater
		FOREIGN KEY(object_id)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE dashboard_events (
	element_id VARCHAR(36) NOT NULL,
	-- The event
	event VARCHAR NOT NULL,
	object_id VARCHAR NOT NULL,

	CONSTRAINT pk PRIMARY KEY (object_id,event),

	CONSTRAINT dashboard_updater
		FOREIGN KEY(element_id)
		REFERENCES dashboard_elements(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	CONSTRAINT underlying_object
		FOREIGN KEY(object_id)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

`

type JSONInterface struct {
	Val interface{}
}

func (s *JSONInterface) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Val)
}

func (s *JSONInterface) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.Val)
}

func (s *JSONInterface) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		return json.Unmarshal(v, &s.Val)
	case string:
		return json.Unmarshal([]byte(v), &s.Val)
	default:
		return fmt.Errorf("Can't unmarshal json object, unsupported type: %T", v)
	}
}
func (s *JSONInterface) Value() (driver.Value, error) {
	return json.Marshal(s.Val)
}

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion >= SQLVersion {
		return errors.New("Dashboard database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

type DashboardElement struct {
	ID       string
	ObjectID string
	Index    int

	Type string

	BackendQuery JSONInterface
	Data         JSONInterface
	Frontend     JSONInterface
}
