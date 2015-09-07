package functions

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/query/transforms"
	"testing"
)

func TestAverage(t *testing.T) {
	TestCase{
		Name:     "average",
		Args:     []transforms.TransformFunc{transforms.ConstantValueGenerator(2, nil)},
		HasError: false,
		Tests: []TestCaseElement{
			TestCaseElement{&datastream.Datapoint{Data: 3}, &datastream.Datapoint{Data: 3}, false, "first datapoint is copy"},
			TestCaseElement{&datastream.Datapoint{Data: 5}, &datastream.Datapoint{Data: 4}, false, "second is average"},
			TestCaseElement{&datastream.Datapoint{Data: 7}, &datastream.Datapoint{Data: 6}, false, "avg of 2"},
			TestCaseElement{nil, nil, false, "nil passthru"},
			TestCaseElement{&datastream.Datapoint{Data: 0}, &datastream.Datapoint{Data: 3.5}, false, "avg of 2 after nil"},
		},
	}.Run(t)
}

func TestFullAverage(t *testing.T) {
	TestCase{
		Name:     "average",
		HasError: false,
		Tests: []TestCaseElement{
			TestCaseElement{&datastream.Datapoint{Data: 3}, &datastream.Datapoint{Data: 3}, false, "first datapoint is copy"},
			TestCaseElement{&datastream.Datapoint{Data: 5}, &datastream.Datapoint{Data: 4}, false, "second is average"},
			TestCaseElement{&datastream.Datapoint{Data: 7}, &datastream.Datapoint{Data: 5}, false, "avg of all"},
			TestCaseElement{nil, nil, false, "nil passthru"},
			TestCaseElement{&datastream.Datapoint{Data: 1}, &datastream.Datapoint{Data: 4}, false, "avg of all after nil"},
		},
	}.Run(t)
}
