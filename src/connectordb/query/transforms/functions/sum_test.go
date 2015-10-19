package functions

import (
	"connectordb/datastream"
	"connectordb/query/transforms"
	"testing"
)

func TestSum(t *testing.T) {
	TestCase{
		Name:     "sum",
		Args:     []transforms.TransformFunc{transforms.ConstantValueGenerator(2, nil)},
		HasError: false,
		Tests: []TestCaseElement{
			TestCaseElement{&datastream.Datapoint{Data: 3}, &datastream.Datapoint{Data: 3}, false, "first datapoint is copy"},
			TestCaseElement{&datastream.Datapoint{Data: 5}, &datastream.Datapoint{Data: 8}, false, "second is asum"},
			TestCaseElement{&datastream.Datapoint{Data: 7}, &datastream.Datapoint{Data: 12}, false, "sum of 2"},
			TestCaseElement{nil, nil, false, "nil passthru"},
			TestCaseElement{&datastream.Datapoint{Data: 1}, &datastream.Datapoint{Data: 8}, false, "sum of 2 after nil"},
		},
	}.Run(t)
}

func TestFullSum(t *testing.T) {
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
