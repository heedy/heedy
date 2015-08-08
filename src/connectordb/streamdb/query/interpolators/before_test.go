package interpolators

import (
	"connectordb/streamdb/datastream"
	"testing"
)

func TestBeforeInterpolator(t *testing.T) {
	interpolatorTester(t, "before", datastream.NewDatapointArrayRange(dpa, 0), []string{"hi"}, true, nil)
	interpolatorTester(t, "before", datastream.NewDatapointArrayRange(dpa, 0), []string{}, false, []testcase{
		testcase{0.5, false, nil},
		testcase{2.5, false, &dpa[1]},
		testcase{5.0, false, &dpa[4]},
		testcase{6.0, false, &dpa[6]},
		testcase{8.0, false, &dpa[8]},
		testcase{20.0, false, &dpa[8]},
		testcase{30.0, false, &dpa[8]},
	})
}
