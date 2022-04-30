package timeseries

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/klauspost/compress/gzip"
	"github.com/mailru/easyjson"

	"github.com/heedy/heedy/api/golang/plugin"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/database/dbutil"
	"github.com/heedy/heedy/backend/events"
)

var queryDecoder = schema.NewDecoder()

type TimeseriesInfo struct {
	plugin.ObjectInfo
	Schema map[string]interface{}
	Actor  bool
}

var ErrNotActor = errors.New("not_actor: The given timeseries does not accept actions")

func GetTimeseriesInfo(r *http.Request) (*TimeseriesInfo, error) {
	si, err := plugin.GetObjectInfo(r)
	if err != nil {
		return nil, err
	}
	schemaInterface, ok := si.Meta["schema"]
	if !ok {
		return nil, plugin.ErrPlugin("Timeseries metadata does not include schema")
	}
	schemaMap, ok := schemaInterface.(map[string]interface{})
	if !ok {
		return nil, plugin.ErrPlugin("Timeseries schema invalid")
	}
	/*
		actorInterface, ok := si.Meta["actor"]
		if !ok {
			return nil, plugin.ErrPlugin("Timeseries has incomplete metadata")
		}

		actor, ok := actorInterface.(bool)
		if !ok {
			return nil, plugin.ErrPlugin("Timeseries actor info invalid")
		}
	*/
	return &TimeseriesInfo{
		ObjectInfo: *si,
		Schema:     schemaMap,
		Actor:      false,
	}, nil
}

func validateRequest(w http.ResponseWriter, r *http.Request, scope string) (*TimeseriesInfo, bool) {
	si, err := GetTimeseriesInfo(r)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return nil, false
	}
	if !si.ObjectInfo.Access.HasScope(scope) {
		rest.WriteJSONError(w, r, http.StatusForbidden, database.ErrAccessDenied("Insufficient permissions"))
		return nil, false
	}
	return si, true
}

func decodeQuery(r *http.Request) (Query, error) {
	// schema can't handle interface{} values for t1 and t2 and t
	var q struct {
		Query
		T1v *string `schema:"t1"`
		T2v *string `schema:"t2"`
		Tv  *string `schema:"t"`
	}

	err := queryDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		return Query{}, err
	}
	if q.Tv != nil {
		q.Query.T = *q.Tv
	}
	if q.T1v != nil {
		q.Query.T1 = *q.T1v
	}
	if q.T2v != nil {
		q.Query.T2 = *q.T2v
	}
	return q.Query, nil
}

func ReadData(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}

	q, err := decodeQuery(r)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	q.Actions = &action
	if q.Timeseries != "" {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("timeseries arg is set automatically when querying objects"))
		return
	}
	q.Timeseries = si.ObjectInfo.ID

	di, err := TSDB.Query(&q)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	defer di.Close()
	ai, err := NewJsonArrayReader(di, 2048)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	defer ai.Close()

	rest.APIHeaders(w)

	if TSDB.CompressQueryResponse {
		err = rest.WriteCompressAsync(w, r, ai, http.StatusOK)
	} else {
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, ai)
	}

	if err != nil {
		c.Log.Warnf("Timeseries read failed: %s", err.Error())
	}

}

func DeleteData(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	q, err := decodeQuery(r)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	q.Actions = &action
	if q.Timeseries != "" {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("timeseries arg is set automatically when querying objects"))
		return
	}
	q.Timeseries = si.ObjectInfo.ID

	err = TSDB.Delete(&q)
	if err == nil {
		c.Events.Fire(&events.Event{
			Event:  "timeseries_data_delete",
			Object: si.ObjectInfo.ID,
			Data:   q,
		})
	}
	rest.WriteResult(w, r, err)
}

func shouldUpdateModifed(d *string) bool {
	if d == nil {
		return true
	}
	t, err := time.Parse("2006-01-02", *d)
	if err != nil {
		return true
	}
	cy, cm, cd := time.Now().UTC().Date()
	dy, dm, dd := t.Date()
	return cd > dd || cm > dm || cy > dy
}

type TimeseriesWriteEvent struct {
	T1    float64    `json:"t1"`
	T2    float64    `json:"t2"`
	Count int64      `json:"count"`
	DP    *Datapoint `json:"dp,omitempty"`
}

//UnmarshalEasyRequestNoLimit unmarshals the input data to the given interface without limiting request size
// This should be replaced at some point probably...
func UnmarshalEasyRequestNoLimit(request *http.Request, unmarshalTo easyjson.Unmarshaler) error {
	defer request.Body.Close()

	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}

	return easyjson.Unmarshal(data, unmarshalTo)
}

func WriteData(w http.ResponseWriter, r *http.Request, action bool) {
	c := rest.CTX(r)
	scope := "write"
	if action {
		scope = "act"
	}
	si, ok := validateRequest(w, r, scope)
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var iq InsertQuery
	err := queryDecoder.Decode(&iq, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	iq.Actions = &action

	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}

	var datapoints DatapointArray

	err = UnmarshalEasyRequestNoLimit(r, &datapoints)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	actor := ""
	if action {
		actor = c.DB.ID()
		apnd := "append"
		iq.Method = &apnd
	}

	// Need to set actor for all datapoints
	for i := range datapoints {
		if datapoints[i] == nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: null datapoint"))
			return
		}
		datapoints[i].Actor = actor
	}

	if len(si.Schema) > 0 && (iq.Validate == nil || iq.Validate != nil && *iq.Validate || action) {
		// JSON schema validation can take a long time, so do it before we start insert so that it doesn't block the database
		dv, err := NewDataValidator(NewDatapointArrayIterator(datapoints), si.Schema, actor)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
		var dp *Datapoint
		for dp, err = dv.Next(); err == nil && dp != nil; dp, err = dv.Next() {
		}
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}
	}

	ii := NewInfoIterator(NewDatapointArrayIterator(datapoints))
	err = TSDB.Insert(si.ObjectInfo.ID, ii, &iq)
	if err == nil && ii.Count > 0 {
		if shouldUpdateModifed(si.ModifiedDate) {
			ne := dbutil.Date(time.Now().UTC())
			// The timeseries is now non-empty, so label it as such
			err = c.DB.AdminDB().UpdateObject(&database.Object{
				Details: database.Details{
					ID: si.ID,
				},
				ModifiedDate: &ne,
			})
		}
		evt := "timeseries_data_write"
		if action {
			evt = "timeseries_actions_write"
		}
		c.Events.Fire(&events.Event{
			Event:  evt,
			Object: si.ObjectInfo.ID,
			Data: &TimeseriesWriteEvent{
				T1:    ii.Tstart,
				T2:    ii.Tend,
				Count: ii.Count,
				DP:    ii.LastPoint,
			},
		})
	}

	rest.WriteResult(w, r, err)
}

func DataLength(w http.ResponseWriter, r *http.Request, action bool) {
	si, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	if action && !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	l, err := TSDB.Length(si.ObjectInfo.ID, action)
	rest.WriteJSON(w, r, l, err)
}

// Act is given just the data portion of a datapoint, and it is inserted at the current timestamp
func Act(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	si, ok := validateRequest(w, r, "act")
	if !ok {
		return
	}
	if !si.Actor {
		rest.WriteJSONError(w, r, http.StatusBadRequest, ErrNotActor)
		return
	}
	var i interface{}
	err := rest.UnmarshalRequest(r, &i)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	dv, err := NewDataValidator(NewDatapointArrayIterator(DatapointArray{NewDatapoint(i)}), si.Schema, c.DB.ID())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	t := "append"
	a := true

	ii := NewInfoIterator(dv)

	err = TSDB.Insert(si.ObjectInfo.ID, ii, &InsertQuery{
		Method:  &t,
		Actions: &a,
	})

	if err == nil && ii.Count > 0 {
		if shouldUpdateModifed(si.ModifiedDate) {
			ne := dbutil.Date(time.Now().UTC())
			// The timeseries is now non-empty, so label it as such
			err = c.DB.AdminDB().UpdateObject(&database.Object{
				Details: database.Details{
					ID: si.ID,
				},
				ModifiedDate: &ne,
			})
		}
		c.Events.Fire(&events.Event{
			Event:  "timeseries_actions_write",
			Object: si.ObjectInfo.ID,
			Data: &TimeseriesWriteEvent{
				T1:    ii.Tstart,
				T2:    ii.Tend,
				Count: ii.Count,
				DP:    ii.LastPoint,
			},
		})
	}

	rest.WriteResult(w, r, err)
}

func GenerateDataset(rw http.ResponseWriter, r *http.Request) {
	// Generate a dataset
	c := rest.CTX(r)
	var d map[string]*Dataset
	err := rest.UnmarshalRequest(r, &d)
	if err != nil {
		rest.WriteJSONError(rw, r, http.StatusBadRequest, err)
		return
	}
	if _, ok := d["error"]; ok {
		rest.WriteJSONError(rw, r, http.StatusBadRequest, errors.New("The key 'error' is disallowed in dataset query."))
		return
	}
	if _, ok := d["error_description"]; ok {
		rest.WriteJSONError(rw, r, http.StatusBadRequest, errors.New("The key 'error_description' is disallowed in datset query."))
		return
	}

	readers := make(map[string]*JsonArrayReader)
	for k, v := range d {
		if v == nil {
			rest.WriteJSONError(rw, r, http.StatusBadRequest, fmt.Errorf("Invalid query at '%s'", k))
			return
		}
		di, err := v.Get(c.DB)
		if err != nil {
			rest.WriteJSONError(rw, r, http.StatusBadRequest, fmt.Errorf("Invalid query at '%s': %w", k, err))
			return
		}
		var pi DatapointIterator
		/*
			if len(d[i].Dataset) > 0 || d[i].Transform != nil {
				pi = NewChanIterator(&TransformIterator{dpi: di, it: di})
			} else {
				pi = &TransformIterator{dpi: di, it: di}
			}
		*/
		pi = &TransformIterator{dpi: di, it: di}
		defer pi.Close()

		ai, err := NewJsonArrayReader(pi, 2048)
		if err != nil {
			rest.WriteJSONError(rw, r, http.StatusInternalServerError, err)
			return
		}
		readers[k] = ai
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	var w io.Writer
	w = rw

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && !assets.Get().Config.Verbose && TSDB.CompressQueryResponse {
		// If gzip is supported, compress the output
		rw.Header().Set("Content-Encoding", "gzip")
		rw.WriteHeader(http.StatusOK)
		gzw := gzip.NewWriter(rw)
		defer gzw.Close()
		aw := rest.NewAsyncWriter(gzw)
		defer aw.Close()
		w = aw

	} else {
		rw.WriteHeader(http.StatusOK)
	}

	_, err = w.Write([]byte(`{`))
	if err != nil {
		c.Log.Warnf("Dataset: %s", err.Error())
		return
	}
	hasPrev := false
	for k, v := range readers {
		if hasPrev {
			_, err = w.Write([]byte{','})
			if err != nil {
				c.Log.Warnf("Dataset: %s", err.Error())
				return
			}
		}
		hasPrev = true
		ks, err := json.Marshal(k)
		if err != nil {
			c.Log.Warnf("Dataset key '%s': %s", k, err.Error())
			return
		}
		_, err = w.Write(ks)
		if err != nil {
			c.Log.Warnf("Dataset key '%s': %s", k, err.Error())
			return
		}
		_, err = w.Write([]byte(`:`))
		if err != nil {
			c.Log.Warnf("Dataset key '%s': %s", k, err.Error())
			return
		}
		_, err = io.Copy(w, v)

		if err != nil {
			c.Log.Warnf("Dataset key '%s': %s", k, err.Error())
			return
		}
	}
	_, err = w.Write([]byte(`}`))
	if err != nil {
		c.Log.Warnf("Dataset: %s", err.Error())
		return
	}
}

/*

func getEvents(d map[string]*Dataset) []dashboard.DashboardEvent {
	if len(d) == 0 {
		return nil
	}
	m := make(map[string]int)
	for _, v := range d {
		m2 := v.GetTimeseries()
		for k := range m2 {
			// don't care about numbers
			m[k] = 1
		}
	}

	arr := make([]dashboard.DashboardEvent, 0, len(m)*2)
	for k := range m {
		arr = append(arr, dashboard.DashboardEvent{
			ObjectID: k,
			Event:    "timeseries_data_write",
		}, dashboard.DashboardEvent{
			ObjectID: k,
			Event:    "timeseries_data_delete",
		})
	}

	return arr
}

type dashboardQueryResult struct {
	Data interface{} `json:"data"`
}

func GenerateDashboardDataset(w http.ResponseWriter, r *http.Request) {
	// Generate a dataset
	c := rest.CTX(r)
	var d map[string]*Dataset
	err := rest.UnmarshalRequest(r, &d)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	if _, ok := d["error"]; ok {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("The key 'error' is disallowed in dataset query."))
		return
	}
	if _, ok := d["error_description"]; ok {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("The key 'error_description' is disallowed in datset query."))
		return
	}

	readers := make(map[string]*JsonArrayReader)
	for k, v := range d {
		if v == nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("Invalid query at '%s'", k))
			return
		}
		di, err := v.Get(c.DB)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("Invalid query at '%s': %w", k, err))
			return
		}
		var pi DatapointIterator
		pi = &TransformIterator{dpi: di, it: di}
		defer pi.Close()

		ai, err := NewJsonArrayReader(pi, 2048)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
		readers[k] = ai
	}

	evts, err := json.Marshal(getEvents(d))
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(`{"events":`))
	if err != nil {
		c.Log.Warnf("Dashboard dataset: %s", err.Error())
		return
	}
	_, err = w.Write(evts)
	if err != nil {
		c.Log.Warnf("Dashboard dataset: %s", err.Error())
		return
	}
	_, err = w.Write([]byte(`,"data":{`))
	if err != nil {
		c.Log.Warnf("Dashboard dataset: %s", err.Error())
		return
	}

	hasPrev := false
	for k, v := range readers {
		if hasPrev {
			_, err = w.Write([]byte{','})
			if err != nil {
				c.Log.Warnf("Dataset: %s", err.Error())
				return
			}
		}
		hasPrev = true
		ks, err := json.Marshal(k)
		if err != nil {
			c.Log.Warnf("Dashboard dataset key '%s': %s", k, err.Error())
			return
		}
		_, err = w.Write(ks)
		if err != nil {
			c.Log.Warnf("Dashboard dataset key '%s': %s", k, err.Error())
			return
		}
		_, err = w.Write([]byte(`:`))
		if err != nil {
			c.Log.Warnf("Dashboard dataset key '%s': %s", k, err.Error())
			return
		}
		_, err = io.Copy(w, v)

		if err != nil {
			c.Log.Warnf("Dashboard dataset key '%s': %s", k, err.Error())
			return
		}
	}
	_, err = w.Write([]byte(`}}`))
	if err != nil {
		c.Log.Warnf("Dashboard dataset: %s", err.Error())
		return
	}
}
*/

// Handler is the global router for the timeseries API
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	m.Get("/object/data", func(w http.ResponseWriter, r *http.Request) {
		ReadData(w, r, false)
	})
	m.Delete("/object/data", func(w http.ResponseWriter, r *http.Request) {
		DeleteData(w, r, false)
	})
	m.Post("/object/data", func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, r, false)
	})
	m.Get("/object/timeseries", func(w http.ResponseWriter, r *http.Request) {
		ReadData(w, r, false)
	})
	m.Delete("/object/timeseries", func(w http.ResponseWriter, r *http.Request) {
		DeleteData(w, r, false)
	})
	m.Post("/object/timeseries", func(w http.ResponseWriter, r *http.Request) {
		WriteData(w, r, false)
	})
	m.Get("/object/timeseries/length", func(w http.ResponseWriter, r *http.Request) {
		DataLength(w, r, false)
	})
	/*
		m.Get("/object/actions", func(w http.ResponseWriter, r *http.Request) {
			ReadData(w, r, true)
		})
		m.Delete("/object/actions", func(w http.ResponseWriter, r *http.Request) {
			DeleteData(w, r, true)
		})
		m.Post("/object/actions", func(w http.ResponseWriter, r *http.Request) {
			WriteData(w, r, true)
		})
		m.Get("/object/actions/length", func(w http.ResponseWriter, r *http.Request) {
			DataLength(w, r, true)
		})

		m.Post("/object/act", Act)
	*/

	m.Post("/api/timeseries/dataset", GenerateDataset)

	//m.Post("/dashboard/", GenerateDashboardDataset)

	m.NotFound(rest.NotFoundHandler)
	m.MethodNotAllowed(rest.NotFoundHandler)

	return m
}()
