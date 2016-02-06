/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package query

import (
	"connectordb/datastream"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func CompareRange(t *testing.T, dr datastream.DataRange, dpa datastream.DatapointArray) {
	for i := range dpa {
		dp, err := dr.Next()
		require.NoError(t, err, dpa[i].String())

		require.Equal(t, dp.String(), dpa[i].String())

	}
	dp, err := dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
}

//MockOperator is used to test queries
type MockOperator struct {
	Data map[string]datastream.DatapointArray
}

func (m *MockOperator) get(streampath string) (datastream.DataRange, error) {
	md, ok := m.Data[streampath]
	if !ok {
		return nil, errors.New("Could not find stream " + streampath)
	}
	return datastream.NewDatapointArrayRange(md, 0), nil
}

func (m *MockOperator) GetStreamIndexRange(streampath string, i1 int64, i2 int64, transform string) (datastream.DataRange, error) {
	return m.get(streampath)
}
func (m *MockOperator) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error) {
	return m.get(streampath)
}
func (m *MockOperator) GetShiftedStreamTimeRange(streampath string, t1 float64, t2 float64, ishift, limit int64, transform string) (datastream.DataRange, error) {
	return m.get(streampath)
}

func NewMockOperator(d map[string]datastream.DatapointArray) *MockOperator {
	return &MockOperator{Data: d}
}

func TestStreamQueryBasics(t *testing.T) {
	s := StreamQuery{}
	require.False(t, s.IsValid())

	s.Stream = "u/d/s"
	require.True(t, s.IsValid())

	require.False(t, s.HasRange())
	s.Limit = 2
	require.True(t, s.HasRange())
}

func TestStreamQueryRun(t *testing.T) {

	dpa := datastream.DatapointArray{
		datastream.Datapoint{Data: 1},
		datastream.Datapoint{Data: 10},
		datastream.Datapoint{Data: 7},
		datastream.Datapoint{Data: 1.0},
		datastream.Datapoint{Data: 3},
		datastream.Datapoint{Data: 2.0},
		datastream.Datapoint{Data: 3.14},
	}

	mq := NewMockOperator(map[string]datastream.DatapointArray{"u/d/s": dpa})

	s := StreamQuery{}
	s.Stream = "u/d/s"
	s.I1 = 5
	s.T1 = 6
	_, err := s.Run(mq)
	require.Error(t, err)
	s.T1 = 0

	dr, err := s.Run(mq)
	require.NoError(t, err)
	require.NotNil(t, dr)
	dp, err := dr.Next()
	require.NoError(t, err)
	require.Equal(t, dp.String(), dpa[0].String())

	s.I1 = 0

	dr, err = s.Run(mq)
	require.NoError(t, err)
	require.NotNil(t, dr)
	dp, err = dr.Next()
	require.NoError(t, err)
	require.Equal(t, dp.String(), dpa[0].String())

}
