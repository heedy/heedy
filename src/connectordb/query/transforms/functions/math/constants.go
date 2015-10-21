package math

import (
	"connectordb/query/transforms"
	"math"
)

func constantMathTransformGenerator(value float64) transforms.TransformGenerator {

	return func(name string, args ...transforms.TransformFunc) (transforms.TransformFunc, error) {
		if len(args) != 0 {
			return transforms.Err(name + " requires zero arguments.")
		}

		return func(te *transforms.TransformEnvironment) *transforms.TransformEnvironment {
			if !te.CanProcess() {
				return te
			}

			return te.Copy().SetData(value)
		}, nil

	}
}

var pi = transforms.Transform{
	Name:         "math.pi",
	Description:  "Returns the constant pi.",
	InputSchema:  ``,
	OutputSchema: `{"type":"number"}`,
	Args:         []transforms.TransformArg{},
	Generator:    constantMathTransformGenerator(math.Pi)}

var e = transforms.Transform{
	Name:         "math.e",
	Description:  "Returns the constant e.",
	InputSchema:  ``,
	OutputSchema: `{"type":"number"}`,
	Args:         []transforms.TransformArg{},
	Generator:    constantMathTransformGenerator(math.E)}

func init() {
	pi.Register()
	e.Register()
}
