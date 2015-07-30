package interpolators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClosestInterpolator(t *testing.T) {
	ai, err := NewClosestInterpolator(ds, 0, 0, "", 2.5)
	require.NoError(t, err)
	defer ai.Close()

	//Make sure it gives the first datapoint in range
	dp, err := ai.Next(0.5)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[1].String())

	//Make sure it gives closest value less
	dp, err = ai.Next(2.1)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[1].String())

	//Make sure that it gives smaller if equal dist
	dp, err = ai.Next(2.5)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[1].String())

	//Make sure that it gives greater if larger distance
	dp, err = ai.Next(2.6)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[2].String())

	//Make sure it can iterate through many
	dp, err = ai.Next(5.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[4].String())

	//Make sure it shows first of 2 when less
	dp, err = ai.Next(5.9)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[5].String())

	//Make sure it shows second of 2 when equal
	dp, err = ai.Next(6.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[6].String())

	//Make sure it ends by keeping oldest datapoint
	dp, err = ai.Next(8.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[8].String())

	dp, err = ai.Next(20.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[8].String())

	dp, err = ai.Next(30.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[8].String())

}
