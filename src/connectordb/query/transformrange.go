/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package query

import (
	"config"
	"connectordb/datastream"

	"github.com/connectordb/pipescript"

	"github.com/connectordb/pipescript/transforms" // Load all available transforms
)

// Register all of pipescript's standard library of transforms
func init() {
	transforms.Register()
}

//TransformArray transforms the given array.
func TransformArray(t *pipescript.Script, dpa *datastream.DatapointArray) (*datastream.DatapointArray, error) {
	// ASSUMING THAT THE SCRIPT IS CLEARED OR UNINITIALIZED
	// Create an array range from the datapoint array, convert it to pipescript iterator, and set as script input
	t.SetInput(&DatapointIterator{datastream.NewDatapointArrayRange(*dpa, 0)})

	resultarray := make(datastream.DatapointArray, 0, dpa.Length())
	for {
		dp, err := t.Next()
		if err != nil {
			return nil, err
		}
		if dp == nil {
			return &resultarray, nil
		}
		resultarray = append(resultarray, datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data})

	}
}

//ExtendedTransformRange is an ExtendedDataRange which passes data through a transform.
type ExtendedTransformRange struct {
	Data      datastream.ExtendedDataRange
	Transform *pipescript.Script
}

//Index returns the index of the next datapoint in the underlying ExtendedDataRange - it does not guarantee that the datapoint won't be filtered by the
//underlying transforms. It also does not guarantee that it is the correct datapoint, as transforms are free to peek into the data sequence.
func (t *ExtendedTransformRange) Index() int64 {
	return t.Data.Index()
}

//Close closes the underlying ExtendedDataRange
func (t *ExtendedTransformRange) Close() {
	t.Data.Close()
}

//Next gets the next datapoint
func (t *ExtendedTransformRange) Next() (*datastream.Datapoint, error) {
	dp, err := t.Transform.Next()
	if err != nil {
		return nil, err
	}
	if dp == nil {
		return nil, nil
	}
	// Convert pipescript datapoint to datastream datapoint
	return &datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil
}

// NextArray is here to fit into the ExtendedDataRange interface - given a batch of data from the underlying
//data store, returns the DatapointArray of transformed data. Since transforms can be filters and have no concept of batching (yet),
// We just get ~250 datapoints the standard way and pretend that's our batch.
// TODO: Use PipeScript batching when available
func (t *ExtendedTransformRange) NextArray() (da *datastream.DatapointArray, err error) {
	bs := config.Get().BatchSize
	resultarray := make(datastream.DatapointArray, 0, bs)
	for i := 0; i < bs; i++ {
		dp, err := t.Next()
		if err != nil {
			return nil, err
		}
		if dp == nil {
			return &resultarray, nil
		}
		resultarray = append(resultarray, *dp)
	}
	return &resultarray, nil
}

//NewExtendedTransformRange generates a transform range from a transfrom pipeline
func NewExtendedTransformRange(dr datastream.ExtendedDataRange, transformpipeline string) (*ExtendedTransformRange, error) {
	t, err := pipescript.Parse(transformpipeline)
	if err != nil {
		return nil, err
	}
	t.SetInput(&DatapointIterator{dr})

	return &ExtendedTransformRange{
		Data:      dr,
		Transform: t,
	}, nil
}

//TransformRange is ExtendedTransformRange's little brother - it works on DataRanges
type TransformRange struct {
	Data      datastream.DataRange
	Transform *pipescript.Script
}

//Close closes the underlying ExtendedDataRange
func (t *TransformRange) Close() {
	t.Data.Close()
}

//Next iterates through a datarange until a datapoint is returned by the transform
func (t *TransformRange) Next() (*datastream.Datapoint, error) {
	dp, err := t.Transform.Next()
	if err != nil {
		return nil, err
	}
	if dp == nil {
		return nil, nil
	}
	// Convert pipescript datapoint to datastream datapoint
	return &datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil
}

//NewTransformRange generates a transform range from a transfrom pipeline
func NewTransformRange(dr datastream.ExtendedDataRange, transformpipeline string) (*TransformRange, error) {
	t, err := pipescript.Parse(transformpipeline)
	if err != nil {
		return nil, err
	}
	t.SetInput(&DatapointIterator{dr})

	return &TransformRange{
		Data:      dr,
		Transform: t,
	}, nil
}
