package streamdb

import (
	"encoding/json"
	"io"
	"streamdb/schema"
	"streamdb/timebatchdb"
	"testing"
)

func TestRangeReader(t *testing.T) {
	timestamps := []int64{1000, 1500, 2001, 2500, 3000}

	dpschema, err := schema.NewSchema(`{"type": "integer"}`)
	if err != nil {
		t.Errorf("Failed to create schema: %v", err)
		return
	}

	dpb := make([][]byte, 5)

	for i := 0; i < 5; i++ {
		dpb[i], err = dpschema.Marshal(i)
		if err != nil {
			t.Errorf("Failed to create data point: %v", err)
			return
		}
	}

	dpa := timebatchdb.CreateDatapointArray(timestamps, dpb, "hello/world")

	rr := NewRangeReader(dpa, dpschema, "user1/device1/stream1")

	dp, err := rr.Next()
	if err != nil || dp.Data.(float64) != 0. || dp.IntTimestamp() != 1000 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}
	dp, err = rr.Next()
	if err != nil || dp.Data.(float64) != 1. || dp.IntTimestamp() != 1500 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}
	dp, err = rr.Next()
	if err != nil || dp.Data.(float64) != 2. || dp.IntTimestamp() != 2001 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}
	dp, err = rr.Next()
	if err != nil || dp.Data.(float64) != 3. || dp.IntTimestamp() != 2500 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}
	dp, err = rr.Next()
	if err != nil || dp.Data.(float64) != 4. || dp.IntTimestamp() != 3000 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}

	dp, err = rr.Next()
	if err != nil || dp != nil {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}
	rr.Close()

}

func TestJsonReader(t *testing.T) {
	timestamps := []int64{1000, 1500, 2001, 2500, 3000}

	dpschema, err := schema.NewSchema(`{"type": "integer"}`)
	if err != nil {
		t.Errorf("Failed to create schema: %v", err)
		return
	}

	dpb := make([][]byte, 5)

	for i := 0; i < 5; i++ {
		dpb[i], err = dpschema.Marshal(i)
		if err != nil {
			t.Errorf("Failed to create data point: %v", err)
			return
		}
	}

	dpa := timebatchdb.CreateDatapointArray(timestamps, dpb, "hello/world")

	jr, err := NewJsonReader(NewRangeReader(dpa, dpschema, "user1/device1/stream1"))

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

	var arr *[]Datapoint
	err = json.Unmarshal(databytes, &arr)
	if err != nil {
		t.Errorf("Failed to unmarshal: %s", string(databytes))
		return
	}

	if len(*arr) != 5 {
		t.Errorf("Incorrect length: %v", len(*arr))
		return
	}

	if dp := (*arr)[0]; dp.Data.(float64) != 0. || dp.IntTimestamp() != 1000 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}

	if dp := (*arr)[1]; dp.Data.(float64) != 1. || dp.IntTimestamp() != 1500 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}

	if dp := (*arr)[2]; dp.Data.(float64) != 2. || dp.IntTimestamp() != 2001 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}

	if dp := (*arr)[3]; dp.Data.(float64) != 3. || dp.IntTimestamp() != 2500 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}

	if dp := (*arr)[4]; dp.Data.(float64) != 4. || dp.IntTimestamp() != 3000 || dp.Sender != "hello/world" || dp.Stream != "user1/device1/stream1" {
		t.Errorf("Incorrect read: %v %s", err, dp)
		return
	}

}
