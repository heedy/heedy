package transforms

type BinaryOperator func(float64, float64) (float64, error)

// BinaryOperatorGenerator creates a generator that gets two values, one from
// the left and one from the right, converts them to floats and calls the given
// function, setting the data to the resulting value or erroring.
func BinaryOperatorGenerator(operator BinaryOperator) TransformGenerator {

	return func(name string, args ...TransformFunc) (TransformFunc, error) {
		if len(args) != 2 {
			return Err(name + " requires two arguments.")
		}

		// Store the left and right transform functions
		left := args[0]
		right := args[1]

		// Actually create the generator
		return binaryOperatorValueWrapper(name, left, right, operator), nil
	}
}

// This is used internally to create transform functions for various binary
// operations like * / + -
func binaryOperatorValueWrapper(name string, left, right TransformFunc, operator BinaryOperator) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		// Get the left and right values as floats.
		lVal, lok := te.Copy().Apply(left).GetFloat()
		rVal, rok := te.Copy().Apply(right).GetFloat()
		if !lok || !rok {
			return te.SetErrorString(name + " could not convert datapoint to number")
		}

		// Call the given math function
		result, err := operator(lVal, rVal)

		if err != nil {
			return te.SetError(err)
		}

		return te.Copy().SetData(result)
	}
}
