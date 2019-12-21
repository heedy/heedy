package timeseries

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonReader(t *testing.T) {
	timestamps := []float64{1000, 1500, 2001, 2500, 3000}

	dpb := make(DatapointArray, 5)
	var dpc *Datapoint

	for i := 0; i < 5; i++ {
		dpb[i] = &Datapoint{Timestamp: timestamps[i], Data: float64(i), Actor: "hello/world"}
	}

	dpa := NewDatapointArrayIterator(dpb)

	jr, err := NewJsonReader(dpa, "", "\n", "")
	require.NoError(t, err)

	dec := json.NewDecoder(jr)
	for i := 0; i < 5; i++ {
		require.NoError(t, dec.Decode(&dpc))
		require.True(t, dpb[i].IsEqual(dpc))
	}
	jr.Close()

}

func TestJsonArrayZeroRead(t *testing.T) {
	dpa := DatapointArray{}
	dpi := NewDatapointArrayIterator(dpa)
	jr, err := NewJsonArrayReader(dpi)
	require.NoError(t, err)

	databytes := make([]byte, 5000)

	n, err := jr.Read(databytes)
	require.EqualError(t, err, io.EOF.Error())
	require.Equal(t, 2, n)
	if databytes[0] != '[' || databytes[1] != ']' {
		t.Error("Zero array invalid")
	}

}
func TestJsonArrayReader(t *testing.T) {
	timestamps := []float64{1000, 1500, 2001, 2500, 3000}

	dpb := make(DatapointArray, 5)

	for i := 0; i < 5; i++ {
		dpb[i] = &Datapoint{Timestamp: timestamps[i], Data: i, Actor: "hello/world"}
	}

	dpa := NewDatapointArrayIterator(dpb)

	jr, err := NewJsonArrayReader(dpa)

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

	var arr *[]*Datapoint
	err = json.Unmarshal(databytes, &arr)
	if err != nil {
		t.Errorf("Failed to unmarshal: %s", string(databytes))
		return
	}

	if len(*arr) != 5 {
		t.Errorf("Incorrect length: %v", len(*arr))
		return
	}

	if dp := (*arr)[0]; dp.Data.(float64) != 0. || dp.Timestamp != 1000 || dp.Actor != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[1]; dp.Data.(float64) != 1. || dp.Timestamp != 1500 || dp.Actor != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[2]; dp.Data.(float64) != 2. || dp.Timestamp != 2001 || dp.Actor != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[3]; dp.Data.(float64) != 3. || dp.Timestamp != 2500 || dp.Actor != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

	if dp := (*arr)[4]; dp.Data.(float64) != 4. || dp.Timestamp != 3000 || dp.Actor != "hello/world" {
		t.Errorf("Incorrect read: %v %v", err, dp.String())
		return
	}

}
