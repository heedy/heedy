/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package interpolators

import (
	"connectordb/datastream"
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

func interpolatorTester(t *testing.T, iname string, dr datastream.DataRange, haserr bool, testcases []testcase) {
	i, err := Get(dr, iname)
	if haserr {
		require.Error(t, err, iname)
		return
	}
	require.NoError(t, err, iname)
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

func TestGet(t *testing.T) {
	_, err := Get(nil, "doesnotexisti")
	require.Error(t, err)
	_, err = Get(nil, "multiple:interpolators")
	require.Error(t, err)

	_, err = Get(datastream.NewDatapointArrayRange(dpa, 0), "closest(arg1)")
	require.Error(t, err)

	i, err := Get(datastream.NewDatapointArrayRange(dpa, 0), "")
	require.NoError(t, err)
	_, ok := i.(*ClosestInterpolator)
	require.True(t, ok)
}
