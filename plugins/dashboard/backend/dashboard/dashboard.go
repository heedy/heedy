package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/jmoiron/sqlx/types"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

type DashboardSettings struct {
	Types map[string]struct {
		API            string                 `mapstructure:"api"`
		QuerySchema    map[string]interface{} `mapstructure:"query_schema"`
		FrontendSchema map[string]interface{} `mapstructure:"frontend_schema"`
	} `mapstructure:"types"`
}

type DashboardType struct {
	URI            string
	Handler        http.Handler
	QuerySchema    *gojsonschema.Schema
	FrontendSchema *gojsonschema.Schema
}

func (dt DashboardType) Validate(de *DashboardElement) error {
	if de.Query != nil && dt.QuerySchema != nil {
		b, _ := de.Query.MarshalJSON() // This never gives an error
		res, err := dt.QuerySchema.Validate(gojsonschema.NewBytesLoader(b))
		if err != nil {
			return err
		}
		if !res.Valid() {
			return errors.New(res.Errors()[0].String())
		}
	}
	if de.Frontend != nil && dt.FrontendSchema != nil {
		b, _ := de.Frontend.MarshalJSON() // This never gives an error
		res, err := dt.FrontendSchema.Validate(gojsonschema.NewBytesLoader(b))
		if err != nil {
			return err
		}
		if !res.Valid() {
			return errors.New(res.Errors()[0].String())
		}
	}
	return nil
}

// The DashboardProcessor is a global object that handles all background tasks that happen with dashboards
type DashboardProcessor struct {
	ADB *database.AdminDB
	h   HandlerGetter

	Types map[string]*DashboardType

	// The actively waiting dashboards are set here
	sync.Mutex
	active map[string][]chan []byte
}

// Dashboard is a global variable that is initialized with NewDashboardProcessor when the plugin is set up
var Dashboard *DashboardProcessor

type HandlerGetter interface {
	GetHandler(uri string) (http.Handler, error)
}

func NewDashboardProcessor(db *database.AdminDB, p *assets.Plugin, h HandlerGetter) (*DashboardProcessor, error) {
	var ds DashboardSettings
	err := mapstructure.Decode(p.Settings, &ds)
	if err != nil {
		return nil, err
	}

	// Now set up the dashboard types
	var dTypes = make(map[string]*DashboardType)
	for t := range ds.Types {
		var dt DashboardType
		if ds.Types[t].QuerySchema != nil {
			dt.QuerySchema, err = gojsonschema.NewSchema(gojsonschema.NewGoLoader(ds.Types[t].QuerySchema))
			if err != nil {
				return nil, err
			}
		}
		if ds.Types[t].FrontendSchema != nil {
			dt.FrontendSchema, err = gojsonschema.NewSchema(gojsonschema.NewGoLoader(ds.Types[t].FrontendSchema))
			if err != nil {
				return nil, err
			}
		}
		dt.URI = ds.Types[t].API

		dTypes[t] = &dt
	}

	return &DashboardProcessor{
		ADB:    db,
		Types:  dTypes,
		active: make(map[string][]chan []byte),
		h:      h,
	}, nil
}

func (dp *DashboardProcessor) RunQ(username, oid string, eid string, etype string, q []byte) ([]byte, error) {
	t, ok := dp.Types[etype]
	if !ok {
		return nil, errors.New("Unrecognized dashboard type")
	}

	if t.Handler == nil {
		// Technically it doesn't have a lock here, but is it really a problem? pointer checks should be atomic, right?
		h, err := dp.h.GetHandler(t.URI)
		if err != nil {
			return nil, err
		}

		t.Handler = h
	}

	logrus.Debugf("Querying dashboard element %s/%s/%s (%s)", username, oid, eid, etype)

	buf, err := run.RequestWithContext(dp.ADB, t.Handler, "POST", "/", q, map[string]string{
		"X-Heedy-As": username,
	})
	if err != nil {
		return nil, err
	}
	data := buf.Bytes()

	// Update the element in the database
	res, err := dp.ADB.Exec(`UPDATE dashboard_elements SET outdated=FALSE,data=? WHERE object_id=? AND element_id=?`, data, oid, eid)
	err = database.GetExecError(res, err)
	return data, err
}

// Query performs a dashboard query as username for element etype, with query q
// Query always returns valid bytes, even if there is an error - in that case it returns json error response
func (dp *DashboardProcessor) Query(username string, oid string, eid string, etype string, q []byte) (data []byte, err error) {
	// Need to have only one query at a time, and return result of that query to all processes that are querying it

	mapkey := etype + "|" + string(q)
	dp.Lock()

	arr, ok := dp.active[mapkey]
	if ok {
		// There is an active query for this element, we set up a channel which will send us the data
		c := make(chan []byte)
		dp.active[mapkey] = append(arr, c)
		dp.Unlock()

		data = <-c
		close(c)
		return
	}

	// It is not active, create an empty array to notify that we're querying for it
	dp.active[mapkey] = make([]chan []byte, 0)
	dp.Unlock()

	// Send the data to all other goroutines waiting for it once we're done here
	defer func() {
		dp.Lock()
		arr = dp.active[mapkey]
		delete(dp.active, mapkey)
		dp.Unlock()

		for i := range arr {
			arr[i] <- data
		}
	}()

	// Now perform the query
	data, err = dp.RunQ(username, oid, eid, etype, q)
	if err != nil {
		// We return valid bytes no matter what

		// TODO: if err is of format err_name: text description
		// the returned error name should be the name
		data, _ = json.Marshal(rest.ErrorResponse{
			ErrorName:        "server_error",
			ErrorDescription: err.Error(),
		})
	}
	return
}

func (dp *DashboardProcessor) UpdateElement(username string, de *DashboardElement) (chan bool, error) {
	c := make(chan bool)
	if de.Query == nil {
		return nil, errors.New("Dashboard query is null")
	}
	q, err := de.Query.MarshalJSON()
	if err != nil {
		return nil, err
	}
	go func() {
		data, _ := dp.Query(username, de.ObjectID, de.ID, de.Type, q)
		jt := types.JSONText(data)
		de.Data = &jt
		c <- false
	}()
	return c, nil
}

// Fire handles events
func (dp *DashboardProcessor) Fire(e *events.Event) {
	// So... Let's check if a dashboard is listening to this event
	if e.Object == "" {
		return // Dashboards only listen to explicit object events
	}

	// Otherwise, get the API calls for matching events
	var s []struct {
		ElementID string                `db:"element_id"`
		Type      string                `db:"type"`
		Query     []byte                `db:"query"`
		Owner     string                `db:"owner"`
		ObjectID  string                `db:"objectID"`
		OnDemand  bool                  `db:"on_demand"`
		App       string                `db:"app"`
		Plugin    *string               `json:"plugin,omitempty" db:"plugin"`
		Key       *string               `json:"key,omitempty" db:"key"`
		Tags      *database.StringArray `json:"tags,omitempty" db:"tags"`
	}

	tx, err := dp.ADB.Beginx()
	if err != nil {
		logrus.Errorf("Failed to get dashboard event (%s,%s): %v", e.Event, e.Object, err)
		return
	}
	err = tx.Select(&s, `SELECT dashboard_elements.element_id,dashboard_elements.type,dashboard_elements.query,dashboard_elements.on_demand,objects.owner,objects.id AS objectID,COALESCE(objects.app,'') AS app,apps.plugin,objects.tags,objects.key FROM dashboard_events
							JOIN dashboard_elements ON (dashboard_elements.element_id=dashboard_events.element_id AND dashboard_elements.object_id=dashboard_events.object_id)
							JOIN objects ON (dashboard_elements.object_id=objects.id)
							LEFT JOIN apps ON (objects.app=apps.id)
							WHERE dashboard_events.event=? AND dashboard_events.event_object_id=?;`, e.Event, e.Object)

	if err != nil {
		logrus.Errorf("Failed to get dashboard events for (%s,%s): %v", e.Event, e.Object, err)
		return
	}

	// Mark them all as outdated
	for ev := range s {
		res, err := tx.Exec(`UPDATE dashboard_elements SET outdated=TRUE WHERE element_id=? AND object_id=?;`, s[ev].ElementID, s[ev].ObjectID)
		err = database.GetExecError(res, err)
		if err != nil {
			logrus.Errorf("Failed to set dashboard element outdated %s/%s/%s, %v", s[ev].Owner, s[ev].ObjectID, s[ev].ElementID, err)
			tx.Rollback()
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		logrus.Errorf("Failed to commit outdated dashboard elements for (%s,%s): %v", e.Event, e.Object, err)
		return
	}

	// For each of the dashboard elements,
	for i := range s {
		sv := s[i]
		if !sv.OnDemand {
			// Dispatch the query, and only once the result is ready do we fire the updated event
			go func() {
				Dashboard.Query(sv.Owner, sv.ObjectID, sv.ElementID, sv.Type, sv.Query)
				events.Fire(&events.Event{
					Event:  "dashboard_element_update",
					User:   sv.Owner,
					App:    sv.App,
					Object: sv.ObjectID,
					Plugin: sv.Plugin,
					Key:    sv.Key,
					Tags:   sv.Tags,
					Type:   "dashboard",
					Data: map[string]interface{}{
						"element_id":   sv.ElementID,
						"element_type": sv.Type,
					},
				})
			}()
		} else {
			// Otherwise, fire the event right away, since it will perform the query when the element is read
			events.Fire(&events.Event{
				Event:  "dashboard_element_update",
				User:   sv.Owner,
				App:    sv.App,
				Object: sv.ObjectID,
				Plugin: sv.Plugin,
				Key:    sv.Key,
				Tags:   sv.Tags,
				Type:   "dashboard",
				Data: map[string]interface{}{
					"element_id":   sv.ElementID,
					"element_type": sv.Type,
				},
			})
		}

	}

}
