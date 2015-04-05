package streamdb

import (
	"encoding/json"
	"errors"
	"io"
	"streamdb/schema"
	"streamdb/timebatchdb"
)

var (
	//ErrUnrecognizedType is thrown when the data type is unrecognized, and can't be marshalled
	ErrUnrecognizedType = errors.New("Data type unrecognized")
)

//DatapointReader is an iterator
type DatapointReader interface {
	Next() (*schema.Datapoint, error)
	Close()
}

//RangeReader allows to read data from a timebatchdb.DataRange
type RangeReader struct {
	drange  timebatchdb.DataRange
	dschema *schema.Schema //The schema to use when reading datapoints
	stream  string
}

//Close shuts down the underlying DataRange
func (r *RangeReader) Close() {
	r.drange.Close()
}

//Next gets the next datapoint in the range
func (r *RangeReader) Next() (dp *schema.Datapoint, err error) {
	tbdp, err := r.drange.Next()
	if err != nil || tbdp == nil {
		return nil, err
	}

	return schema.LoadDatapoint(r.dschema, tbdp.Timestamp(), tbdp.Data(), tbdp.Key(), r.stream)
}

//NewRangeReader opens a streamreader with the given
func NewRangeReader(drange timebatchdb.DataRange, schema *schema.Schema, stream string) *RangeReader {
	return &RangeReader{drange, schema, stream}
}

//JsonReader wraps a RangeReader object into one that implements the io.Reader interface, allowing it to be directly
//used in places where json is wanted
type JsonReader struct {
	dreader       DatapointReader //The DatapointReader to read from
	currentbuffer []byte          //The buffer of the current datapoint's bytes

}

//Close shuts down the internal RangeReader
func (r *JsonReader) Close() {
	if r.dreader != nil {
		r.dreader.Close()
	}
}

//Read reads the given number of bytes from the range
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
		if r.dreader == nil {
			return n, io.EOF
		}
		if len(r.currentbuffer) == 0 {
			//The current buffer is empty - read in a new datapoint
			dp, err := r.dreader.Next()
			if err != nil {
				return n, err
			}
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
func NewJsonReader(dreader DatapointReader) (*JsonReader, error) {
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
