package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

func pipelineGeneratorCompare(left, right TransformFunc, operator string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {

		if dp == nil {
			return nil, nil
		}

		leftResult, err := left(dp)
		if err := handleResultError("compare", leftResult, err, true); err != nil {
			return nil, err
		}

		rightResult, err := right(dp)
		if err := handleResultError("compare", rightResult, err, true); err != nil {
			return nil, err
		}

		var ok bool
		result := CopyDatapoint(dp)

		switch operator {
		case ">":
			result.Data, ok = duck.Gt(leftResult.Data, rightResult.Data)
		case ">=":
			result.Data, ok = duck.Gte(leftResult.Data, rightResult.Data)
		case "<":
			result.Data, ok = duck.Lt(leftResult.Data, rightResult.Data)
		case "<=":
			result.Data, ok = duck.Lte(leftResult.Data, rightResult.Data)
		case "!=":
			var eq bool
			eq, ok = duck.Eq(leftResult.Data, rightResult.Data)
			result.Data = !eq
		case "==":
			result.Data, ok = duck.Eq(leftResult.Data, rightResult.Data)
		default:
			return nil, errors.New("comparison: incorrectly initialized! (internal error)")
		}

		if ok != true {
			return nil, errors.New("comparison: invalid comparison types")
		}

		return result, nil
	}
}
