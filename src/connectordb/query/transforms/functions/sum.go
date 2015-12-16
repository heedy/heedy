/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package functions

import (
	"connectordb/query/transforms"
	"container/list"

	"github.com/connectordb/duck"
)

func singleSumGenerator() (transforms.TransformFunc, error) {
	total := float64(0)
	return func(te *transforms.TransformEnvironment) *transforms.TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		val, ok := te.GetFloat()
		if !ok {
			return te.SetErrorString("sum cannot convert datapoint to number")
		}

		total += val

		return te.Copy().SetData(total)

	}, nil
}

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
			return singleSumGenerator()
		}

		argval, ok := args[0].PrimitiveValue()
		if !ok || argval == nil {
			return transforms.Err("sum requires a constant argument.")
		}

		num, ok := duck.Int(argval)
		if !ok {
			return transforms.Err("The argument to sum must be an integer")
		}

		if num <= 1 || num > 1000 {
			return transforms.Err("sum must be called with 1000 >= arg > 1")
		}

		cursum := float64(0)
		//The linked list of the last num datapoints
		dplist := list.New()

		return func(te *transforms.TransformEnvironment) *transforms.TransformEnvironment {
			if !te.CanProcess() {
				return te
			}

			val, ok := te.GetFloat()
			if !ok {
				return te.SetErrorString("sum could not convert datapoint to number")
			}

			cursum += val
			dplist.PushFront(val)

			if dplist.Len() > int(num) {
				elem := dplist.Back()
				cursum -= elem.Value.(float64)
				dplist.Remove(elem)
			}

			return te.Copy().SetData(cursum)
		}, nil
	},
}
