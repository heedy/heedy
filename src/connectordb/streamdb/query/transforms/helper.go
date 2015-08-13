package transforms

import (
	"connectordb/streamdb/datastream"
	"errors"
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

func getCustomFunction(identifier string, children ...TransformFunc) (TransformFunc, error) {
	return Get(identifier, children...)
}
