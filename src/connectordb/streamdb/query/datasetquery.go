package query

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/query/interpolators"
	"connectordb/streamdb/query/transforms"
	"errors"
	"fmt"
)

var (
	//DatasetIndexBacktrack is the number of elements to backtrack in a dataset
	//element before the given time range, in order to minimize the number of starting nils
	//in the first elements of an interpolated dataset
	DatasetIndexBacktrack = int64(10)

	//TDatasetMaxSize is the maximum number of elements that a dataset can allow
	TDatasetMaxSize = 10000000
	//TDatasetMinDt is the minimum time delta that a dataset allows
	TDatasetMinDt = float64(1e-3)
)

//DatasetQueryElement specifies the information necessary to generate a single column of a Dataset
type DatasetQueryElement struct {
	Stream        string         `json:"stream,omitempty"`       //The stream name to use in the dataset if merge is off
	Transform     string         `json:"transform,omitempty"`    //The transform to use on the stream if merge is off
	Merge         []*StreamQuery `json:"merge,omitempty"`        //The DatasetElement can also be a merge operation - so we allow that too
	Interpolator  string         `json:"interpolator,omitempty"` //The interpolator to use for the element
	PostTransform string         `json:"itransform,omitempty"`   //The transform to run on the interpolated data. ifs don't filter but give nils
	AllowNil      bool           `json:"allownil,omitempty"`     //Whether or not a nil value is accepted, or whether it disqualifies the row
}

//Get is given the start time for the dataset, and returns the DatasetRangeElement associated
//with the QueryElement. It internally performs all necessary validation
func (dqe *DatasetQueryElement) Get(o Operator, tstart float64) (dre *DatasetRangeElement, err error) {
	var dr datastream.DataRange
	var tr transforms.DatapointTransform

	if dqe.PostTransform != "" {
		tr, err = transforms.NewTransformPipeline(dqe.PostTransform)
		if err != nil {
			return nil, err
		}
	}

	//First, we create the DataRange from the query - either a merge or straight query
	if dqe.Stream != "" || dqe.Transform != "" {
		//The transform is a simple stream - no merge
		if len(dqe.Merge) > 0 {
			return nil, errors.New("Dataset element cannot have both a merge and a stream/transform.")
		}
		dr, err = o.GetShiftedStreamTimeRange(dqe.Stream, tstart, 0, -DatasetIndexBacktrack, 0, dqe.Transform)
	} else {
		//The dataset is a merge
		if len(dqe.Merge) == 0 {
			return nil, errors.New("No stream(s) were selected for dataset element")
		}

		//First off, we set the start time of all the merge elements
		for i := range dqe.Merge {
			if !dqe.Merge[i].IsValid() || dqe.Merge[i].HasRange() {
				return nil, errors.New("Dataset merge array element must have only a stream and an optional transform")
			}
			dqe.Merge[i].T1 = tstart
			dqe.Merge[i].indexbacktrack = DatasetIndexBacktrack
		}
		dr, err = Merge(o, dqe.Merge)
	}
	if err != nil {
		return nil, err
	}

	//The element's datarange is ready - set up the interpolator

	intpltr, err := interpolators.Get(dr, dqe.Interpolator)
	if err != nil {
		dr.Close()
		return nil, err
	}

	return &DatasetRangeElement{
		Interpolator: intpltr,
		Transform:    tr,
		AllowNil:     dqe.AllowNil,
	}, nil
}

//DatasetQuery represents the full dataset generation query, used both for Ydatasets and Tdatasets
type DatasetQuery struct {
	StreamQuery                                   //This is used for Ydatasets - setting the Stream variable will make it a Ydataset - it also holds the range
	Merge         []*StreamQuery                  `json:"merge,omitempty"`      //optional merge for Ydatasets
	Dt            float64                         `json:"dt,omitempty"`         //Used for TDatasets - setting this variable makes it a time based query
	Dataset       map[string]*DatasetQueryElement `json:"dataset"`              //The dataset to generate
	PostTransform string                          `json:"itransform,omitempty"` //The transform to run on the full datapoint after the dataset element is created
}

//GetDatasetElements returns the DatasetRangeElement map which is used to generate the dataset
func (d *DatasetQuery) GetDatasetElements(o Operator, tstart float64) (map[string]*DatasetRangeElement, error) {
	if len(d.Dataset) == 0 {
		return nil, errors.New("The dataset query must have a dataset!")
	}
	res := make(map[string]*DatasetRangeElement)

	for key := range d.Dataset {
		de, err := d.Dataset[key].Get(o, tstart)
		if err != nil {
			for k := range res {
				res[k].Close()
			}
			return nil, err
		}
		res[key] = de
	}

	return res, nil
}

//GetYRange gets the DataRange of the Y query stream
func (d *DatasetQuery) GetYRange(o Operator) (dr datastream.DataRange, err error) {
	if d.IsValid() {
		if len(d.Merge) > 0 {
			return nil, errors.New("Dataset can't be based both on a stream and on a merge!")
		}
		return d.StreamQuery.Run(o)

	}

	//It is a merge!
	return Merge(o, d.Merge)
}

//Run executes the query to get the dataset
func (d DatasetQuery) Run(o Operator) (dr datastream.DataRange, err error) {
	var posttransform transforms.DatapointTransform

	if d.PostTransform != "" {
		posttransform, err = transforms.NewTransformPipeline(d.PostTransform)
		if err != nil {
			return nil, err
		}
	}

	//first find out if we are doing a Tdataset or a ydataset
	if !d.IsValid() && len(d.Merge) == 0 {
		//It is a Tdataset - make sure that no funnybusiness is going on
		if d.T1 < 0 || d.T2 < d.T1+d.Dt || d.I1 != 0 || d.I2 != 0 {
			return nil, errors.New("Tdataset range invalid")
		}
		if d.Dt < TDatasetMinDt || (d.T2-d.T1)/d.Dt > float64(TDatasetMaxSize) {
			return nil, fmt.Errorf("To avoid abuse, Tdataset is limited to a max of %d datapoints with min dt %f", TDatasetMaxSize, TDatasetMinDt)
		}
		dsetrange, err := d.GetDatasetElements(o, d.T1)
		dr = &TDatasetRange{
			Data:    dsetrange,
			Dt:      d.Dt,
			CurTime: d.T1,
			EndTime: d.T2,
		}
		if posttransform != nil {
			dr = &TransformRange{
				Data:      dr,
				Transform: posttransform,
			}
		}
		return dr, err
	}

	//It is a ydataset!
	if d.Dt != 0 {
		return nil, errors.New("Dataset must be either time or stream based. Not both.")
	}

	_, ok := d.Dataset["y"]
	if ok {
		return nil, errors.New("The 'y' label is reserved for the query stream in Ydatasets")
	}

	dr, err = d.GetYRange(o)
	if err != nil {
		return nil, err
	}

	dp, err := dr.Next()
	if err != nil {
		dr.Close()
		return nil, err
	}
	if dp == nil {
		dr.Close()
		return nil, errors.New("There are no datapoints in the chosen Y dataset range")
	}

	//The datapoint is not nil
	dsetrange, err := d.GetDatasetElements(o, dp.Timestamp)
	if err != nil {
		dr.Close()
		return nil, err
	}
	dr = &YDatasetRange{
		Data:   dsetrange,
		YRange: dr,
		Ydp:    dp,
	}
	if posttransform != nil {
		dr = &TransformRange{
			Data:      dr,
			Transform: posttransform,
		}
	}
	return dr, nil

}
