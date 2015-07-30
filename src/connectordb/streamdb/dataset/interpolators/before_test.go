package interpolators

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeforeInterpolator(t *testing.T) {
	ai, err := NewBeforeInterpolator(ds, 0, 0, "", 2.5)
	require.NoError(t, err)
	defer ai.Close()

	//Make sure it started at right time
	dp, err := ai.Next(0.5)
	require.NoError(t, err)
	require.Nil(t, dp)

	//Make sure that it moves back at least one index
	dp, err = ai.Next(2.5)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[1].String())

	//Make sure it can iterate through many
	dp, err = ai.Next(5.0)
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa[4].String())

	//Make sure it shows second of 2
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
