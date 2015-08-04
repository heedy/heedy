package transforms

import (
	"connectordb/streamdb/dataset/pipeline"
	. "connectordb/streamdb/datastream"
	"errors"
)

type TransformPipeline struct {
	transforms []DatapointTransform
}

func (p *TransformPipeline) Transform(dp *Datapoint) (rdp *Datapoint, err error) {
	if len(p.transforms) == 0 {
		return dp, nil
	}
	if dp != nil {
		//go through the transforms one by one. If one transform returns nil, short-circuit
		//returning nil, asking for another datapoint
		for i := 0; i < len(p.transforms); i++ {
			dp, err = p.transforms[i].Transform(dp)
			if err != nil || dp == nil {
				return dp, err
			}
		}
	} else {
		//the datapoint inputted was nil - that means that the DataRange is done,
		//and it is asking for any cached datapoints. clear the transforms that pass
		//nil, and loop through the ones that don't return nil until a datapoint gets
		//through the full transform, or the full array is cleared

		for dp == nil && len(p.transforms) > 0 {
			//get rid of starting nils
			for dp == nil && len(p.transforms) > 0 {
				dp, err = p.transforms[0].Transform(dp)
				if err != nil {
					return dp, err
				}
				if dp == nil {
					p.transforms = p.transforms[1:]
					break
				}
			}

			//Try finishing the pipeline
			for i := 1; i < len(p.transforms); i++ {
				dp, err = p.transforms[i].Transform(dp)
				if err != nil {
					return dp, err
				}
				if dp == nil {
					break //damn, need to try again - the value got filtered
				}
			}
		}
	}
	return dp, err
}

func NewTransformPipeline(pipe string) (*TransformPipeline, error) {
	p, err := pipeline.ParsePipeline(pipe)
	if err != nil {
		return nil, err
	}

	transforms := make([]DatapointTransform, 0)

	for i := 0; i < len(p); i++ {
		tfnc, ok := Transforms[p[i].Symbol]
		if !ok {
			return nil, errors.New("Could not find '" + p[i].Symbol + "' transform.")
		}
		t, err := tfnc(p[i].Args)
		if err != nil {
			return nil, err
		}
		transforms = append(transforms, t)
	}

	return &TransformPipeline{transforms}, nil
}
