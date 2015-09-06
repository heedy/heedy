package transforms

// This is the key pipeline structure.
func pipeline(functionPipeline []TransformFunc) TransformFunc {

	// Pass on items of length 1
	if len(functionPipeline) == 1 {
		return functionPipeline[0]
	}

	//var archievedErr *TransformEnvironment

	// The inputs to the ith function. The +1 is for the pipe output
	// imagine a pipeline like (p0) f0 (p1) f1 (p2) <end>
	//inputValues := make([]*TransformEnvironment, len(functionPipeline)+1)

	return func(te *TransformEnvironment) *TransformEnvironment {
		if !te.CanProcess() {
			return te
		}

		for _, item := range functionPipeline {
			te = te.Apply(item)
		}

		return te
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
