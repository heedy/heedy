package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatapointArrayIterator(t *testing.T) {
	da := NewDatapointArrayIterator(dpa7)
	defer da.Close()

	d, err := da.Next()
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 1.0, d.Timestamp)
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

func TestNumIterator(t *testing.T) {
	da := NewDatapointArrayIterator(dpa7)

	tr := NewNumIterator(da, 5)

	dp, err := tr.Next()
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

	defer tr.Close()
}

func TestArrayFromIterator(t *testing.T) {
	da := NewDatapointArrayIterator(dpa7)
	d2, err := NewArrayFromIterator(da)
	require.NoError(t, err)
	require.True(t, d2.IsEqual(dpa7))
}
