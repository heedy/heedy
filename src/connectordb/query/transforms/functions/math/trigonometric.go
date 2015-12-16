/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package math

// This file supplies trigonometric functions to the math package.

import (
	"connectordb/query/transforms"
	"math"
)

var sin = transforms.Transform{
	Name:         "math.sin",
	Description:  "Returns the sine of the given argument. If no argument given, returns the sine value of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the sine value of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Sin(value), nil
	})}

var cos = transforms.Transform{
	Name:         "math.cos",
	Description:  "Returns the cosine of the given argument. If no argument given, returns the tangent cosine of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the cosine value of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Cos(value), nil
	})}

var tan = transforms.Transform{
	Name:         "math.tan",
	Description:  "Returns the tangent of the given argument. If no argument given, returns the tangent value of the data passed through.",
	InputSchema:  `{"type":"number"}`,
	OutputSchema: `{"type":"number"}`,
	Args: []transforms.TransformArg{
		transforms.TransformArg{
			Description: "The value to take the tangent value of.",
			Constant:    false,
			Optional:    false,
		},
	},
	Generator: transforms.UnaryOperatorGenerator(func(value float64) (float64, error) {
		return math.Tan(value), nil
	})}

func init() {
	sin.Register()
	cos.Register()
	tan.Register()
}
