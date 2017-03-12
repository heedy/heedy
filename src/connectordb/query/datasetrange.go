/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"

	"github.com/connectordb/pipescript"
	"github.com/connectordb/pipescript/interpolator"
)

//DatasetRangeElement is the element that includes the interpolator and transform for a given dataset
type DatasetRangeElement struct {
	Interpolator interpolator.InterpolatorInstance
	Range        datastream.DataRange
	AllowNil     bool
}

//Close closes the internal database connections
func (dre *DatasetRangeElement) Close() {
	dre.Range.Close()
}

type DatasetRange struct {
	Data map[string]*DatasetRangeElement
	Iter pipescript.DatapointIterator
}

//Close closes the open DataRanges
func (dr *DatasetRange) Close() {
	for key := range dr.Data {
		dr.Data[key].Close()
	}
}

//Next gets the next datapoint from the DatasetRange
func (dr *DatasetRange) Next() (*datastream.Datapoint, error) {
	dp, err := dr.Iter.Next()
	if err != nil || dp == nil {
		return nil, err
	}

	// The datapoint exists. Now ensure that it is not nil
	v, ok := dp.Data.(map[string]interface{})
	if !ok {
		// if it is not OK, it means that there was a transform that was used to generate the final values
		return &datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil
	}

	for key := range dr.Data {
		v2, ok := v[key]
		if !dr.Data[key].AllowNil && (v2 == nil || !ok) {
			return dr.Next() // Recursively remove nils
		}
	}

	return &datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil

}
