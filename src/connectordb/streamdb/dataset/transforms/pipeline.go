package transforms

import "connectordb/streamdb/datastream"

func pipelineGeneratorTransform(left, right TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		leftResult, err := left(dp)
		if err != nil || leftResult == nil {
			return nil, err
		}

		// pass the data through the pipeline to do a transform
		rightResult, err := right(leftResult)
		if err != nil || rightResult == nil {
			return nil, err
		}

		return rightResult, nil
	}
}

func pipelineGeneratorIf(child TransformFunc) TransformFunc {
	return func(dp *datastream.Datapoint) (*datastream.Datapoint, error) {
		if dp == nil {
			return nil, nil
		}

		passOn, err := readBool("if", dp, child)
		if err != nil {
			return nil, err
		}

		if passOn == true {
			return dp, nil
		}

		return nil, nil
	}
}
