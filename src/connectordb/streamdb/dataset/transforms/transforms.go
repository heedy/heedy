package transforms

import . "connectordb/streamdb/datastream"

func CopyDatapoint(dp *Datapoint) *Datapoint {
	var result Datapoint
	result.Timestamp = dp.Timestamp
	result.Data = dp.Data //Note: most likely this is not a deep copy
	result.Sender = dp.Sender
	return &result
}

//DatapointTransform is an interface that transforms one datapoint at a time. It is guaranteed
//to be called ordered by datapoints in the stream, so state is allowed to be kept.
//To allow more complicated states, once the DataRange runs out of data, a nil is passed through
//the transform until the transform returns nil, to allow internally queued datapoints to be returned.
//To filter datapoins, returning a null datapoint without error means the daatpoint was filtered (or internally cached)
type DatapointTransform interface {
	Transform(dp *Datapoint) (tdp *Datapoint, err error)
}

//TransformGenerator is the signature of a function that generates a transform
type TransformGenerator func(args []string) (DatapointTransform, error)

//Transforms is the map of all registered transforms
var Transforms = map[string]TransformGenerator{
	"lt":    Lt,
	"gt":    Gt,
	"lte":   Lte,
	"gte":   Gte,
	"eq":    Eq,
	"iflt":  IfLt,
	"ifgt":  IfGt,
	"iflte": IfLte,
	"ifgte": IfGte,
	"ifeq":  IfEq,
}
