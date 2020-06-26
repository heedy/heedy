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

	buf, err := run.Request(t.Handler, "POST", "/", q, map[string]string{
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
	/*

		// Otherwise, get the API calls for matching events
		var s []struct {
			ID           string `db:"id"`
			Type         string `db:"type"`
			BackendQuery []byte `db:"backend_query"`
			Owner        string `db:"owner"`
			ObjectID     string `db:"objectID"`
		}
		err := eh.db.Select(&s, `SELECT dashboard_elements.id,dashboard_elements.type,dashboard_elements.backend_query,objects.owner,objects.id AS objectID FROM dashboard_events
							JOIN dashboard_elements ON (dashboard_elements.id=dashboard_events.element_id)
							JOIN objects ON (dashboard_elements.object_id=objects.id)
							WHERE dashboard_events.event=? AND dashboard_events.object_id=?;`, e.Event, e.Object)

		if err != nil {
			logrus.Errorf("Failed to get dashboard events for (%s,%s): %v", e.Event, e.Object, err)
			return
		}

		// For each of the dashboard elements, fire the updated event
		for i := range s {
			events.Fire(&events.Event{})
			logrus.Debugf("Dashboard element %s/%s is outdated", s[i].ObjectID, s[i].ID)
		}
	*/
}
