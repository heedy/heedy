package interpolators

import (
	"connectordb/streamdb/datastream"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var dpa = datastream.DatapointArray{
	datastream.Datapoint{1., "test0", ""},
	datastream.Datapoint{2., "test1", ""},
	datastream.Datapoint{3., "test2", ""},
	datastream.Datapoint{4., "test3", ""},
	datastream.Datapoint{5., "test4", ""},
	datastream.Datapoint{6., "test5", ""},
	datastream.Datapoint{6., "test6", ""},
	datastream.Datapoint{7., "test7", ""},
	datastream.Datapoint{8., "test8", ""},
}

type testcase struct {
	Timestamp float64
	Haserror  bool
	Result    *datastream.Datapoint
}

func interpolatorTester(t *testing.T, iname string, dr datastream.DataRange, args []string, haserr bool, testcases []testcase) {
	i, err := Interpolators[iname](dr, args)
	if haserr {
		require.Error(t, err, fmt.Sprintf("%s: %v", iname, args))
		return
	}
	require.NoError(t, err, fmt.Sprintf("%s: %v", iname, args))
	for _, c := range testcases {
		dp, err := i.Interpolate(c.Timestamp)
		if c.Haserror {
			require.Error(t, err, fmt.Sprintf("%s: %v", iname, c))
			return
		}
		require.NoError(t, err, fmt.Sprintf("%s: %v", iname, c))
		if c.Result == nil {
			require.Nil(t, dp, fmt.Sprintf("%s: %v", iname, c))
		} else {
			require.NotNil(t, dp, fmt.Sprintf("%s: %v", iname, c))
			require.Equal(t, c.Result.String(), dp.String(), fmt.Sprintf("%s: %v", iname, c))
		}
	}
	i.Close()
}
