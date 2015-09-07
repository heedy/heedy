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

		return func(te *transforms.TransformEnvironment) *transforms.TransformEnvironment {
			if !te.CanProcess() {
				return te
			}

			iseq := reflect.DeepEqual(te.Datapoint, previous)
			previous = te.Datapoint

			return te.Copy().SetData(!iseq)
		}, nil
	},
}
