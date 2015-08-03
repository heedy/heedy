package transforms

import (
	. "connectordb/streamdb/datastream"
	"testing"
)

func TestFilterTransform(t *testing.T) {
	statelessTransformTester(t, []statelesstestcase{
		{[]string{"20.0"}, "iflt", false, false, Datapoint{Data: 15}, &Datapoint{Data: 15}},
		{[]string{"20.0"}, "iflt", false, false, Datapoint{Data: 25}, nil},
		{[]string{"test"}, "iflt", true, false, Datapoint{Data: 25}, nil},
		{[]string{"20.0"}, "iflt", false, false, Datapoint{Data: 20}, nil},
		{[]string{"20.0"}, "iflte", false, false, Datapoint{Data: 20}, &Datapoint{Data: 20}},
		{[]string{"20.0"}, "ifgt", false, false, Datapoint{Data: 20}, nil},
		{[]string{"20.0"}, "ifgte", false, false, Datapoint{Data: 20}, &Datapoint{Data: 20}},
		{[]string{"20.0"}, "ifgt", false, false, Datapoint{Data: 15}, nil},
		{[]string{"20.0"}, "ifgt", false, false, Datapoint{Data: 25}, &Datapoint{Data: 25}},
		{[]string{"20.0"}, "iflt", false, true, Datapoint{Data: "hi"}, &Datapoint{Data: "hi"}},
		{[]string{"try", "20.0"}, "iflt", false, true, Datapoint{Data: 15}, &Datapoint{Data: 15}},
		{[]string{"test", "20.0"}, "iflt", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, &Datapoint{Data: map[string]interface{}{"test": 15}}},
		{[]string{"test", "20.0"}, "iflt", false, false, Datapoint{Data: map[string]interface{}{"test": 25}}, nil},
		{[]string{"nexist", "20.0"}, "iflt", false, true, Datapoint{Data: map[string]interface{}{"test": 15}}, nil},
		{[]string{}, "iflt", true, false, Datapoint{Data: 15}, &Datapoint{Data: 15}},
		{[]string{"20.0", "hi", "hi2"}, "iflt", true, false, Datapoint{Data: 15}, &Datapoint{Data: 15}},
	})
}
