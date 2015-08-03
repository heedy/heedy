package transforms

import (
	. "connectordb/streamdb/datastream"

	"testing"
)

func TestComparison(t *testing.T) {
	transformTester(t, []testcase{
		{[]string{"20.0"}, "lt", false, false, Datapoint{Data: 15}, Datapoint{Data: true}},
		{[]string{"20.0"}, "lt", false, false, Datapoint{Data: 25}, Datapoint{Data: false}},
		{[]string{"test"}, "lt", true, false, Datapoint{Data: 25}, Datapoint{Data: false}},
		{[]string{"20.0"}, "lt", false, false, Datapoint{Data: 20}, Datapoint{Data: false}},
		{[]string{"20.0"}, "lte", false, false, Datapoint{Data: 20}, Datapoint{Data: true}},
		{[]string{"20.0"}, "gt", false, false, Datapoint{Data: 20}, Datapoint{Data: false}},
		{[]string{"20.0"}, "gte", false, false, Datapoint{Data: 20}, Datapoint{Data: true}},
		{[]string{"20.0"}, "gt", false, false, Datapoint{Data: 15}, Datapoint{Data: false}},
		{[]string{"20.0"}, "gt", false, false, Datapoint{Data: 25}, Datapoint{Data: true}},
		{[]string{"20.0"}, "lt", false, true, Datapoint{Data: "hi"}, Datapoint{Data: true}},
		{[]string{"try", "20.0"}, "lt", false, true, Datapoint{Data: 15}, Datapoint{Data: true}},
		{[]string{"test", "20.0"}, "lt", false, false, Datapoint{Data: map[string]interface{}{"test": 15}}, Datapoint{Data: true}},
		{[]string{"test", "20.0"}, "lt", false, false, Datapoint{Data: map[string]interface{}{"test": 25}}, Datapoint{Data: false}},
		{[]string{"nexist", "20.0"}, "lt", false, true, Datapoint{Data: map[string]interface{}{"test": 15}}, Datapoint{Data: false}},
		{[]string{}, "lt", true, false, Datapoint{Data: 15}, Datapoint{Data: true}},
		{[]string{"20.0", "hi", "hi2"}, "lt", true, false, Datapoint{Data: 15}, Datapoint{Data: true}},
	})
}

func TestEq(t *testing.T) {
	transformTester(t, []testcase{
		{[]string{"20.0"}, "eq", false, false, Datapoint{Data: 20}, Datapoint{Data: true}},
		{[]string{"20.0"}, "eq", false, false, Datapoint{Data: 25}, Datapoint{Data: false}},
		{[]string{"20.0"}, "eq", false, false, Datapoint{Data: "20.0"}, Datapoint{Data: true}},
		{[]string{"test"}, "eq", false, false, Datapoint{Data: "test"}, Datapoint{Data: true}},
		{[]string{"test"}, "eq", false, false, Datapoint{Data: "bad"}, Datapoint{Data: false}},
		{[]string{"try", "20.0"}, "eq", false, true, Datapoint{Data: 15}, Datapoint{Data: true}},
		{[]string{"test", "20.0"}, "eq", false, false, Datapoint{Data: map[string]interface{}{"test": 20}}, Datapoint{Data: true}},
		{[]string{"test", "20.0"}, "eq", false, false, Datapoint{Data: map[string]interface{}{"test": 25}}, Datapoint{Data: false}},
		{[]string{"nexist", "20.0"}, "eq", false, true, Datapoint{Data: map[string]interface{}{"test": 15}}, Datapoint{Data: false}},
		{[]string{}, "eq", true, false, Datapoint{Data: 15}, Datapoint{Data: true}},
		{[]string{"20.0", "hi", "hi2"}, "eq", true, false, Datapoint{Data: 15}, Datapoint{Data: true}},
	})
}
