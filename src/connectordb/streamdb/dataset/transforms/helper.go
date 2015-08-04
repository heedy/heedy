package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
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

// Calls transform and tries to read a bool, fails on no bool or error
func readBool(prefix string, dp *datastream.Datapoint, transform TransformFunc) (bool, error) {
	if dp == nil {
		return false, errors.New(prefix + " received nil value")
	}

	tdp, err := transform(dp)
	filter, ok := tdp.Data.(bool)

	if err := handleResultError(prefix, tdp, err, ok); err != nil {
		return false, err
	}

	return filter, nil
}
