package datastream

import (
	"bytes"
	"compress/gzip"
	"errors"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	//ErrInvalidDatapoint is thrown when inserts are attempted using datapoints which don't conform to the schema
	ErrInvalidDatapoint = errors.New("At least one datapoint did not conform to the stream's schema")

	//ErrorVersion is returned when the data returned from the database is of an unknown binary version
	ErrorVersion = errors.New("Unrecognized binary data version.")

	//ErrDecompress is called when decompress panics
	ErrDecompress = errors.New("Failed to decompress the data")
)

const (
	MsgPackVersion           = 1 //MsgPackVersion is the version of data encoding which uses MsgPack
	CompressedMsgPackVersion = 2 //CompressedMsgPackVersion is msgpack compressed with gzip
)

//A DatapointArray holds a couple useful functions that act on it
type DatapointArray []Datapoint

//DatapointArrayFromBytes reads a DatapointArray from its corresponding bytes
func DatapointArrayFromBytes(data []byte) (dpa DatapointArray, err error) {
	err = msgpack.Unmarshal(data, &dpa)
	return dpa, err
}

//DecodeDatapointArray - given bytes and a version string returns the decoded DatapointArray
func DecodeDatapointArray(data []byte, version int) (dpa *DatapointArray, err error) {
	switch version {
	case MsgPackVersion:
		da, err := DatapointArrayFromBytes(data)
		if err != nil {
			return nil, err
		}
		return &da, err
	case CompressedMsgPackVersion:
		da, err := DatapointArrayFromCompressedBytes(data)
		if err != nil {
			return nil, err
		}
		return &da, err
	default:
		return nil, ErrorVersion

	}
}

//DatapointArrayFromDataStrings is given the strings of data, each associated with the bytes
//of a datapoint, and it converts them to a DatapointArray
func DatapointArrayFromDataStrings(data []string) (dpa DatapointArray, err error) {
	if len(data) == 0 {
		return DatapointArray{}, nil
	}

	dpa = make(DatapointArray, len(data))

	for i := range data {
		dpa[i], err = DatapointFromBytes([]byte(data[i]))
		if err != nil {
			return nil, err
		}
	}
	return dpa, nil
}

//Encode encodes the dataponit array according to the data version
func (dpa DatapointArray) Encode(version int) ([]byte, error) {
	switch version {
	case MsgPackVersion:
		return dpa.Bytes()
	case CompressedMsgPackVersion:
		return dpa.CompressedBytes()
	default:
		return nil, ErrorVersion
	}
}

//Bytes writes the DatapointArray into binary data
func (dpa DatapointArray) Bytes() ([]byte, error) {
	return msgpack.Marshal(dpa)
}

//DatapointArrayFromCompressedBytes decmpresses the correctly sized byte array for the compressed representation of a DatapointArray
func DatapointArrayFromCompressedBytes(cdata []byte) (dpa DatapointArray, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrDecompress
		}
	}()

	r, _ := gzip.NewReader(bytes.NewBuffer(cdata))
	dec := msgpack.NewDecoder(r)
	err = dec.Decode(&dpa)
	return dpa, err
}

//CompressedBytes returns the gzipped bytes of the entire array of datapoints
func (dpa DatapointArray) CompressedBytes() ([]byte, error) {
	dpab, err := dpa.Bytes()
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(dpab)
	w.Close()
	return b.Bytes(), nil
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
	if len(dpa) == 0 {
		return tot + "}"
	}
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

//Length returns the length of the DatapointArray
func (dpa DatapointArray) Length() int {
	return len(dpa)
}

//IsTimestampOrdered checks if the DatapointArray is ordered with increasing timestamps
func (dpa DatapointArray) IsTimestampOrdered() bool {
	for i := 1; i < len(dpa); i++ {
		if dpa[i].Timestamp < dpa[i-1].Timestamp {
			return false
		}
	}
	return true
}

//FindTimeIndex finds the index of the first datapoint in the array which has a timestamp strictly greater
//than the given reference timestamp.
//If no datapoints fit this, returns -1
//(ie, no datapoint in array has a timestamp greater than the given time)
func (dpa DatapointArray) FindTimeIndex(timestamp float64) int {
	if len(dpa) == 0 {
		return -1
	}

	leftbound := 0
	leftts := dpa[0].Timestamp

	//If the timestamp is earlier than the earliest datapoint
	if leftts > timestamp {
		return 0
	}

	rightbound := len(dpa) - 1 //Len is guaranteed > 0
	rightts := dpa[rightbound].Timestamp

	if rightts <= timestamp {
		return -1
	}

	//We do this shit logn style
	for rightbound-leftbound > 1 {
		midpoint := (leftbound + rightbound) / 2
		ts := dpa[midpoint].Timestamp
		if ts <= timestamp {
			leftbound = midpoint
			leftts = ts
		} else {
			rightbound = midpoint
			rightts = ts
		}
	}
	return rightbound
}

//TStart returns a DatapointArray which has the given starting bound (like DatapointTRange, but without upperbound)
func (dpa DatapointArray) TStart(timestamp float64) DatapointArray {
	i := dpa.FindTimeIndex(timestamp)
	if i == -1 {
		return nil
	}
	return dpa[i:]
}

//TRange returns the DatapointArray of datapoints which fit within the time range:
//  (timestamp1,timestamp2]
func (dpa DatapointArray) TRange(timestamp1 float64, timestamp2 float64) DatapointArray {
	i1 := dpa.FindTimeIndex(timestamp1)
	if i1 == -1 {
		return nil
	}
	i2 := dpa.FindTimeIndex(timestamp2)
	if i2 == -1 {
		//The endrange is out of bounds - read until the end of the array
		return dpa[i1:]
	}
	return dpa[i1:i2]
}

//IRange returns the DatapointArray in the given range - it is just a convenience function since I use pointers
func (dpa DatapointArray) IRange(i1, i2 int) *DatapointArray {
	tmp := dpa[i1:i2]
	return &tmp
}
