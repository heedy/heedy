package functions

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/query/transforms"
	"container/list"
	"errors"

	"github.com/connectordb/duck"
)

var sum = transforms.Transform{
	Name:         "sum",
	Description:  "Returns the running sum of the values in the datapoints seen. If given an argument, returns the sum of the last n datapoints.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The number of datapoints backwards from the current datapoint to sum over",
			Constant:    true,
			Optional:    true,
		},
	},

	Generator: func(name string, args ...transforms.TransformFunc) (transforms.TransformFunc, error) {
		if len(args) > 1 {
			return transforms.Err("sum must have at most one argument")
		}

		if len(args) == 0 {
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
		}

		//Set up a linked list of the datapoints within the wanted number
		argval, err := args[0](nil)
		if err != nil || argval == nil {
			return transforms.Err("sum requires a constant argument.")
		}
		num, ok := duck.Int(argval.Data)
		if !ok {
			return transforms.Err("The argument to sum must be an integer")
		}
		if num <= 1 || num > 1000 {
			return transforms.Err("average must be called with 1000 >= arg > 1")
		}

		cursum := float64(0)
		//The linked list of the last num datapoints
		dplist := list.New()

		return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
			if dp == nil {
				return nil, nil
			}

			val, ok := duck.Float(dp.Data)
			if !ok {
				return nil, errors.New("sum could not convert datapoint to number")
			}

			cursum += val
			dplist.PushFront(val)

			if dplist.Len() > int(num) {
				elem := dplist.Back()
				cursum -= elem.Value.(float64)
				dplist.Remove(elem)
			}

			returnval := dp.Copy()
			returnval.Data = cursum
			return returnval, nil
		}, nil
	},
}
