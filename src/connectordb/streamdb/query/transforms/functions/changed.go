package functions

import (
	"connectordb/streamdb/datastream"
	"reflect"
)

var changed = Transform{
	Name:         "changed",
	Description:  "Returns true if the current datapoint has a different value than the previous datapoint",
	OutputSchema: `{"type": "boolean"}`,

	Generator: func(name string, args ...TransformFunc) (TransformFunc, error) {
		if len(args) != 0 {
			return Err("sum does not accept arguments")
		}

		var previous *datastream.Datapoint

		return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
			if dp == nil {
				return nil, nil
			}

			iseq := reflect.DeepEqual(dp, previous)
			previous = dp

			returnvalue := dp.Copy()
			returnvalue.Data = !iseq
			return returnvalue, nil

		}, nil
	},
}
