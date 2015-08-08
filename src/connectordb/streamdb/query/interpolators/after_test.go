package interpolators

import (
	"connectordb/streamdb/datastream"
	"testing"
)

func TestAfterInterpolator(t *testing.T) {
	interpolatorTester(t, "after", datastream.NewDatapointArrayRange(dpa, 0), []string{"hi"}, true, nil)
	interpolatorTester(t, "after", datastream.NewDatapointArrayRange(dpa, 0), []string{}, false, []testcase{
		testcase{0.5, false, &dpa[0]},
		testcase{0.7, false, &dpa[0]},
		testcase{2.0, false, &dpa[2]}, //Make sure it does not see = as after
		testcase{5.5, false, &dpa[5]}, //Make sure it sees first of multiple times
		testcase{6.0, false, &dpa[7]},
		testcase{8.0, false, nil},
	})
}
