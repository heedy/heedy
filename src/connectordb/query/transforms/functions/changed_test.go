package functions

import (
	"connectordb/datastream"
	"testing"
)

func TestChanged(t *testing.T) {
	TestCase{
		Name:     "changed",
		HasError: false,
		Tests: []TestCaseElement{
			TestCaseElement{&datastream.Datapoint{Data: 3}, &datastream.Datapoint{Data: true}, false, "first datapoint is yes"},
			TestCaseElement{&datastream.Datapoint{Data: 3}, &datastream.Datapoint{Data: false}, false, "second is same"},
			TestCaseElement{&datastream.Datapoint{Data: 7}, &datastream.Datapoint{Data: true}, false, "changed"},
			TestCaseElement{nil, nil, false, "nil passthru"},
			TestCaseElement{&datastream.Datapoint{Data: 7}, &datastream.Datapoint{Data: false}, false, "no change after passthru"},
		},
	}.Run(t)
}
