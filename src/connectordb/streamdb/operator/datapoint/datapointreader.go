package datapoint

import (
	"encoding/json"
	"errors"
	"io"

	"connectordb/streamdb/datastream"
)

var (
	//ErrUnrecognizedType is thrown when the data type is unrecognized, and can't be marshalled
	ErrUnrecognizedType = errors.New("Data type unrecognized")
)

//JsonReader wraps a DataRange object into one that implements the io.Reader interface, allowing it to be directly
//used in places where json is wanted
type JsonReader struct {
	dreader       datastream.DataRange //The DatapointReader to read from
	currentbuffer []byte               //The buffer of the current datapoint's bytes

}

//Close shuts down the internal RangeReader
func (r *JsonReader) Close() {
	if r.dreader != nil {
		r.dreader.Close()
	}
}

//Read reads the given number of bytes from the range
// p is the buffer to read into
//
func (r *JsonReader) Read(p []byte) (n int, err error) {
	n = 0
	for len(p) > 0 {
		if len(r.currentbuffer) > 0 {
			//There is still some stuff left in the current buffer - first copy that
			i := copy(p, r.currentbuffer)
			r.currentbuffer = r.currentbuffer[i:]
			p = p[i:]
			n += i
		}

		// If datapoint reader is done, return number of bytes read and EOF.
		if r.dreader == nil {
			return n, io.EOF
		}

		if len(r.currentbuffer) == 0 {
			//The current buffer is empty - read in a new datapoint
			dp, err := r.dreader.Next()
			if err != nil {
				return n, err
			}
			// Datapoint reader is over
			if dp == nil {
				r.currentbuffer = []byte("]")
				r.dreader.Close()
				r.dreader = nil
			} else {
				v, err := json.Marshal(dp)
				if err != nil {
					return n, err
				}
				r.currentbuffer = []byte("," + string(v))
			}
		}
	}
	return n, nil
}

//NewJsonReader creates a new json reader object. Allows using a RangeReader as an io.Reader type which outputs json.
func NewJsonReader(dreader datastream.DataRange) (*JsonReader, error) {
	dp, err := dreader.Next()
	if err != nil {
		return nil, err
	}
	if dp == nil {
		return nil, io.EOF
	}
	v, err := json.Marshal(dp)
	if err != nil {
		return nil, err
	}
	return &JsonReader{dreader, []byte("[" + string(v))}, nil
}
