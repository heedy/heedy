package functions

import (
	"connectordb/streamdb/datastream"
	"testing"
)

func TestSum(t *testing.T) {
	TestCase{
		Name:     "sum",
		HasError: false,
		Tests: []TestCaseElement{
			TestCaseElement{&datastream.Datapoint{Data: 1}, &datastream.Datapoint{Data: 1}, false, "first datapoint is copy"},
			TestCaseElement{&datastream.Datapoint{Data: 1}, &datastream.Datapoint{Data: 2}, false, "second is average"},
			TestCaseElement{&datastream.Datapoint{Data: -3.1}, &datastream.Datapoint{Data: -1.1}, false, "avg of 2"},
			TestCaseElement{nil, nil, false, "nil passthru"},
			TestCaseElement{&datastream.Datapoint{Data: 0}, &datastream.Datapoint{Data: -1.1}, false, "avg of 2 after nil"},
		},
	}.Run(t)
}
