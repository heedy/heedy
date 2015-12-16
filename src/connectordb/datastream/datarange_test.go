/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package datastream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatapointArrayRange(t *testing.T) {
	da := NewDatapointArrayRange(dpa7, 2)
	defer da.Close()

	require.EqualValues(t, da.Index(), 2)
	d, err := da.Next()
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 1.0, d.Timestamp)
	require.EqualValues(t, da.Index(), 3)
	d, err = da.Next()
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 2.0, d.Timestamp)

	dpa, err := da.NextArray()
	require.NoError(t, err)
	require.NotNil(t, dpa)
	require.True(t, dpa.IsEqual(dpa7[2:]))

	d, err = da.Next()
	require.NoError(t, err)
	require.Nil(t, d)

	dpa, err = da.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)
}

func TestTimeRange(t *testing.T) {
	da := NewDatapointArrayRange(dpa7, 1)

	tr, err := NewTimeRange(da, 3, 6)
	require.NoError(t, err)

	require.EqualValues(t, tr.Index(), 4)

	dp, err := tr.Next()
	require.EqualValues(t, tr.Index(), 5)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 4., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 5., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 6., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 6., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	tr.Close()

	tr, err = NewTimeRange(da, 3, 6)
	require.NoError(t, err)
	dpa, err := tr.NextArray()
	require.NoError(t, err)
	require.NotNil(t, dpa)
	require.True(t, dpa.IsEqual(dpa7[3:7]))

	dpa, err = tr.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)

	defer tr.Close()
}

func TestNumRange(t *testing.T) {
	da := NewDatapointArrayRange(dpa7, 1)

	tr := NewNumRange(da, 5)

	require.EqualValues(t, tr.Index(), 1)

	dp, err := tr.Next()
	require.EqualValues(t, tr.Index(), 2)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 1., dp.Timestamp)

	err = tr.Skip(2)
	require.NoError(t, err)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 4., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 5., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 6., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, 6., dp.Timestamp)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	tr.Close()

	tr = NewNumRange(da, 4)
	dpa, err := tr.NextArray()
	require.NoError(t, err)
	require.NotNil(t, dpa)
	require.True(t, dpa.IsEqual(dpa7[:4]))

	dpa, err = tr.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)

	defer tr.Close()
}

//This is just to increase test coverage...
func TestEmptyRange(t *testing.T) {
	er := EmptyRange{}
	require.EqualValues(t, er.Index(), 0)
	dp, err := er.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dpa, err := er.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)

	er.Close()
}
