/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package datapoint

import (
	"connectordb/datastream"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonReader(t *testing.T) {
	timestamps := []float64{1000, 1500, 2001, 2500, 3000}

	dpb := make([]datastream.Datapoint, 5)
	var dpc datastream.Datapoint

	for i := 0; i < 5; i++ {
		dpb[i] = datastream.Datapoint{Timestamp: timestamps[i], Data: float64(i), Sender: "hello/world"}
	}

	dpa := datastream.NewDatapointArrayRange(dpb, 0)

	jr, err := NewJsonReader(dpa, "", "\n", "")
	require.NoError(t, err)

	dec := json.NewDecoder(jr)
	for i := 0; i < 5; i++ {
		require.NoError(t, dec.Decode(&dpc))
		require.True(t, dpb[i].IsEqual(dpc))
	}
	jr.Close()

}
