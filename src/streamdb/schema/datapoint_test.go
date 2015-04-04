package schema

import (
	"testing"
)

func TestDatapoint(t *testing.T) {
	s_string, err := NewSchema(`{"type": "string"}`)
	if err != nil {
		t.Errorf("Failed to create schema : %s", err)
		return
	}

	dp := NewDatapoint(s_string)

	dp.Timestamp = 13.234
	dp.Data = "Hello!"

	if dp.IntTimestamp() != 13234000000 {
		t.Errorf("Timestamp conversion faield: %v", dp.IntTimestamp())
		return
	}

	val, err := dp.DataBytes()
	if err != nil {
		t.Errorf("Failed to get byte array: %s", err)
		return
	}

	dp2, err := LoadDatapoint(s_string, dp.IntTimestamp(), val, "sender", "stream")
	if err != nil {
		t.Errorf("Failed to load datapoint: %s", err)
		return
	}

	if dp2.Data.(string) != "Hello!" || dp2.Timestamp != 13.234 || dp2.Sender != "sender" || dp2.Stream != "stream" {
		t.Errorf("Datapoint loaded incorrectly: %v", dp2)
		return
	}
}
