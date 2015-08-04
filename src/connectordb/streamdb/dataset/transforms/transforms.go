package transforms

import "connectordb/streamdb/datastream"

func CopyDatapoint(dp *datastream.Datapoint) *datastream.Datapoint {
	var result datastream.Datapoint
	result.Timestamp = dp.Timestamp
	result.Data = dp.Data //Note: most likely this is not a deep copy
	result.Sender = dp.Sender
	return &result
}

//DatapointTransform is an interface that transforms one Datapoint at a time. It is guaranteed
//to be called ordered by Datapoints in the stream, so state is allowed to be kept.
//To allow more complicated states, once the DataRange runs out of data, a nil is passed through
//the transform until the transform returns nil, to allow internally queued Datapoints to be returned.
//To filter datapoins, returning a null Datapoint without error means the daatpoint was filtered (or internally cached)
type DatapointTransform interface {
	Transform(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error)
}

// A straightforward wrapper for functions that adhere to DatapointTransform
type DatapointTransformWrapper struct {
	Transformer func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error)
}

func (d DatapointTransformWrapper) Transform(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
	return d.Transformer(dp)
}

//go:generate go tool yacc -o transform_generator_y.go -p Transform pipeline_generator.y
