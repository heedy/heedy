package transforms

func pipelineGeneratorTransform(left, right TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		return te.Apply(left).Apply(right)
	}
}

func pipelineGeneratorIf(child TransformFunc) TransformFunc {
	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		passOn, ok := te.Copy().Apply(child).GetBool()
		if !ok {
			return te.SetErrorString("If value not a boolean")
		}

		if passOn {
			return te
		}

		te.Datapoint = nil
		return te
	}
}
