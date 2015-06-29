package datastream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatapointArrayRange(t *testing.T) {
	da := NewDatapointArrayRange(dpa7)
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

func TestTimeRange(t *testing.T) {
	da := NewDatapointArrayRange(dpa7)

	tr := NewTimeRange(da, 3, 6)

	dp, err := tr.Next()
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

	tr = NewTimeRange(da, 3, 6)
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
	da := NewDatapointArrayRange(dpa7)

	tr := NewNumRange(da, 5)

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

	dp, err := er.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dpa, err := er.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)

	er.Close()
}
