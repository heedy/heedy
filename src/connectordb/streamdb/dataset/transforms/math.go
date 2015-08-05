package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"

	"github.com/connectordb/duck"
)

func getTransformFloat(dp *datastream.Datapoint, function TransformFunc) (float64, error) {
	if dp == nil {
		return 0, errors.New("Nil datapoint")
	}

	transformed, err := function(dp)
	if err != nil {
		return 0, err
	}

	result, ok := duck.Float(transformed.Data)

	if !ok {
		return result, errors.New("not a number")
	}

	return result, nil
}

func getTransformLeftRight(dp *datastream.Datapoint, left, right TransformFunc) (leftVal float64, rightVal float64, err error) {
	leftVal, err = getTransformFloat(dp, left)
	if err != nil {
		return 0, 0, err
	}

	rightVal, err = getTransformFloat(dp, right)
	if err != nil {
		return 0, 0, err
	}

	return leftVal, rightVal, err
}

// Adds the left and right hand side
func addTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		left, right, err := getTransformLeftRight(dp, left, right)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = left + right

		return result, nil
	}
}

// Multiplies the left and right hand side
func multiplyTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		left, right, err := getTransformLeftRight(dp, left, right)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = left * right

		return result, nil
	}
}

// Divides the left and right hand side
func divideTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		left, right, err := getTransformLeftRight(dp, left, right)
		if err != nil {
			return nil, err
		}

		if right == 0.0 {
			return nil, errors.New("zero division error")
		}

		result := CopyDatapoint(dp)
		result.Data = left / right

		return result, nil
	}
}

// Subtracts the left and right hand side
func subtractTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		left, right, err := getTransformLeftRight(dp, left, right)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = left - right

		return result, nil
	}
}

// Subtracts the left and right hand side
func inverseTransformGenerator(transform TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		right, err := getTransformFloat(dp, transform)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = -right

		return result, nil
	}
}
