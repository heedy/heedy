package functions

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/query/transforms"
	"container/list"
	"errors"

	"github.com/connectordb/duck"
)

var average = transforms.Transform{
	Name:         "average",
	Description:  "Given a datapoint number to average over, returns the average of the last number of datapoints. If given no arguments, averages over entire dataset.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The number of datapoints backwards from the current datapoint to average over.",
			Constant:    true,
			Optional:    true,
		},
	},
	Generator: func(name string, args ...transforms.TransformFunc) (transforms.TransformFunc, error) {
		if len(args) > 1 {
			return transforms.Err("average must have at most one argument")
		}

		//If there are no args, we have a simplified world
		if len(args) == 0 {
			dpnum := int64(0)
			cursum := float64(0)
			return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
				if dp == nil {
					return nil, nil
				}

				val, ok := duck.Float(dp.Data)
				if !ok {
					return nil, errors.New("average could not convert datapoint to number")
				}

				dpnum++
				cursum += val
				dp.Data = cursum / float64(dpnum)
				return dp, nil
			}, nil
		}

		//Set up a linked list of the datapoints within the wanted period
		//The # datapoints must be a constant - if it is a constant, can pull
		//it in now with a nil arg
		argval, err := args[0](nil)
		if err != nil || argval == nil {
			return transforms.Err("average requires a constant argument.")
		}
		num, ok := duck.Int(argval.Data)
		if !ok {
			return transforms.Err("The argument to average must be an integer")
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
				return nil, errors.New("average could not convert datapoint to number")
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
