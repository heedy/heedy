package interpolators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAfterInterpolator(t *testing.T) {
	ai, err := NewAfterInterpolator(ds, 0, 0, "", 1.)
	require.NoError(t, err)

	//Make sure it started at right time
	dp, err := ai.Next(0.5)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[1].String())

	//Make sure that it keeps the same value for multiple nexts
	dp, err = ai.Next(0.7)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[1].String())

	//Make sure it does not see = as after
	dp, err = ai.Next(2.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[2].String())

	//Make sure it sees first of multiple times
	dp, err = ai.Next(5.5)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[5].String())

	//Make sure it passes over multiple time datapoints
	dp, err = ai.Next(6.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[7].String())

	//Make sure it ends with nil
	dp, err = ai.Next(8.0)
	require.NoError(t, err)
	require.Nil(t, dp)
}
