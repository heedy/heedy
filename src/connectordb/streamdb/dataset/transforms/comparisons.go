package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

//ComparisonTransform is the transform that implements <,>,<=,>= comparisons
type ComparisonTransform struct {
	tocompare  float64
	subobject  string
	comparison int
}

func (t *ComparisonTransform) Transform(dp *datastream.Datapoint) (*datastream.Datapoint, error) {

	if dp == nil {
		return nil, nil
	}

	result := CopyDatapoint(dp)

	if t.subobject != "" {
		var ok bool
		result.Data, ok = duck.Get(dp.Data, t.subobject)
		if !ok {
			return nil, errors.New("comparison: Could not find property '" + t.subobject + "'.")
		}
	}

	dnum, ok := duck.Float(result.Data)
	if !ok {
		return nil, errors.New("comparison: Could not convert Datapoint to number (" + duck.JSONString(result) + ")")
	}

	switch t.comparison {
	case 1:
		result.Data = dnum > t.tocompare
	case 2:
		result.Data = dnum >= t.tocompare
	case -1:
		result.Data = dnum < t.tocompare
	case -2:
		result.Data = dnum <= t.tocompare
	default:
		return nil, errors.New("comparison: incorrectly initialized! (internal error)")
	}
	return result, nil
}

func NewComparisonTransform(args []string, comptype int) (*ComparisonTransform, error) {
	if len(args) > 2 || len(args) < 1 {
		return nil, errors.New("comparison: incorrect number of arguments")
	}
	var ok bool
	var result ComparisonTransform
	result.comparison = comptype
	if len(args) == 2 {
		result.subobject = args[0]
		result.tocompare, ok = duck.Float(args[1])
	} else {
		result.tocompare, ok = duck.Float(args[0])
	}
	if !ok {
		return nil, errors.New("comparison: could not convert arg to float")
	}
	return &result, nil
}

func Lt(args []string) (DatapointTransform, error) {
	return NewComparisonTransform(args, -1)
}
func Gt(args []string) (DatapointTransform, error) {
	return NewComparisonTransform(args, 1)
}
func Lte(args []string) (DatapointTransform, error) {
	return NewComparisonTransform(args, -2)
}
func Gte(args []string) (DatapointTransform, error) {
	return NewComparisonTransform(args, 2)
}

//EqualTransform checks equality
type EqualTransform struct {
	compareto string
	subobject string
}

func (t *EqualTransform) Transform(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
	var ok bool
	if dp == nil {
		return nil, nil
	}

	result := CopyDatapoint(dp)

	if t.subobject != "" {

		result.Data, ok = duck.Get(dp.Data, t.subobject)
		if !ok {
			return nil, errors.New("eq: Could not find property '" + t.subobject + "'.")
		}
	}

	result.Data, ok = duck.Eq(result.Data, t.compareto)
	if !ok {
		return nil, errors.New("eq: could not compare")
	}

	return result, nil
}

func Eq(args []string) (DatapointTransform, error) {
	if len(args) > 2 || len(args) < 1 {
		return nil, errors.New("eq: incorrect number of arguments")
	}

	var result EqualTransform
	if len(args) == 2 {
		result.subobject = args[0]
		result.compareto = args[1]
	} else {
		result.compareto = args[0]
	}
	return &result, nil
}
