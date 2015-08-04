package interpolators

import (
	"connectordb/streamdb/datastream"
	"testing"
)

func TestClosestInterpolator(t *testing.T) {
	interpolatorTester(t, "closest", datastream.NewDatapointArrayRange(dpa, 0), []string{"hi"}, true, nil)
	interpolatorTester(t, "closest", datastream.NewDatapointArrayRange(dpa, 0), []string{}, false, []testcase{
		testcase{0.5, false, &dpa[0]},
		testcase{2.1, false, &dpa[1]}, //Make sure it gives closest value less
		testcase{2.5, false, &dpa[1]}, //Make sure that it gives smaller if equal dist
		testcase{2.6, false, &dpa[2]}, //Make sure that it gives greater if larger distance
		testcase{5.0, false, &dpa[4]}, //Make sure it can iterate through many
		testcase{5.9, false, &dpa[5]}, //Make sure it shows first of 2 when less
		testcase{6.0, false, &dpa[6]}, //Make sure it shows second of 2 when equal
		testcase{8.0, false, &dpa[8]}, //Make sure it ends by keeping oldest datapoint
		testcase{20.0, false, &dpa[8]},
		testcase{30.0, false, &dpa[8]},
	})

}
