package datastream

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	//ErrInvalidDatapoint is thrown when inserts are attempted using datapoints which don't conform to the schema
	ErrInvalidDatapoint = errors.New("At least one datapoint did not conform to the stream's schema")
)

//A DatapointArray holds a couple useful functions that act on it
type DatapointArray []Datapoint

//LoadDatapointArray reads a DatapointArray from its corresponding bytes
func LoadDatapointArray(data []byte) (dpa DatapointArray, err error) {
	err = msgpack.Unmarshal(data, &dpa)
	return dpa, err
}

//Bytes writes the DatapointArray into binary data
func (dpa DatapointArray) Bytes() ([]byte, error) {
	return msgpack.Marshal(dpa)
}

//IsEqual checks if two DatapointArrays contain the same data
func (dpa DatapointArray) IsEqual(d DatapointArray) bool {
	if len(d) != len(dpa) {
		return false
	}
	for i := range d {
		if !d[i].IsEqual(dpa[i]) {
			return false
		}
	}
	return true
}

//VerifySchema ensures that all of the datapoints in the given array conform to the given schema
func (dpa DatapointArray) VerifySchema(schema *gojsonschema.Schema) (err error) {
	for i := range dpa {
		if !dpa[i].HasSchema(schema) {
			return ErrInvalidDatapoint
		}
	}
	return nil
}

func (dpa DatapointArray) String() string {
	tot := "DatapointArray{"
	for i := range dpa {
		tot += dpa[i].String() + ","
	}
	return tot[:len(tot)-1] + "}"
}

//SplitIntoChunks splits the datapoint array into chunks of up to chunksize, and then converts the chunks to byte arrays.
func (dpa DatapointArray) SplitIntoChunks(chunksize int) ([][]byte, error) {

	//First prepare the array for all the chunks (with an extra element for a possible non-full element)
	chunknum := len(dpa) / chunksize
	chunks := make([][]byte, chunknum, chunknum+1)

	for i := 0; i < chunknum; i++ {
		b, err := dpa[i*chunksize : (i+1)*chunksize].Bytes()
		if err != nil {
			return nil, err
		}
		chunks[i] = b
	}

	//Now fill the partial chunks
	if len(dpa)%chunksize > 0 {
		b, err := dpa[chunknum*chunksize : len(dpa)].Bytes()
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, b)
	}

	return chunks, nil
}
