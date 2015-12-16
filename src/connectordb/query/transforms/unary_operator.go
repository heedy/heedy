/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package transforms

type UnaryOperator func(float64) (float64, error)

// UnaryOperatorGenerator creates a generator that gets one value or the value passed through
func UnaryOperatorGenerator(operator UnaryOperator) TransformGenerator {

	return func(name string, args ...TransformFunc) (TransformFunc, error) {

		switch len(args) {
		default:
			return Err(name + " requires zero or one arguments.")
		case 0:
			// The identity function passes on what was passed in, this way
			// we don't have to write a seperate wrapper that handles zero args.
			return unaryOperatorValueWrapper(name, PipelineGeneratorIdentity(), operator), nil
		case 1:
			return unaryOperatorValueWrapper(name, args[0], operator), nil
		}
	}
}

// This is used internally to create transform functions for various unary
// operations like uminus, sin, cos, tan...
func unaryOperatorValueWrapper(name string, transform TransformFunc, operator UnaryOperator) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		// Get the value as a float
		value, ok := te.Copy().Apply(transform).GetFloat()
		if !ok {
			return te.SetErrorString(name + " could not convert datapoint to number")
		}

		// Call the given math function
		result, err := operator(value)

		if err != nil {
			return te.SetError(err)
		}

		return te.Copy().SetData(result)
	}
}
