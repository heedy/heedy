package transforms

import "errors"

// Adds the left and right hand side
func addTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return binaryOperatorValueWrapper("+", left, right, func(leftVal, rightVal float64) (float64, error) {
		return leftVal + rightVal, nil
	})
}

// Multiplies the left and right hand side
func multiplyTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return binaryOperatorValueWrapper("*", left, right, func(leftVal, rightVal float64) (float64, error) {
		return leftVal * rightVal, nil
	})
}

// Divides the left and right hand side
func divideTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return binaryOperatorValueWrapper("/", left, right, func(leftVal, rightVal float64) (float64, error) {
		if rightVal == 0.0 || rightVal == -0.0 {
			return 0, errors.New("Cannot divide by zero")
		}

		return leftVal / rightVal, nil
	})
}

// Subtracts the left and right hand side
func subtractTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return binaryOperatorValueWrapper("-", left, right, func(leftVal, rightVal float64) (float64, error) {
		return leftVal - rightVal, nil
	})
}

// Performs a unary minus
func inverseTransformGenerator(transform TransformFunc) TransformFunc {
	return unaryOperatorValueWrapper("negate (-)", transform, func(value float64) (float64, error) {
		return -value, nil
	})
}
