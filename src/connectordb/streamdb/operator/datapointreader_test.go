package operator

import (
	"connectordb/streamdb/datastream"
	"encoding/json"
	"io"
	"testing"
)

func TestJsonReader(t *testing.T) {
	timestamps := []float64{1000, 1500, 2001, 2500, 3000}

	dpb := make([]datastream.Datapoint, 5)

	for i := 0; i < 5; i++ {
		dpb[i] = datastream.Datapoint{Timestamp: timestamps[i], Data: i, Sender: "hello/world"}
	}

	dpa := datastream.NewDatapointArrayRange(dpb)

	jr, err := NewJsonReader(dpa)

	databytes := make([]byte, 5000)

	i, err := jr.Read(databytes[:5])
	if i != 5 || err != nil {
		t.Errorf("Incorrect read: %v %v", err, i)
		return
	}
	i, err = jr.Read(databytes[5:20])
	if i != 15 || err != nil {
		t.Errorf("Incorrect read: %v %v", err, i)
		return
	}
	i, err = jr.Read(databytes[20:])
	if i <= 0 || err != io.EOF {
		t.Errorf("Incorrect read: %v %v", err, i)
		return
	}
	jr.Close()

	databytes = databytes[:20+i]

	var arr *[]datastream.Datapoint
	err = json.Unmarshal(databytes, &arr)
	if err != nil {
		t.Errorf("Failed to unmarshal: %s", string(databytes))
		return
	}

	if len(*arr) != 5 {
		t.Errorf("Incorrect length: %v", len(*arr))
		return
	}

	if dp := (*arr)[0]; dp.Data.(float64) != 0. || dp.Timestamp != 1000 || dp.Sender != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[1]; dp.Data.(float64) != 1. || dp.Timestamp != 1500 || dp.Sender != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[2]; dp.Data.(float64) != 2. || dp.Timestamp != 2001 || dp.Sender != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[3]; dp.Data.(float64) != 3. || dp.Timestamp != 2500 || dp.Sender != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[4]; dp.Data.(float64) != 4. || dp.Timestamp != 3000 || dp.Sender != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

}
