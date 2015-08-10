package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
	"fmt"
)

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

// Calls transform and tries to read a bool, fails on no bool or error
func readBool(prefix string, dp *datastream.Datapoint, transform TransformFunc) (bool, error) {
	if dp == nil || dp.Data == nil {
		return false, errors.New(prefix + " received nil value")
	}

	tdp, err := transform(dp)
	if err != nil {
		return false, err
	}
	if tdp.Data == nil {
		return false, errors.New(prefix + " received nil value after transform")
	}

	filter, ok := tdp.Data.(bool)

	if err := handleResultError(prefix, tdp, err, ok); err != nil {
		return false, err
	}

	return filter, nil
}

func logTransform(child TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			fmt.Printf("Nil datapoint input")
			return nil, nil
		}

		fmt.Printf("processing: %v\n", dp.String())
		res, err := child(dp)

		if res != nil {
			fmt.Printf("result: %v err: %v\n", res.String(), err)
		} else {
			fmt.Printf("nil result err: %v\n", err)
		}

		return res, err
	}
}

func getCustomFunction(identifier string, children ...TransformFunc) (TransformFunc, error) {
	return Get(identifier, children...)
}
