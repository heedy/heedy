package dashboard

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/jmoiron/sqlx/types"
	"github.com/klauspost/compress/zstd"
)

var SQLVersion = 1

const sqlSchema = `

CREATE TABLE dashboard_elements (
	object_id VARCHAR(36) NOT NULL,
	element_id VARCHAR(36) NOT NULL,

	element_index INT NOT NULL,

	-- The element type specifies the API call to make for backend data
	type VARCHAR NOT NULL,

	-- To save on computation, dashboards are updated on-demand
	outdated BOOL NOT NULL DEFAULT TRUE,
	on_demand BOOL NOT NULL DEFAULT TRUE,

	title VARCHAR(100) NOT NULL,

	-- The query to run on the backend to update data
	query BLOB NOT NULL,
	-- Saved output of query, compressed with zstandard (can be large)
	data BLOB DEFAULT NULL,
	-- Settings for displaying the data on the frontend
	settings BLOB NOT NULL,

	PRIMARY KEY (object_id,element_id),

	CONSTRAINT all_valid CHECK (json_valid(query) AND json_valid(settings)),

	CONSTRAINT object_updater
		FOREIGN KEY(object_id)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE dashboard_events (
	object_id VARCHAR(36) NOT NULL,
	element_id VARCHAR(36) NOT NULL,

	-- The event
	event VARCHAR NOT NULL,
	event_object_id VARCHAR NOT NULL,

	PRIMARY KEY (object_id,element_id,event_object_id,event),

	CONSTRAINT underlying_element
		FOREIGN KEY(object_id,element_id)
		REFERENCES dashboard_elements(object_id,element_id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	CONSTRAINT event_object_c
		FOREIGN KEY(event_object_id)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
CREATE INDEX events_idx ON dashboard_events(event_object_id,event);
`

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, h run.BuiltinHelper, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion >= SQLVersion {
		return errors.New("Dashboard database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

var zencoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(2)))
var zdecoder, _ = zstd.NewReader(nil)

type CompressedJSON types.JSONText

func (j CompressedJSON) MarshalJSON() ([]byte, error) {
	return types.JSONText(j).MarshalJSON()
}

// UnmarshalJSON sets *j to a copy of data
func (j *CompressedJSON) UnmarshalJSON(data []byte) error {
	*j = data
	return nil
}

// Value returns j as a value.  This does a validating unmarshal into another
// RawMessage.  If j is invalid json, it returns an error.
func (j CompressedJSON) Value() (driver.Value, error) {
	return zencoder.EncodeAll([]byte(j), make([]byte, 0, len(j))), nil
}

// Scan stores the src in *j.  No validation is done.
func (j *CompressedJSON) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case string:
		source = []byte(t)
	case []byte:
		source = t
	case nil:
		*j = CompressedJSON("null")
		return nil
	default:
		return errors.New("Incompatible type for Compressed JSON")
	}
	b, err := zdecoder.DecodeAll(source, make([]byte, 0, 10*len(source)))
	*j = CompressedJSON(b)
	return err
}

type DashboardEvent struct {
	ObjectID string `json:"object" db:"event_object_id"`
	Event    string `json:"event"`
}

type DashboardElement struct {
	ID       string `json:"id,omitempty" db:"element_id"`
	ObjectID string `json:"object_id,omitempty" db:"object_id"`
	Index    *int   `json:"index,omitempty" db:"element_index"`

	Type     string `json:"type,omitempty" db:"type"`
	OnDemand *bool  `json:"on_demand,omitempty" db:"on_demand"`

	Title    *string         `json:"title,omitempty"`
	Query    *types.JSONText `json:"query,omitempty"`
	Data     *CompressedJSON `json:"data,omitempty"`
	Settings *types.JSONText `json:"settings,omitempty"`

	// Internal variable
	Outdated bool `json:"-"`
}

// ReadDashboard returns the full dashboard data
func ReadDashboard(adb *database.AdminDB, as string, oid string, include_query bool) ([]DashboardElement, error) {
	// Read the full dashboard
	var elements []DashboardElement

	err := adb.Select(&elements, `SELECT * FROM dashboard_elements WHERE object_id=? ORDER BY element_index ASC;`, oid)
	if err != nil {
		return nil, err
	}

	// At this point, some of the dashboard elements might be outdated, and therefore need to be queried.
	// We need to query these elements, and replace the current data with the query results
	queue := make([]chan bool, 0)
	for j := range elements {
		if elements[j].Outdated && *elements[j].OnDemand {
			// If it is on-demand, we actually run the query, and return the results
			c, err := Dashboard.UpdateElement(as, &elements[j])
			if err != nil {
				// This error is weird, since it shouldn't actually ever happen. We simply don't do anything in this case,
				// since we'd need to close a bunch of channels on an error here
			} else {
				queue = append(queue, c)
			}

		}
	}

	// We wait for all the data to update
	for j := range queue {
		<-queue[j]
		close(queue[j])
	}

	if !include_query {
		for i := range elements {
			elements[i].Query = nil
			elements[i].OnDemand = nil
		}
	}

	return elements, nil
}

func WriteDashboard(adb *database.AdminDB, as string, oid string, elements []DashboardElement) error {
	// The write query is an ordered list of inserts/updates to dashboard elements.

	// First perform basic validation
	for _, el := range elements {
		if el.ObjectID != "" && el.ObjectID != oid {
			return database.ErrInvalidQuery
		}
		if el.ID == "" {
			// We're creating a new element
			if el.Type == "" {
				return fmt.Errorf("Can't create dashboard element without a type")
			}
		}

	}

	// Fill in the event template, since all events here will have the same template
	eventTemplate := events.Event{
		Event:  "DASHBOARD_EVENT",
		Object: oid,
	}
	err := database.FillEvent(adb, &eventTemplate)
	if err != nil {
		return err
	}

	// Perform the entire modification of the dashboard as a single transaction
	tx, err := adb.BeginImmediatex()
	if err != nil {
		return err
	}

	// Get the max index of the elements
	var maxIndex int
	err = tx.Get(&maxIndex, `SELECT COALESCE(MAX(element_index),-1) FROM dashboard_elements WHERE object_id=?;`, oid)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Prepare an array of events to fire and dashboard queries to initiate
	requery := make([]*DashboardElement, 0)
	evts := make([]*events.Event, 0)

	for _, el := range elements {
		if el.ID != "" {
			// If there is an ID, check if the element already exists
			var de DashboardElement
			err = adb.Get(&de, `SELECT element_id,element_index,type,on_demand,query,settings,title FROM dashboard_elements WHERE element_id=? AND object_id=?;`, el.ID, oid)
			if err == nil {
				// The element exists
				if el.Type == "" {
					el.Type = de.Type
				}

				willRequery := false

				if el.Query != nil || el.Settings != nil || el.Type != de.Type {
					t, ok := Dashboard.Types[el.Type]
					if !ok {
						tx.Rollback()
						return fmt.Errorf("Unrecognized element type '%s'", el.Type)
					}
					err = t.Validate(&el)
					if err != nil {
						tx.Rollback()
						return err
					}
					de.Outdated = true
					if el.OnDemand == nil && !*de.OnDemand || el.OnDemand != nil && !*el.OnDemand {
						requery = append(requery, &de)
						willRequery = true
					}

				}

				// Now update the de element with all new values
				de.Type = el.Type
				if el.Query != nil {
					de.Query = el.Query
				}
				if el.Settings != nil {
					de.Settings = el.Settings
				}
				if el.OnDemand != nil {
					de.OnDemand = el.OnDemand
				}
				if el.Title != nil {
					de.Title = el.Title
				}
				if el.Index != nil {
					// We are setting the index of a dashboard element, so make sure that the indices of all elements
					// are shifted correctly
					if *el.Index < 0 || *el.Index > maxIndex {
						*el.Index = maxIndex
					}
					if *el.Index != *de.Index {
						// Shift the indices to place the current element in the correct spot
						if *el.Index > *de.Index {
							_, err = tx.Exec(`UPDATE dashboard_elements SET element_index=element_index-1 WHERE object_id=? AND element_index>? AND element_index<=?`, oid, *de.Index, *el.Index)
							if err != nil {
								tx.Rollback()
								return err
							}
						} else {
							_, err = tx.Exec(`UPDATE dashboard_elements SET element_index=element_index+1 WHERE object_id=? AND element_index>=? AND element_index<?`, oid, *el.Index, *de.Index)
							if err != nil {
								tx.Rollback()
								return err
							}
						}
					}
					// Update this index
					de.Index = el.Index
				}

				res, err := tx.Exec(`UPDATE dashboard_elements SET 
							title=?,
							type=?,
							settings=?,
							query=?,
							on_demand=?,
							element_index=?,outdated=?
						WHERE element_id=? AND object_id=?;`,
					de.Title, de.Type, de.Settings, de.Query, de.OnDemand, de.Index, de.Outdated, el.ID, oid)
				err = database.GetExecError(res, err)
				if err != nil {
					tx.Rollback()
					return err
				}

				if !willRequery {
					// If not requerying, add the event right now
					newEvent := eventTemplate
					newEvent.Event = "dashboard_element_update"
					newEvent.Data = map[string]interface{}{
						"element_id":   el.ID,
						"element_type": de.Type,
					}
					evts = append(evts, &newEvent)
				}

				continue
			}
			if err != sql.ErrNoRows {
				tx.Rollback()
				return err
			}
		}
		// No such element exists, so we create a new element
		if el.ID == "" {
			el.ID = uuid.New().String()
		}
		if el.Type == "" {
			tx.Rollback()
			return fmt.Errorf("Can't create dashboard element without a type")
		}

		// Make sure that there is a query and frontend, and that they match the schema
		t, ok := Dashboard.Types[el.Type]
		if !ok {
			tx.Rollback()
			return fmt.Errorf("Unrecognized element type '%s'", el.Type)
		}
		if el.Query == nil {
			tx.Rollback()
			return fmt.Errorf("Element has no query")
		}
		if el.Settings == nil {
			v := types.JSONText("{}")
			el.Settings = &v
		}
		if el.Title == nil {
			v := ""
			el.Title = &v
		}
		err = t.Validate(&el)
		if err != nil {
			tx.Rollback()
			return err
		}
		if el.OnDemand == nil {
			defaultOd := true
			el.OnDemand = &defaultOd
		}
		if !*el.OnDemand {
			requery = append(requery, &el)
		}

		// If there is an index, and we are inserting somewhere inside the array, we need to shift indices
		if el.Index == nil || *el.Index == -1 || *el.Index > maxIndex+1 {
			midx := maxIndex + 1
			el.Index = &midx
		}
		if *el.Index <= maxIndex {
			// Shift existing elements to make room for the new one
			res, err := tx.Exec(`UPDATE dashboard_elements SET element_index=element_index+1 WHERE object_id=? AND element_index>=?`, oid, *el.Index)
			err = database.GetExecError(res, err)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		res, err := tx.Exec(`INSERT INTO dashboard_elements(title,type,settings,query,on_demand,element_index,data,outdated,object_id,element_id) VALUES (?,?,?,?,?,?,NULL,TRUE,?,?);`,
			el.Title, el.Type, el.Settings, el.Query, el.OnDemand, el.Index, oid, el.ID)
		err = database.GetExecError(res, err)
		if err != nil {
			tx.Rollback()
			return err
		}

		newEvent := eventTemplate
		newEvent.Event = "dashboard_element_create"
		newEvent.Data = map[string]interface{}{
			"element_id":   el.ID,
			"element_type": el.Type,
		}
		evts = append(evts, &newEvent)

		// We added an element, so increment the max index
		maxIndex++

	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	for i := range requery {
		e := requery[i]
		// Dispatch requery requests for all of the objects that are being changed which are not ondemand
		q, err := e.Query.MarshalJSON()
		if err == nil {
			go func() {
				Dashboard.Query(as, e.ObjectID, e.ID, e.Type, q)
				newEvent := eventTemplate
				newEvent.Event = "dashboard_element_update"
				newEvent.Data = map[string]interface{}{
					"element_id":   e.ID,
					"element_type": e.Type,
				}
				events.Fire(&newEvent)
			}()
		}

	}

	// Now fire the update events for all ondemand elements
	for _, e := range evts {
		events.Fire(e)
	}
	return nil
}

func ReadDashboardElement(adb *database.AdminDB, as string, oid string, deid string, include_query bool) (*DashboardElement, error) {
	var de DashboardElement
	err := adb.Get(&de, `SELECT * FROM dashboard_elements WHERE element_id=? AND object_id=?;`, deid, oid)

	if err != nil {
		if err == sql.ErrNoRows {
			err = database.ErrNotFound
		}
		return nil, err
	}

	if de.Outdated && *de.OnDemand {
		// If it is on-demand, we actually run the query, and return the results
		c, err := Dashboard.UpdateElement(as, &de)
		if err != nil {
			// This error is weird, since it shouldn't actually ever happen. We simply don't do anything in this case,
			// since we'd need to close a bunch of channels on an error here
		} else {
			<-c
			close(c)
		}

	}

	if !include_query {
		de.Query = nil
		de.OnDemand = nil
	}

	return &de, nil
}

func DeleteDashboardElement(adb *database.AdminDB, oid string, deid string) error {
	evt := &events.Event{
		Event:  "dashboard_element_delete",
		Object: oid,
		Data: map[string]interface{}{
			"element_id": deid,
		},
	}
	err := database.FillEvent(adb, evt)
	if err != nil {
		return err
	}

	tx, err := adb.Beginx()
	if err != nil {
		return err
	}
	// Deleting the element requires shifting all indices down
	_, err = tx.Exec(`UPDATE dashboard_elements SET element_index=element_index-1 WHERE object_id=? AND element_index>(SELECT element_index FROM dashboard_elements WHERE object_id=? AND element_id=?)`, oid, oid, deid)
	if err != nil {
		tx.Rollback()
		return err
	}

	result, err := tx.Exec("DELETE FROM dashboard_elements WHERE element_id=? AND object_id=?", deid, oid)
	err = database.GetExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err == nil {
		events.Fire(evt)
	}
	return err
}
