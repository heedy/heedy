package functions

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/query/transforms"
	"errors"

	"github.com/connectordb/duck"
)

var sum = transforms.Transform{
	Name:         "sum",
	Description:  "Returns the running sum of the values in the datapoints seen",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,

	Generator: func(name string, args ...transforms.TransformFunc) (transforms.TransformFunc, error) {
		if len(args) != 0 {
			return transforms.Err("sum does not accept arguments")
		}

		total := float64(0)

		return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
			if dp == nil {
				return nil, nil
			}
			val, ok := duck.Float(dp.Data)
			if !ok {
				return nil, errors.New("sum cannot convert datapoint to number")
			}

			total += val

			returnvalue := dp.Copy()
			returnvalue.Data = total
			return returnvalue, nil

		}, nil
	},
}
