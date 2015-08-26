package math

import (
	"connectordb/streamdb/query/transforms"
	"math"
)

/* Math provides general mathematical functions and constants under the math
package.
*/

type mathOperation func(float64) float64

func mathTransformGenerator(mathFunc mathOperation) transforms.TransformGenerator {

	return func(name string, args ...transforms.TransformFunc) (transforms.TransformFunc, error) {
		if len(args) != 1 {
			return transforms.Err(name + " requires one argument.")
		}

		return func(te *transforms.TransformEnvironment) *transforms.TransformEnvironment {
			if !te.CanProcess() {
				return te
			}

			te = args[0](te)

			val, ok := te.GetFloat()
			if !ok {
				return te.SetErrorString(name + " could not convert datapoint to number")
			}

			result := mathFunc(val)

			return te.Copy().SetData(result)
		}, nil

	}
}

var floor = transforms.Transform{
	Name:         "math.floor",
	Description:  "Returns the floor of the given argument. If no argument given, returns the floor of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the floor of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: mathTransformGenerator(math.Floor)}

var abs = transforms.Transform{
	Name:         "math.abs",
	Description:  "Returns the absolute value of the given argument. If no argument given, returns the absolute value of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the absolute value of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: mathTransformGenerator(math.Abs)}

var ceil = transforms.Transform{
	Name:         "math.ceil",
	Description:  "Returns the ceiling of the given argument. If no argument given, returns the ceiling value of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the ceiling value of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: mathTransformGenerator(math.Ceil)}

var sqrt = transforms.Transform{
	Name:         "math.sqrt",
	Description:  "Returns the square root of the given argument. If no argument given, returns the square root of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the square root of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: mathTransformGenerator(math.Sin)}

var ln = transforms.Transform{
	Name:         "math.ln",
	Description:  "Returns the log base 2 of the given argument. If no argument given, returns the log base 2 of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the log base 2 of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: mathTransformGenerator(math.Log2)}

func init() {
	ln.Register()
	sqrt.Register()
	ceil.Register()
	abs.Register()
	floor.Register()
}
