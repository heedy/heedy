package math

import (
	"connectordb/streamdb/query/transforms"
	"math"
)

/* Math provides general mathematical functions and constants under the math
package.
*/

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
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Floor(value), nil
	})}

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
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Abs(value), nil
	})}

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
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Ceil(value), nil
	})}

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
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Sqrt(value), nil
	})}

var ln = transforms.Transform{
	Name:         "math.log2",
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
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Log2(value), nil
	})}

func init() {
	ln.Register()
	sqrt.Register()
	ceil.Register()
	abs.Register()
	floor.Register()
}
