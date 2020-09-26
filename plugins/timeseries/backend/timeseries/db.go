package timeseries

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/pipescript"
	"github.com/karrick/tparse/v2"
)

type Datapoint struct {
	Timestamp float64     `json:"t" db:"timestamp" msgpack:"t,omitempty"`
	Duration  float64     `json:"dt,omitempty" db:"duration" msgpack:"dt,omitempty"`
	Data      interface{} `json:"d" db:"data" msgpack:"d,omitempty"`

	Actor string `json:"a,omitempty" db:"actor" msgpack:"a,omitempty"`
}

//IsEqual checks if the datapoint is equal to another datapoint
func (d *Datapoint) IsEqual(dp *Datapoint) bool {
	return (dp.Timestamp == d.Timestamp && dp.Actor == d.Actor && reflect.DeepEqual(d.Data, dp.Data))
}

// String returns a json representation of the datapoint
func (d *Datapoint) String() string {
	b, _ := json.Marshal(d)
	return string(b)
}

// NewDatapoint returns a datapoint with the current timestamp
func NewDatapoint(data interface{}) *Datapoint {
	return &Datapoint{
		Timestamp: float64(time.Now().UnixNano()) * 1e-9,
		Data:      data,
	}
}

//A DatapointArray holds a couple useful functions that act on it
type DatapointArray []*Datapoint

// String returns a json representation of the datapoint
func (dpa DatapointArray) String() string {
	b, _ := json.Marshal(dpa)
	return string(b)
}

//IsEqual checks if two DatapointArrays contain the same data
func (dpa DatapointArray) IsEqual(d DatapointArray) bool {
	if len(d) != len(dpa) {
		return false
	}
	for i := range d {
		if !d[i].IsEqual(dpa[i]) {
			return false
		}
	}
	return true
}

type DatapointIterator interface {
	Next() (*Datapoint, error)
	Close() error
}

// ChanIterator runs the iteration in a goroutine, so that post-processing data and pre-processing
// can happen in parallel
type ChanIterator struct {
	closer      chan bool
	datapointer chan *Datapoint
	err         error
}

func (c *ChanIterator) Close() error {
	if c.closer != nil {
		c.closer <- true
		c.closer = nil
	}
	return nil
}

func (c *ChanIterator) Next() (*Datapoint, error) {
	dp := <-c.datapointer
	if dp == nil {
		return dp, c.err
	}
	return dp, nil
}

func NewChanIterator(di DatapointIterator) *ChanIterator {
	closer := make(chan bool)
	datapointer := make(chan *Datapoint, 10000)
	ci := &ChanIterator{
		closer:      make(chan bool, 1),
		datapointer: datapointer,
		err:         nil,
	}

	go func() {
		for {
			dp, err := di.Next()
			if err != nil {
				ci.err = err
				dp = nil
			}

			select {
			case datapointer <- dp:
			case <-closer:
				close(datapointer)
				return
			}
			if dp == nil {
				return
			}
		}
	}()
	return ci
}

type Query struct {
	Timeseries string      `json:"timeseries,omitempty"`
	T1         interface{} `json:"t1,omitempty"`
	T2         interface{} `json:"t2,omitempty"`
	I1         *int64      `json:"i1,omitempty"`
	I2         *int64      `json:"i2,omitempty"`
	Limit      *int64      `json:"limit,omitempty"`
	Reversed   *bool       `json:"reversed,omitempty"`
	T          interface{} `json:"t,omitempty"`
	I          *int64      `json:"i,omitempty"`
	Transform  *string     `json:"transform,omitempty"`
	Actions    *bool       `json:"actions,omitempty"`
}

func (q *Query) RawRead(adb *database.AdminDB) (*SQLIterator, error) {
	query, values, err := querySQL(q, true)
	if err != nil {
		return nil, err
	}
	if q.Actions != nil && *q.Actions {
		rows, err := adb.Queryx("SELECT timestamp,duration,actor,data FROM "+query, values...)

		return &SQLIterator{rows.Rows, true}, err
	}
	rows, err := adb.Queryx("SELECT timestamp,duration,data FROM "+query, values...)

	return &SQLIterator{rows.Rows, false}, err
}

func (q *Query) Get(db database.DB, tstart float64) (*DatasetIterator, error) {
	if q.T1 == nil && q.I1 == nil && q.T == nil && q.I == nil {
		q.T1 = tstart
	}

	obj, err := db.ReadObject(q.Timeseries, &database.ReadObjectOptions{
		Icon: false,
	})
	if err != nil {
		return nil, err
	}
	if *obj.Type != "timeseries" {
		return nil, fmt.Errorf("bad_query: Object '%s' is not a timeseries", q.Timeseries)
	}
	if !obj.Access.HasScope("read") {
		return nil, errors.New("access_denied: The given object can't be read")
	}

	iter, err := q.RawRead(db.AdminDB())

	var piter pipescript.Iterator
	piter = PipeIterator{iter}

	if q.Transform != nil && *q.Transform != "" {
		p, err := pipescript.Parse(*q.Transform)
		if err != nil {
			iter.Close()
			return nil, err
		}
		p.InputIterator(piter)
		piter = p
	}

	return &DatasetIterator{
		closers: []Closer{iter},
		it:      piter,
	}, nil
}

func ParseTimestamp(ts interface{}) (float64, error) {
	tss, ok := ts.(string)
	if ok {
		t, err := tparse.ParseNow(time.RFC3339, tss)
		return Unix(t), err
	}
	f, ok := ts.(float64)
	if ok {
		return f, nil
	}
	return 0, errors.New("Could not parse timestamp")
}

type InsertQuery struct {
	Actions *bool `json:"actions,omitempty"`

	// insert, append, update - default is update
	Method *string `json:"method,omitempty"`
}

func Unix(t time.Time) float64 {
	return float64(t.UnixNano()) * 1e-9
}

type PipeIterator struct {
	it DatapointIterator
}

func (pi PipeIterator) Next(out *pipescript.Datapoint) (*pipescript.Datapoint, error) {
	dp, err := pi.it.Next()
	if dp == nil || err != nil {
		return nil, err
	}
	out.Timestamp = dp.Timestamp
	out.Duration = dp.Duration
	out.Data = dp.Data
	return out, nil
}

type FromPipeIterator struct {
	dpi Closer
	it  pipescript.Iterator
	dp  pipescript.Datapoint
}

func (pi *FromPipeIterator) Next() (*Datapoint, error) {
	dp, err := pi.it.Next(&pi.dp)
	if dp == nil || err != nil {
		return nil, err
	}
	return &Datapoint{
		Timestamp: dp.Timestamp,
		Duration:  dp.Duration,
		Data:      dp.Data,
	}, nil
}

func (pi *FromPipeIterator) Close() error {
	return pi.dpi.Close()
}

func MkTransform(transform string, it DatapointIterator) (DatapointIterator, error) {
	p, err := pipescript.Parse(transform)
	if err != nil {
		return nil, err
	}
	p.InputIterator(PipeIterator{it})
	return &FromPipeIterator{dpi: it, it: p}, nil
}
