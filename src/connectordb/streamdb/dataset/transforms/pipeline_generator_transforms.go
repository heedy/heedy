package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

type TransformFunc func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error)

func handleResultError(prefix string, dp *datastream.Datapoint, err error, coersionOk bool) error {
	if err != nil {
		return err
	}

	if dp == nil {
		return errors.New(prefix + " received nil value")
	}

	if !coersionOk {
		return errors.New(prefix + " Incorrect Type")
	}

	return nil
}

// Does a logical or on the pipeline
func pipelineGeneratorOr(left TransformFunc, right TransformFunc) TransformFunc {

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		result := CopyDatapoint(dp)

		for _, transform := range []TransformFunc{left, right} {

			tdp, err = transform(dp)
			filter, ok := tdp.Data.(bool)

			if err := handleResultError("or", tdp, err, ok); err != nil {
				return nil, err
			}

			if filter {
				result.Data = true
				return result, nil
			}
		}

		result.Data = false
		return result, nil
	}

}

// Does a logical or on the pipeline
func pipelineGeneratorAnd(left TransformFunc, right TransformFunc) TransformFunc {

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		// Process the left data
		tdp, err = left(dp)
		leftRes, ok := tdp.Data.(bool)

		if err := handleResultError("and", tdp, err, ok); err != nil {
			return nil, err
		}

		// Process the right data
		tdp, err = right(dp)
		rightRes, ok := tdp.Data.(bool)

		if err := handleResultError("and", tdp, err, ok); err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = leftRes && rightRes
		return result, nil
	}

}

// Does a logical or on the pipeline
func pipelineGeneratorNot(right TransformFunc) TransformFunc {

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		// Process the right data
		tdp, err = right(dp)
		rightRes, ok := tdp.Data.(bool)

		if err := handleResultError("not", tdp, err, ok); err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = !rightRes
		return result, nil
	}

}

func pipelineGeneratorCompare(left, right TransformFunc, operator string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {

		if dp == nil {
			return nil, nil
		}

		leftResult, err := left(dp)
		if err := handleResultError("compare", leftResult, err, true); err != nil {
			return nil, err
		}

		rightResult, err := left(dp)
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
			return result, errors.New("comparison: invalid comparison types")
		}

		return result, nil
	}
}

// The identity function, returns whatever was passed in.
func pipelineGeneratorConstant(value interface{}, inputError error) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		result := CopyDatapoint(dp)
		result.Data = value
		return result, inputError
	}
}

func pipelineGeneratorIdentity() TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		return dp, nil
	}
}

func pipelineGeneratorGet(propertyName string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		var ok bool
		result := CopyDatapoint(dp)
		result.Data, ok = duck.Get(dp.Data, propertyName)

		if !ok {
			return nil, errors.New("Could not find element '" + propertyName + "' in " + duck.JSONString(dp))
		}

		return result, nil
	}
}

func pipelineGeneratorHas(propertyName string) TransformFunc {
	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}
		result := CopyDatapoint(dp)
		_, result.Data = duck.Get(dp.Data, propertyName)
		return result, nil

	}
}
