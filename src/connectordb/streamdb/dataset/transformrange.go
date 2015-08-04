package dataset

import (
	"connectordb/streamdb/dataset/transforms"
	"connectordb/streamdb/datastream"
)

//TransformArray transforms the given array. Note: Since it assumes that the transform is happening
//within a stream, it does not pass through nils, as would be needed if the transform got to
//the end of a stream range.
func TransformArray(t transforms.DatapointTransform, dpa *datastream.DatapointArray) (*datastream.DatapointArray, error) {
	if dpa == nil {
		//If the DatapointArray is nil, return the nil-cache of the DatapointTransform

		resultarray := make(datastream.DatapointArray, 0)
		dp, err := t.Transform(nil)
		for err == nil && dp != nil {
			resultarray = append(resultarray, *dp)
			dp, err = t.Transform(nil)
		}
		return &resultarray, err

	}
	resultarray := make(datastream.DatapointArray, 0, dpa.Length())
	for i := 0; i < dpa.Length(); i++ {
		dp, err := t.Transform(&((*dpa)[i]))
		if err != nil {
			return nil, err
		}
		if dp != nil {
			resultarray = append(resultarray, *dp)
		}
	}
	return &resultarray, nil
}

type TransformRange struct {
	Data      datastream.DataRange
	Transform transforms.DatapointTransform
}

//Index returns the index of the next datapoint in the underlying DataRange - it does not guarantee that the datapoint won't be filtered by the
//underlying transforms
func (t *TransformRange) Index() int64 {
	return t.Data.Index()
}

//Close closes the underlying DataRange
func (t *TransformRange) Close() {
	t.Data.Close()
}

//Next iterates through a datarange until a datapoint is returned by the transform
func (t *TransformRange) Next() (dp *datastream.Datapoint, err error) {
	for {

		dp1, err := t.Data.Next()
		if err != nil {
			return nil, err
		}
		dp, err = t.Transform.Transform(dp1)
		if err != nil || dp != nil {
			return dp, err
		}
		if dp1 == nil && dp == nil {
			return nil, nil
		}
	}
}

//NextArray is here to fit into the DataRange interface - given a batch of data from the underlying
//data store, returns the DatapointArray of transformed data
func (t *TransformRange) NextArray() (da *datastream.DatapointArray, err error) {
	for {

		da1, err := t.Data.NextArray()
		if err != nil {
			return nil, err
		}
		da, err = TransformArray(t.Transform, da1)
		if err != nil || len(*da) > 0 {
			return da, err
		}
		if da1 == nil && (da == nil || len(*da) == 0) {
			return nil, nil
		}
	}
}

//NewTransformRange generates a transform range from a transfrom pipeline
func NewTransformRange(dr datastream.DataRange, transformpipeline string) (*TransformRange, error) {
	t, err := transforms.NewTransformPipeline(transformpipeline)
	if err != nil {
		return nil, err
	}
	return &TransformRange{
		Data:      dr,
		Transform: t,
	}, nil
}
