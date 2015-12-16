/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"
	"errors"
)

//Operator is an interface describing the functions that are needed for query. The standard operator implements these,
//	but for import sake and for simplified mocking, only the necessary interface is shown here
type Operator interface {
	GetStreamIndexRange(streampath string, i1 int64, i2 int64, transform string) (datastream.DataRange, error)
	GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error)
	GetShiftedStreamTimeRange(streampath string, t1 float64, t2 float64, ishift, limit int64, transform string) (datastream.DataRange, error)
}

//StreamQuery contains all the necessary information to perform a query on the given stream. It is the structure used
//to encode a query for merge and dataset. It uses the Operator's functions internally.
//Note that while both index-based and time based elements are in the struct, it is only valid to use one at a time
type StreamQuery struct {
	Stream    string  `json:"stream"`              //The stream name in form usr/dev/stream
	Transform string  `json:"transform,omitempty"` //The transform to perform on the stream
	I1        int64   `json:"i1,omitempty"`        //The first index to get
	I2        int64   `json:"i2,omitempty"`        //The end index of the range to get
	T1        float64 `json:"t1,omitempty"`        //The start time of the range to get
	T2        float64 `json:"t2,omitempty"`        //The end time of the range to get
	Limit     int64   `json:"limit,omitempty"`     //The limit of number of datapoints to allow

	indexbacktrack int64 //The number of elements to backtrack before a starting time (used for time queries)
}

//IsValid checks if the StreamQuery encodes a valid query. It does not check whether
//the Stream is a valid stream, but only that it exists
func (s *StreamQuery) IsValid() bool {
	return s.Stream != ""
}

//HasRange returns True if the stream has some sort of range-based query non-zero
//meaning that one of the indices or times or limit is non-zero
func (s *StreamQuery) HasRange() bool {
	return s.I1 != 0 || s.I2 != 0 || s.T1 != 0 || s.T2 != 0 || s.Limit != 0
}

//Run runs the query that the struct encodes on the given operator
func (s *StreamQuery) Run(qm Operator) (datastream.DataRange, error) {

	if s.T1 != 0 || s.T2 != 0 || s.Limit != 0 {
		//First check that only one method of querying is active
		if s.I1 != 0 || s.I2 != 0 {
			//query by index is also active. Not cool. Not cool at all
			return nil, errors.New("Only one query method (index or time) can be used at a time")
		}

		//Alright, query by time
		if s.indexbacktrack > 0 {
			return qm.GetShiftedStreamTimeRange(s.Stream, s.T1, s.T2, -s.indexbacktrack, s.Limit, s.Transform)
		}
		return qm.GetStreamTimeRange(s.Stream, s.T1, s.T2, s.Limit, s.Transform)
	}

	//The query method is by integer (or no query method is chosen, meaning whole stream)
	return qm.GetStreamIndexRange(s.Stream, s.I1, s.I2, s.Transform)
}
