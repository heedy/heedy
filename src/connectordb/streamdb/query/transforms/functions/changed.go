package functions

import (
	"connectordb/streamdb/datastream"

	"connectordb/streamdb/query/transforms"
	"reflect"
)

var changed = transforms.Transform{
	Name:         "changed",
	Description:  "Returns true if the current datapoint has a different value than the previous datapoint",
	OutputSchema: `{"type": "boolean"}`,

	Generator: func(name string, args ...transforms.TransformFunc) (transforms.TransformFunc, error) {
		if len(args) != 0 {
			return transforms.Err("changed does not accept arguments")
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
