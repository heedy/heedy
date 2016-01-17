/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"
	"errors"
	"fmt"

	"github.com/connectordb/pipescript"
)

//MaxMergeNumber represents the maximum number of streams to merge. Any number greater than this will result in an error
var MaxMergeNumber = 10

//MergeRange is a DataRange that merges several DataRanges together. It is used to implement the Merge command
type MergeRange struct {
	datarange []datastream.DataRange
	iterator  pipescript.DatapointIterator
}

//Close closes the merge
func (mr *MergeRange) Close() {
	for i := range mr.datarange {
		mr.datarange[i].Close()
	}
}

//Next gets the next datapoint of the merged stream
func (mr *MergeRange) Next() (*datastream.Datapoint, error) {
	dp, err := mr.iterator.Next()
	if err != nil || dp == nil {
		return nil, err
	}
	return &datastream.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil
}

//NewMergeRange generates a MergeRange given an array of DataRanges
func NewMergeRange(dr []datastream.DataRange) (*MergeRange, error) {
	iarray := make([]pipescript.DatapointIterator, len(dr))

	for i := range dr {
		iarray[i] = &DatapointIterator{dr[i]}
	}
	mrg, err := pipescript.Merge(iarray)
	if err != nil {
		for i := range dr {
			dr[i].Close()
		}
	}

	return &MergeRange{dr, mrg}, err
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
