package transforms

/*
type binaryTransformGen func(left, right TransformFunc) TransformFunc

type binaryTransformOper func(left, right interface{}) (value interface{}, errorString string)

func binaryTransformWrapper(left, right TransformFunc)

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
*/
