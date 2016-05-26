/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"
	"errors"
	"fmt"

	"github.com/connectordb/pipescript"
	"github.com/connectordb/pipescript/interpolator"
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
	StreamQuery                 // Allows to query the stream by its own values
	Merge        []*StreamQuery `json:"merge,omitempty"`        //The DatasetElement can also be a merge operation - so we allow that too
	Interpolator string         `json:"interpolator,omitempty"` //The interpolator to use for the element
	AllowNil     bool           `json:"allownil,omitempty"`     //Whether or not a nil value is accepted, or whether it disqualifies the row
}

// Get is given the start time for the dataset, and returns the DatasetRangeElement
func (dqe *DatasetQueryElement) Get(o Operator, tstart float64) (dre *DatasetRangeElement, err error) {
	var dr datastream.DataRange

	//First, we create the DataRange from the query - either a merge or straight query
	if dqe.Stream != "" {
		//The transform is a simple stream - no merge
		if len(dqe.Merge) > 0 {
			return nil, errors.New("Dataset element cannot have both a merge and a stream")
		}

		// First check if the stream has some form of range associated with it
		if !dqe.HasRange() {
			dqe.T1 = tstart
			dqe.indexbacktrack = DatasetIndexBacktrack
		}

		dr, err = dqe.StreamQuery.Run(o)

	} else {
		//The dataset is a merge
		if dqe.Transform != "" {
			return nil, errors.New("Set transforms within each merge element instead of overall for dataset element")
		}
		if len(dqe.Merge) == 0 {
			return nil, errors.New("No stream(s) were selected for dataset element")
		}

		//First off, we set the start time of all the merge elements
		for i := range dqe.Merge {
			if !dqe.Merge[i].IsValid() {
				return nil, errors.New("Dataset merge array element invalid")
			}
			if !dqe.Merge[i].HasRange() {
				dqe.Merge[i].T1 = tstart
				dqe.Merge[i].indexbacktrack = DatasetIndexBacktrack
			}
		}
		dr, err = Merge(o, dqe.Merge)

	}
	if err != nil {
		return nil, err
	}

	//The element's datarange is ready - set up the interpolator
	intpltr, err := interpolator.Parse(dqe.Interpolator, &DatapointIterator{dr})
	if err != nil {
		dr.Close()
		return nil, err
	}

	return &DatasetRangeElement{
		Interpolator: intpltr,
		Range:        dr,
		AllowNil:     dqe.AllowNil,
	}, nil
}

//DatasetQuery represents the full dataset generation query, used both for Ydatasets and Tdatasets
type DatasetQuery struct {
	StreamQuery                                   //This is used for Ydatasets - setting the Stream variable will make it a Ydataset - it also holds the range
	Merge         []*StreamQuery                  `json:"merge,omitempty"`         //optional merge for Ydatasets
	Dt            float64                         `json:"dt,omitempty"`            //Used for TDatasets - setting this variable makes it a time based query
	Dataset       map[string]*DatasetQueryElement `json:"dataset"`                 //The dataset to generate
	PostTransform string                          `json:"posttransform,omitempty"` //The transform to run on the full datapoint after the dataset element is created
}

//GetDatasetElements returns the range element map used for generating the datasets
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

//GetXRange gets the DataRange of the X query stream
func (d *DatasetQuery) GetXRange(o Operator) (dr datastream.DataRange, err error) {
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
	var posttransform *pipescript.Script
	var iiter pipescript.DatapointIterator
	if d.PostTransform != "" {
		posttransform, err = pipescript.Parse(d.PostTransform)
		if err != nil {
			return nil, err
		}
	}

	// Get the dataset elements, and prepare the interpolator map
	dsetrange, err := d.GetDatasetElements(o, d.T1)
	if err != nil {
		return nil, err
	}
	dsetipltr := make(map[string]interpolator.InterpolatorInstance)
	for key := range dsetrange {
		dsetipltr[key] = dsetrange[key].Interpolator
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

		iiter, err = interpolator.GetTDataset(d.T1, d.T2, d.Dt, dsetipltr)
	} else {
		//It is an xdataset!
		if d.Dt != 0 {
			return nil, errors.New("Dataset must be either time or stream based. Not both.")
		}

		dr, err = d.GetXRange(o)
		if err != nil {
			return nil, err
		}

		iiter, err = interpolator.GetXDataset(&DatapointIterator{dr}, "x", dsetipltr)
	}
	if err != nil {
		return nil, err
	}

	if posttransform != nil {
		posttransform.SetInput(iiter)
		iiter = posttransform
	}
	return &DatasetRange{dsetrange, iiter}, nil

}
