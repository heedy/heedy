/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"
	"errors"

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

type DatasetNullChecker struct {
	Data map[string]*DatasetRangeElement
	Iter pipescript.DatapointIterator
}

//Close closes the open DataRanges
func (dr *DatasetNullChecker) Close() {
	for key := range dr.Data {
		dr.Data[key].Close()
	}
}

//Next gets the next datapoint from the DatasetRange
func (dr *DatasetNullChecker) Next() (*pipescript.Datapoint, error) {
	dp, err := dr.Iter.Next()
	if err != nil || dp == nil {
		return nil, err
	}

	// The datapoint exists. Now ensure that it is not nil
	v, ok := dp.Data.(map[string]interface{})
	if !ok {
		// It is not OK - something went wrong...
		return nil, errors.New("Data does not conform to dataset format... Something went wrong!")
	}

	for key := range dr.Data {
		v2, ok := v[key]
		if !dr.Data[key].AllowNil && (v2 == nil || !ok) {
			return dr.Next() // Recursively remove nils
		}
	}

	return dp, nil

}

// The DatasetRange is split into a DatasetNullChecker, which checks the dataset keys for null,
// and the Iter resulting from adding the DatasetNullChecker. The reason the component had to be split
// into two parts is because the posttransform can only be applied AFTER the null checker, so Iter might
// actually be the post-transform
type DatasetRange struct {
	Dnc  *DatasetNullChecker
	Iter pipescript.DatapointIterator
}

func (dc *DatasetRange) Close() {
	dc.Dnc.Close()
}

func (dc *DatasetRange) Next() (*datastream.Datapoint, error) {
	dp, err := dc.Iter.Next()
	if err != nil || dp == nil {
		return nil, err
	}
	return &datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil
}
