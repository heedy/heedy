package functions

import (
	"connectordb/streamdb/datastream"
	"container/list"
	"errors"

	"github.com/connectordb/duck"
)

var smooth = Transform{
	Name:         "smooth",
	Description:  "Given a datapoint number to smooth over, returns the average of the last number of datapoints",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []TransformArg{
		TransformArg{
			Description: "The number of datapoints backwards from the current datapoint to smooth over.",
			Constant:    true,
		},
	},
	Generator: func(name string, args ...TransformFunc) (TransformFunc, error) {
		if len(args) != 1 {
			return Err("smooth must have one argument")
		}

		//Set up a linked list of the datapoints within the wanted time period
		//The time period must be a constant - if it is a constant, can pull
		//it in now with a nil arg
		argval, err := args[0](nil)
		if err != nil || argval == nil {
			return Err("smooth requires a constant argument.")
		}
		num, ok := duck.Int(argval.Data)
		if !ok {
			return Err("The argument to smooth must be an integer")
		}
		if num <= 1 || num > 1000 {
			return Err("Smooth must be called with 1000 >= arg > 1")
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
				return nil, errors.New("smooth could not convert datapoint to number")
			}

			cursum += val
			dplist.PushFront(val)

			if dplist.Len() > int(num) {
				elem := dplist.Back()
				cursum -= elem.Value.(float64)
				dplist.Remove(elem)
			}

			returnval := dp.Copy()
			returnval.Data = cursum / float64(dplist.Len())
			return returnval, nil
		}, nil
	},
}
