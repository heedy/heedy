package transforms

import "errors"

func getTransformLeftRight(te *TransformEnvironment, left, right TransformFunc) (leftVal float64, rightVal float64, err error) {
	leftVal, lok := te.Copy().Apply(left).GetFloat()
	rightVal, rok := te.Copy().Apply(right).GetFloat()

	if lok == false || rok == false {
		return 0, 0, errors.New("Illegal conversion")
	}

	return leftVal, rightVal, nil
}

// Adds the left and right hand side
func addTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {

	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		leftVal, rightVal, err := getTransformLeftRight(te, left, right)
		if err != nil {
			return te.SetError(err)
		}

		return te.SetData(leftVal + rightVal)
	}
}

// Multiplies the left and right hand side
func multiplyTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		leftVal, rightVal, err := getTransformLeftRight(te, left, right)
		if err != nil {
			return te.SetError(err)
		}

		return te.SetData(leftVal * rightVal)
	}
}

// Divides the left and right hand side
func divideTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		leftVal, rightVal, err := getTransformLeftRight(te, left, right)
		if err != nil {
			return te.SetError(err)
		}

		if rightVal == 0.0 {
			return te.SetErrorString("Zero division error")
		}

		return te.SetData(leftVal / rightVal)
	}
}

// Subtracts the left and right hand side
func subtractTransformGenerator(left TransformFunc, right TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		leftVal, rightVal, err := getTransformLeftRight(te, left, right)
		if err != nil {
			return te.SetError(err)
		}

		return te.SetData(leftVal - rightVal)
	}

}

// Subtracts the left and right hand side
func inverseTransformGenerator(transform TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		val, ok := te.Copy().Apply(transform).GetFloat()
		if ok == false {
			return te.SetError(ErrNotFloat)
		}

		return te.SetData(-val)
	}
}
