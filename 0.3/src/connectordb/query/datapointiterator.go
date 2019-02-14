package query

import (
	"connectordb/datastream"

	"github.com/connectordb/pipescript"
)

// DatapointIterator converts to pipescript's datapoint
type DatapointIterator struct {
	Range datastream.DataRange
}

func (dpi *DatapointIterator) Next() (*pipescript.Datapoint, error) {
	dp, err := dpi.Range.Next()
	if err != nil || dp == nil {
		return nil, err
	}
	return &pipescript.Datapoint{Timestamp: dp.Timestamp, Data: dp.Data}, nil
}
