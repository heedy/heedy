package query

import (
	"connectordb/streamdb/datastream"
	"errors"
	"fmt"
)

//MaxMergeNumber represents the maximum number of streams to merge. Any number greater than this will result in an error
var MaxMergeNumber = 10

//MergeRange is a DataRange that merges several DataRanges together. It is used to implement the Merge command
type MergeRange struct {
	datarange []datastream.DataRange
	datapoint []*datastream.Datapoint
}

//Close closes the merge
func (mr *MergeRange) Close() {
	for i := range mr.datarange {
		mr.datarange[i].Close()
	}
}

//Next gets the next datapoint of the merged stream
func (mr *MergeRange) Next() (dp *datastream.Datapoint, err error) {
	//TODO: There are several inefficiencies in this implementation: First off, it is O(n), where
	//it can be made O(logn) by using a tree. Second, I just keep nulls in the array, which is
	//totally BS, the array could be made shorter when one range empties. But I just want to get this
	//thing working atm, so making it efficient is a task for later.
	mini := -1
	mint := float64(0)
	for i := range mr.datapoint {
		//DataRanges that are empty will be nil
		if mr.datapoint[i] != nil {
			//Get the datapoint with smallest timestamp
			if mr.datapoint[i].Timestamp < mint || mini == -1 {
				mini = i
				mint = mr.datapoint[i].Timestamp
			}
		}
	}
	if mini == -1 {
		//There are no datapoints left
		return nil, nil
	}
	dp = mr.datapoint[mini]

	mr.datapoint[mini], err = mr.datarange[mini].Next()

	return dp, err
}

//NewMergeRange generates a MergeRange given an array of DataRanges
func NewMergeRange(dr []datastream.DataRange) (*MergeRange, error) {
	dpa := make([]*datastream.Datapoint, 0, len(dr))

	for i := range dr {
		dp, err := dr[i].Next()
		if err != nil {
			//We've got an error - close all the ranges
			for j := range dr {
				dr[j].Close()
			}
			return nil, err
		}
		dpa = append(dpa, dp)
	}

	return &MergeRange{dr, dpa}, nil
}

//Merge returns a MergeRange which merges the given streams into one large stream
func Merge(qo Operator, sq []*StreamQuery) (*MergeRange, error) {
	if len(sq) > MaxMergeNumber {
		return nil, errors.New(fmt.Sprintf("Merging more than %d streams is disabled.", MaxMergeNumber))
	}

	dr := make([]datastream.DataRange, 0, len(sq))
	for i := range sq {
		d, err := sq[i].Run(qo)
		if err != nil {
			//We've got an error - close all the ranges
			for j := range dr {
				dr[j].Close()
			}
			return nil, err
		}
		dr = append(dr, d)
	}

	return NewMergeRange(dr)
}
