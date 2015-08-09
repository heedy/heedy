package query

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/query/interpolators"
	"connectordb/streamdb/query/transforms"
)

//DatasetRangeElement is the element that includes the interpolator and transform for a given dataset
type DatasetRangeElement struct {
	Interpolator interpolators.Interpolator
	Transform    transforms.DatapointTransform
	AllowNil     bool
}

//Close closes the internal database connections
func (dre *DatasetRangeElement) Close() {
	dre.Interpolator.Close()
}

//GetDatasetPoint gets a point from the dataset for the given timestamp
func GetDatasetPoint(timestamp float64, ranges map[string]*DatasetRangeElement) (dpp *datastream.Datapoint, err error) {
	datamap := make(map[string]*datastream.Datapoint)
	for key := range ranges {
		datamap[key], err = ranges[key].Interpolator.Interpolate(timestamp)
		if err != nil {
			return nil, err
		}
		if ranges[key].Transform != nil {
			datamap[key], err = ranges[key].Transform.Transform(datamap[key])
			if err != nil {
				return nil, err
			}
		}
	}
	return &datastream.Datapoint{Timestamp: timestamp, Data: datamap}, nil
}

//TDatasetRange is a DataRange used for TDatasets
type TDatasetRange struct {
	Data    map[string]*DatasetRangeElement
	Dt      float64
	CurTime float64
	EndTime float64
}

//Close closes the open DataRanges
func (dr *TDatasetRange) Close() {
	for key := range dr.Data {
		dr.Data[key].Close()
	}
}

//Next gets the next datapoint from the TDatasetRange
func (dr *TDatasetRange) Next() (dp *datastream.Datapoint, err error) {
	if dr.CurTime > dr.EndTime {
		return nil, nil
	}
	dp, err = GetDatasetPoint(dr.CurTime, dr.Data)
	dr.CurTime += dr.Dt
	return dp, err
}

type YDatasetRange struct {
	Data   map[string]*DatasetRangeElement
	YRange datastream.DataRange
	Ydp    *datastream.Datapoint
}

//Close closes the open DataRanges
func (dr *YDatasetRange) Close() {
	for key := range dr.Data {
		dr.Data[key].Close()
	}
	dr.YRange.Close()
}

//Next gets the next datapoint from the TDatasetRange
func (dr *YDatasetRange) Next() (dp *datastream.Datapoint, err error) {
	if dr.Ydp == nil {
		return nil, nil
	}
	dp, err = GetDatasetPoint(dr.Ydp.Timestamp, dr.Data)
	if err != nil {
		return dp, err
	}
	//Set the "y" datapoint
	dp.Data.(map[string]*datastream.Datapoint)["y"] = dr.Ydp

	dr.Ydp, err = dr.YRange.Next()
	return dp, err
}
