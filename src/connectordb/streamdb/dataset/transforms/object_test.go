package transforms

import (
	. "connectordb/streamdb/datastream"
	"testing"
)

func TestHasGetTransform(t *testing.T) {
	statelessTransformTester(t, []statelesstestcase{
		{[]string{"try", "b"}, "has", true, false, Datapoint{Data: 15}, nil},
		{[]string{}, "has", true, false, Datapoint{Data: 15}, nil},
		{[]string{"try", "b"}, "get", true, false, Datapoint{Data: 15}, nil},
		{[]string{}, "get", true, false, Datapoint{Data: 15}, nil},
		{[]string{"try", "b"}, "ifhas", true, false, Datapoint{Data: 15}, nil},
		{[]string{}, "ifhas", true, false, Datapoint{Data: 15}, nil},
		{[]string{"test"}, "get", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, &Datapoint{Data: 15}},
		{[]string{"test"}, "has", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, &Datapoint{Data: true}},
		{[]string{"test"}, "ifhas", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, &Datapoint{Data: map[string]interface{}{"test": 15}}},
		{[]string{"tes"}, "get", false, true, Datapoint{Data: map[string]interface{}{"test": 15}}, &Datapoint{Data: 15}},
		{[]string{"tes"}, "has", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, &Datapoint{Data: false}},
		{[]string{"tes"}, "ifhas", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, nil},
	})
}
