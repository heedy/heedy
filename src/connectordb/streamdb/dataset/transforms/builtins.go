package transforms

// This file provides the basic built-in functions the interpolator uses.

import (
	"connectordb/streamdb/datastream"
	"errors"
)

func init() {
	RegisterCustomFunction("gte", implicitComparisonTransform)
	RegisterCustomFunction("gt", implicitComparisonTransform)
	RegisterCustomFunction("lte", implicitComparisonTransform)
	RegisterCustomFunction("lt", implicitComparisonTransform)
	RegisterCustomFunction("eq", implicitComparisonTransform)
	RegisterCustomFunction("ne", implicitComparisonTransform)

	RegisterCustomFunction("sum", SumTransform)
}

// Implicitly compares the datapoint
func implicitComparisonTransform(name string, children ...TransformFunc) (TransformFunc, error) {
	// We need one child transform.
	if len(children) != 1 {
		return pipelineGeneratorIdentity(), errors.New("gte() Exactly one child required.")
	}

	// This is the thing we're testing gte from.
	child := children[0]

	// The identity function gets the existing value
	identity := pipelineGeneratorIdentity()

	switch name {
	case "gte":
		return pipelineGeneratorCompare(identity, child, ">="), nil
	case "gt":
		return pipelineGeneratorCompare(identity, child, ">"), nil
	case "lte":
		return pipelineGeneratorCompare(identity, child, "<="), nil
	case "lt":
		return pipelineGeneratorCompare(identity, child, "<"), nil
	case "eq":
		return pipelineGeneratorCompare(identity, child, "=="), nil
	case "ne":
		return pipelineGeneratorCompare(identity, child, "!="), nil
	default:
		return identity, errors.New("Internal comparison error, unknown comparison" + name)
	}
}

// Sums the values of the datapoints passing through
func SumTransform(name string, children ...TransformFunc) (TransformFunc, error) {

	// We need one child transform.
	if len(children) != 0 {
		return pipelineGeneratorIdentity(), errors.New("sum() required: no children.")
	}

	total := 0.0
	identity := pipelineGeneratorIdentity()

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return dp, nil
		}

		// Gets the floating point value of a datpoint
		value, err := getTransformFloat(dp, identity)
		if err != nil {
			return dp, err
		}

		// save the running total
		total += value

		// return a copy of the result
		result := dp.Copy()
		result.Data = total
		return result, nil
	}, nil
}
