package transforms

import "connectordb/streamdb/datastream"

// Does a logical or on the pipeline
func pipelineGeneratorOr(left TransformFunc, right TransformFunc) TransformFunc {

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		result := CopyDatapoint(dp)

		for _, transform := range []TransformFunc{left, right} {

			filter, err := readBool("or", dp, transform)
			if err != nil {
				return nil, err
			}

			if filter {
				result.Data = true
				return result, nil
			}
		}

		result.Data = false
		return result, nil
	}

}

// Does a logical or on the pipeline
func pipelineGeneratorAnd(left TransformFunc, right TransformFunc) TransformFunc {

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		// Process the left data
		leftRes, err := readBool("and", dp, left)
		if err != nil {
			return nil, err
		}

		// Process the right data
		rightRes, err := readBool("and", dp, right)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = leftRes && rightRes
		return result, nil
	}

}

// Does a logical or on the pipeline
func pipelineGeneratorNot(transform TransformFunc) TransformFunc {

	return func(dp *datastream.Datapoint) (tdp *datastream.Datapoint, err error) {
		if dp == nil {
			return nil, nil
		}

		notResult, err := readBool("not", dp, transform)
		if err != nil {
			return nil, err
		}

		result := CopyDatapoint(dp)
		result.Data = !notResult
		return result, nil
	}

}
