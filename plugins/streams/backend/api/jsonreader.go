package api

import (
	"encoding/json"
	"io"
)

// JsonReader imitates an io.Reader interface
type JsonReader struct {
	data           DatapointIterator // The DataRange to read from
	currentbuffer  []byte            // The buffer of the current datapoint's bytes
	Separator      []byte            // The separator to use between datapoints
	Ender          []byte
	separatorIndex int // The index in the separator to use
}

// Close shuts down the internal DataRange
func (r *JsonReader) Close() {
	if r.data != nil {
		r.data.Close()
	}
}

// Read reads the given number of bytes from the DataRange, and p is the buffer to read into
func (r *JsonReader) Read(p []byte) (n int, err error) {
	n = 0
	for len(p) > 0 {
		// if we are at a positive separator index, write as much of the separator as possible
		if len(r.Separator) > r.separatorIndex {
			i := copy(p, r.Separator[r.separatorIndex:])
			p = p[i:]
			r.separatorIndex += i
			n += i

			// Since we just wrote the separator, check if we have to return
			if len(p) == 0 {
				return n, nil
			}
		}

		if len(r.currentbuffer) > 0 {
			// There is still some stuff left in the current buffer - first copy that
			i := copy(p, r.currentbuffer)
			r.currentbuffer = r.currentbuffer[i:]
			p = p[i:]
			n += i
		}

		// If DataRange is done, return number of bytes read and EOF.
		if r.data == nil {
			return n, io.EOF
		}

		if len(r.currentbuffer) == 0 {

			//The current buffer is empty - read in a new datapoint
			dp, err := r.data.Next()
			if err != nil {
				return n, err
			}
			// Datapoint reader is over
			if dp == nil {
				r.currentbuffer = r.Ender
				r.data.Close()
				r.data = nil
			} else {
				v, err := json.Marshal(dp)
				if err != nil {
					return n, err
				}
				r.currentbuffer = v
				r.separatorIndex = 0
			}
		}
	}
	return n, nil
}

// NewJsonReader creates a JsonReader with the given separator
func NewJsonReader(data DatapointIterator, starter string, separator string, footer string) (*JsonReader, error) {
	dp, err := data.Next()
	if err != nil {
		return nil, err
	}
	if dp == nil {
		// If there is no data, read as an empty array
		return &JsonReader{nil,[]byte(starter+footer),[]byte(separator),[]byte(footer),len(separator)},nil
	}
	v, err := json.Marshal(dp)
	if err != nil {
		return nil, err
	}
	return &JsonReader{data, []byte(starter + string(v)), []byte(separator), []byte(footer), len(separator)}, nil
}

// NewJsonArrayReader creates a new json array reader object. Allows using a RangeReader as an io.Reader type which outputs json.
// This reads the DataRange as a json array. (ie, [{},[],])
func NewJsonArrayReader(data DatapointIterator) (*JsonReader, error) {

	return NewJsonReader(data, "[", ",", "]")
}
