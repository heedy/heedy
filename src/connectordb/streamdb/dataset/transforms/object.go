package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

//GetTransform takes a datapoint, and returns the given property
type GetTransform struct {
	property string
}

func (t GetTransform) Transform(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
	if dp == nil {
		return nil, nil
	}
	var ok bool
	result := CopyDatapoint(dp)
	result.Data, ok = duck.Get(dp.Data, t.property)
	if !ok {
		return nil, errors.New("Could not find element '" + t.property + "' in " + duck.JSONString(dp))
	}
	return result, nil
}

func Get(args []string) (DatapointTransform, error) {
	if len(args) >= 2 || len(args) < 1 {
		return nil, errors.New("get: incorrect number of arguments")
	}
	return GetTransform{args[0]}, nil
}

type HasTransform struct {
	property string
}

func (t HasTransform) Transform(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
	if dp == nil {
		return nil, nil
	}
	result := CopyDatapoint(dp)
	_, result.Data = duck.Get(dp.Data, t.property)
	return result, nil
}

func Has(args []string) (DatapointTransform, error) {
	if len(args) >= 2 || len(args) < 1 {
		return nil, errors.New("get: incorrect number of arguments")
	}
	return HasTransform{args[0]}, nil
}

func IfHas(args []string) (DatapointTransform, error) {
	t, err := Has(args)
	return BooleanFilterTransform{t}, err
}
