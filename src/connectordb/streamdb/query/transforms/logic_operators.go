package transforms

// Does a logical or on the pipeline
func pipelineGeneratorOr(left TransformFunc, right TransformFunc) TransformFunc {

	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		for _, transform := range []TransformFunc{left, right} {

			filter, ok := te.Copy().Apply(transform).GetBool()
			if !ok {
				return te.SetErrorString("or value not a boolean")
			}

			if filter {
				return te.SetData(true)
			}
		}

		return te.SetData(false)
	}

}

// Does a logical or on the pipeline
func pipelineGeneratorAnd(left TransformFunc, right TransformFunc) TransformFunc {

	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		// Process the left data
		leftRes, ok := te.Copy().Apply(left).GetBool()
		if !ok {
			return te.SetErrorString("and value not a boolean")
		}

		// Process the right data
		rightRes, ok := te.Copy().Apply(right).GetBool()
		if !ok {
			return te.SetErrorString("and value not a boolean")
		}

		return te.SetData(leftRes && rightRes)
	}

}

// Does a logical or on the pipeline
func pipelineGeneratorNot(transform TransformFunc) TransformFunc {

	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		// Process the left data
		notResult, ok := te.Copy().Apply(transform).GetBool()
		if !ok {
			return te.SetErrorString("not value not a boolean")
		}

		return te.SetData(!notResult)
	}
}
