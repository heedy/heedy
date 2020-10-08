package timeseries

import (
	"encoding/json"
	"io"

	"github.com/mailru/easyjson"
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
		return &JsonReader{nil, []byte(starter + footer), []byte(separator), []byte(footer), len(separator)}, nil
	}
	v, err := json.Marshal(dp)
	if err != nil {
		return nil, err
	}
	return &JsonReader{data, []byte(starter + string(v)), []byte(separator), []byte(footer), len(separator)}, nil
}

type JsonArrayReader struct {
	DatapointIterator
	buffer    []byte
	done      bool
	batchsize int
}

func (r *JsonArrayReader) fillArray() (dpa DatapointArray, err error) {
	dpa = make(DatapointArray, 0, r.batchsize)
	var dp *Datapoint
	for i := 0; i < r.batchsize; i++ {
		dp, err = r.DatapointIterator.Next()
		if dp == nil || err != nil {
			r.done = true
			break
		}
		dpa = append(dpa, dp)
	}
	return dpa, err
}

func (r *JsonArrayReader) fillBuffer() error {
	dpa, err := r.fillArray()
	if err != nil {
		return err
	}
	if len(dpa) == 0 {
		// The data ended
		r.buffer = []byte{']'}
		return nil
	}
	b, err := easyjson.Marshal(dpa)
	if err != nil {
		return err
	}
	b[0] = ','
	if !r.done {
		b = b[:len(b)-1] // remove end bracket
	}
	r.buffer = b
	return nil
}

// Read reads the given number of bytes from the DataRange, and p is the buffer to read into
func (r *JsonArrayReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.buffer)
	r.buffer = r.buffer[n:]
	if n == len(p) {
		if r.done && len(r.buffer) == 0 {
			err = io.EOF
		}
		return
	}
	if r.done {
		err = io.EOF
		return
	}
	err = r.fillBuffer()
	if err != nil {
		return
	}
	var m int
	m, err = r.Read(p[n:])
	return m + n, err
}

// NewJsonArrayReader converts a DatapointIterator into an io.Reader, allowing writing of an arbitrarily large-sized response
func NewJsonArrayReader(data DatapointIterator, batchsize int) (*JsonArrayReader, error) {
	jar := &JsonArrayReader{
		DatapointIterator: data,
		buffer:            []byte{},
		done:              false,
		batchsize:         batchsize,
	}
	err := jar.fillBuffer()
	if err == nil {
		if len(jar.buffer) == 1 {
			// An empty array is what we actually want
			jar.buffer = []byte{'[', ']'}
		} else {
			jar.buffer[0] = '['
		}
	}
	return jar, err
}
