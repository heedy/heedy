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

//TransformGenerator is the signature of a function that generates a transform
type TransformGenerator func(args []string) (DatapointTransform, error)

//Transforms is the map of all registered transforms
var Transforms map[string]TransformGenerator

func init() {
	Transforms = map[string]TransformGenerator{
		//comparisons
		"lt":  Lt,
		"gt":  Gt,
		"lte": Lte,
		"gte": Gte,
		"eq":  Eq,
		//ifcomparisons
		"iflt":  IfLt,
		"ifgt":  IfGt,
		"iflte": IfLte,
		"ifgte": IfGte,
		"ifeq":  IfEq,
		"or":    Or,
		"if":    If,
		//object
		"get":   Get,
		"has":   Has,
		"ifhas": IfHas,
	}
}
