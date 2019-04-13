package streams

import (
	"encoding/json"
	"reflect"
)

type Datapoint struct {
	Timestamp float64     `json:"t" db:"timestamp" msgpack:"t,omitempty"`
	Data      interface{} `json:"d" db:"data" msgpack:"d,omitempty"`
	Actor     string      `json:"a,omitempty" msgpack:"a,omitempty"`
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

type Query struct {
	T1        *float64 `json:"t1,omitempty"`
	T2        *float64 `json:"t2,omitempty"`
	I1        *int64   `json:"i1,omitempty"`
	I2        *int64   `json:"i2,omitempty"`
	Limit     *int64   `json:"limit,omitempty"`
	Reversed  *bool    `json:"reversed,omitempty"`
	T         *float64 `json:"t,omitempty"`
	I         *int64   `json:"i,omitempty"`
	Transform *string  `json:"transform,omitempty"`
	Actions   *bool    `json:"actions,omitempty"`
}

type InsertQuery struct {
	Actions *bool `json:"actions,omitempty"`

	// insert, append, upsert - default is insert
	Type *string `json:"type,omitempty"`
}

// StreamDB is the interface for storage of stream datapoints
type StreamDB interface {
	WriteStreamData(StreamID string, data DatapointIterator, q *InsertQuery) error
	RemoveStreamData(StreamID string, q *Query) error
	ReadStreamData(StreamID string, q *Query) (DatapointIterator, error)
	StreamDataLength(StreamID string, actor bool) (uint64, error)
}
