package api

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
)

// DataValidator performs all of the validation necessary on a stream for it to conform to a permissions-based
// system. This includes validating the schema and ensuring that the actor is set correctly.
type DataValidator struct {
	data   DatapointIterator
	schema *gojsonschema.Schema
	actor  string
}

// NewDataValidator ensures that the stream data fits the given schema and has actor set properly
func NewDataValidator(data DatapointIterator, schema interface{}, actor string) (*DataValidator, error) {
	s, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(schema))
	if err != nil {
		return nil, err
	}
	return &DataValidator{
		data:   data,
		schema: s,
		actor:  actor,
	}, nil
}

// Next sets the actor, and performs schema validation
func (s *DataValidator) Next() (*Datapoint, error) {
	dp, err := s.data.Next()
	if dp == nil || err != nil {
		return dp, err
	}

	result, err := s.schema.Validate(gojsonschema.NewGoLoader(dp.Data))
	if err != nil {
		s.data.Close()
		return dp, err
	}
	if !result.Valid() {
		s.data.Close()
		return dp, errors.New("bad_query: The data failed schema validation")
	}
	if s.actor != "" {
		dp.Actor = s.actor
	}
	return dp, nil
}

// Close closes the underlying stream
func (s *DataValidator) Close() error {
	return s.data.Close()
}

// ------------------------------------------------------------------------------------------------------

//DatapointArrayIterator allows DatapointArray to conform to the DatapointIterator interface
type DatapointArrayIterator struct {
	rangeindex int
	da         DatapointArray
}

//Close resets the range
func (d *DatapointArrayIterator) Close() error {
	d.rangeindex = 0
	return nil
}

//Index returns the index of the DatapointArray
func (d *DatapointArrayIterator) Index() int64 {
	return int64(d.rangeindex)
}

//Next returns the next datapoint
func (d *DatapointArrayIterator) Next() (*Datapoint, error) {
	if d.rangeindex >= len(d.da) {
		return nil, nil
	}
	d.rangeindex++
	return d.da[d.rangeindex-1], nil
}

//NextArray returns what is left of the array
func (d *DatapointArrayIterator) NextArray() (DatapointArray, error) {
	if d.rangeindex >= len(d.da) {
		return nil, nil
	}
	dpa := d.da[d.rangeindex:]
	d.rangeindex = len(d.da)
	return dpa, nil
}

//NewDatapointArrayIterator does exactly what the function says
func NewDatapointArrayIterator(da DatapointArray) *DatapointArrayIterator {
	return &DatapointArrayIterator{0, da}
}

// ------------------------------------------------------------------------------------------------------

//NumIterator returns only the first given number of datapoints (with an optional skip param) from a DatapointIterator
type NumIterator struct {
	di      DatapointIterator
	numleft int64 //The number of datapoints left to return
}

//Close closes the internal DatapointIterator
func (r *NumIterator) Close() error {
	return r.di.Close()
}

//Next returns the next datapoint from the underlying DatapointIterator so long as the datapoint is within the
//amonut of datapoints to return.
func (r *NumIterator) Next() (*Datapoint, error) {
	if r.numleft == 0 {
		r.di.Close()
		return nil, nil
	}
	r.numleft--
	return r.di.Next()
}

//Skip the given number of datapoints without changing the number of datapoints left to return
func (r *NumIterator) Skip(num int) error {
	for i := 0; i < num; i++ {
		_, err := r.di.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

//NewNumIterator initializes a new NumIterator which will return up to the given amount of datapoints.
func NewNumIterator(dr DatapointIterator, datapoints int64) *NumIterator {
	return &NumIterator{dr, datapoints}
}

// ------------------------------------------------------------------------------------------------------

// NewArrayFromIterator creates a datapoint array from the given iterator
func NewArrayFromIterator(di DatapointIterator) (DatapointArray, error) {
	d := DatapointArray{}

	dp, err := di.Next()
	for dp != nil && err == nil {
		d = append(d, dp)
		dp, err = di.Next()
	}
	return d, err
}
