package timebatchdb

import (
	"bytes"
	"compress/gzip"
	"container/list"
)

//DatapointArray represents an array of Datapoints. It can be used as a DataRange, and has methods for reading and writing
//from/to binary.
type DatapointArray struct {
	Datapoints []Datapoint //The array of datapoints
	array      []byte      //A (possibly nil) single byte array which holds all of the datapoints
	rangeindex int         //Allows to make this a DataRange object
}

//Len gives the number of datapoints contained in the array
func (d *DatapointArray) Len() int {
	return len(d.Datapoints)
}

//Size returns total size of the array in bytes (uncompressed)
func (d *DatapointArray) Size() int {
	if d.array != nil {
		return len(d.array)
	}
	size := 0
	for i := 0; i < len(d.Datapoints); i++ {
		size += d.Datapoints[i].Len()
	}
	return size
}

//Init is a dummy function - it allows datapointArray to conform to the DataRange interface. It doesn't actually do anything.
func (d *DatapointArray) Init() error {
	return nil
}

//Close is a DataRange function - allows DatapointArray to conform to the DataRange interface.
//The DatapointArray doesn't need closing - this method resets the DataRange iterator.
func (d *DatapointArray) Close() {
	d.rangeindex = 0
}

//Next allows to use DatapointArray as a DataRange - starts from the first datapoint, and
//successively returns datapoint ptrs until there are none left, at which point it returns nil.
//It is an iterator.
func (d *DatapointArray) Next() (*Datapoint, error) {
	if d.rangeindex >= d.Len() {
		return nil, nil
	}
	dp := &d.Datapoints[d.rangeindex]
	d.rangeindex++
	return dp, nil
}

//IsTimestampOrdered checks if the DatapointArray is ordered with strictly increasing timestamps
func (d *DatapointArray) IsTimestampOrdered() bool {
	if d.Len() <= 1 {
		return true
	}
	prevt := d.Datapoints[0].Timestamp()
	for i := 1; i < d.Len(); i++ {
		ts := d.Datapoints[i].Timestamp()
		if ts <= prevt {
			return false
		}
		prevt = ts
	}
	return true
}

//TimestampIRange returns the timestamps associated with the index range
func (d *DatapointArray) TimestampIRange(start int, end int) (timestamps []int64) {
	if end > d.Len() {
		end = d.Len()
	}
	if start > end {
		return nil
	}
	timestamps = make([]int64, end-start)
	for i := start; i < end; i++ {
		timestamps[i] = d.Datapoints[i].Timestamp()
	}
	return timestamps
}

//Timestamps returns the array of timestamps
func (d *DatapointArray) Timestamps() (timestamps []int64) {
	return d.TimestampIRange(0, d.Len())
}

//DataIRange returns the data associated with the index range
func (d *DatapointArray) DataIRange(start int, end int) (data [][]byte) {
	if end > d.Len() {
		end = d.Len()
	}
	if start > end {
		return nil
	}
	data = make([][]byte, end-start)
	for i := start; i < end; i++ {
		data[i] = d.Datapoints[i].Data()
	}
	return data
}

//Data returns the array of data for all the datapoints in the array
func (d *DatapointArray) Data() (data [][]byte) {
	return d.DataIRange(0, d.Len())
}

//Get returns the array of timestamps and the array of associated data
func (d *DatapointArray) Get() (timestamps []int64, data [][]byte) {
	return d.Timestamps(), d.Data()
}

//FindTimeIndex finds the index of the first datapoint in the array which has a timestamp strictly greater
//than the given reference timestamp.
//If no datapoints fit this, returns -1
//(ie, no datapoint in array has a timestamp greater than the given time)
func (d *DatapointArray) FindTimeIndex(timestamp int64) int {
	//BUG(daniel): This code makes no guarantees about nanosecond-level precision.
	if d.Len() == 0 {
		return -1
	}

	leftbound := 0
	leftts := d.Datapoints[0].Timestamp()

	//If the timestamp is earlier than the earliest datapoint
	if leftts > timestamp {
		return 0
	}

	rightbound := d.Len() - 1 //Len is guaranteed > 0
	rightts := d.Datapoints[rightbound].Timestamp()

	if rightts <= timestamp {
		return -1
	}

	//We do this shit logn style
	for rightbound-leftbound > 1 {
		midpoint := (leftbound + rightbound) / 2
		ts := d.Datapoints[midpoint].Timestamp()
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
func (d *DatapointArray) TStart(timestamp int64) *DatapointArray {
	i := d.FindTimeIndex(timestamp)
	if i == -1 {
		return nil
	}
	return NewDatapointArray(d.Datapoints[i:])
}

//DatapointTRange returns the DatapointArray of datapoints which fit within the time range:
//  (timestamp1,timestamp2]
func (d *DatapointArray) DatapointTRange(timestamp1 int64, timestamp2 int64) *DatapointArray {
	i1 := d.FindTimeIndex(timestamp1)
	if i1 == -1 {
		return nil
	}
	i2 := d.FindTimeIndex(timestamp2)
	if i2 == -1 {
		//The endrange is out of bounds - read until the end of the array
		return NewDatapointArray(d.Datapoints[i1:])
	}
	return NewDatapointArray(d.Datapoints[i1:i2])
}

//DataTRange returns the data array of datapoint between the two given timestamps
func (d *DatapointArray) DataTRange(timestamp1 int64, timestamp2 int64) [][]byte {
	return d.DatapointTRange(timestamp1, timestamp2).Data()
}

//TimestampTRange returns the array of timestamps of dataponits between the two given timestamps
func (d *DatapointArray) TimestampTRange(timestamp1 int64, timestamp2 int64) []int64 {
	return d.DatapointTRange(timestamp1, timestamp2).Timestamps()
}

//GetTRange returns the array of timestamps and data which fit in the given time range:
//  (timestamp1,timestamp2]
func (d *DatapointArray) GetTRange(timestamp1 int64, timestamp2 int64) ([]int64, [][]byte) {
	return d.DatapointTRange(timestamp1, timestamp2).Get()
}

//Bytes returns the byte array associated with the entire page of datapoints (uncompressed)
func (d *DatapointArray) Bytes() []byte {
	if d.array != nil {
		return d.array
	}
	//The array does not exist. We therefore create it
	arr := make([]byte, d.Size())
	n := 0
	for i := 0; i < len(d.Datapoints); i++ {
		numWritten := copy(arr[n:], d.Datapoints[i].Bytes())
		//In the interest of saving memory, have the datapoints refer to slices of the newly created
		//byte array, rather than having multiple copies of the same data
		d.Datapoints[i] = Datapoint{arr[n : n+numWritten]}
		n += numWritten
	}
	d.array = arr
	return arr
}

//CompressedBytes returns the gzipped bytes of the entire array of datapoints
func (d *DatapointArray) CompressedBytes() []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(d.Bytes())
	w.Close()
	return b.Bytes()
}

//NewDatapointArray creates a DatapointArray given an actual array of Datapoint
func NewDatapointArray(d []Datapoint) *DatapointArray {
	return &DatapointArray{d, nil, 0}
}

//CreateDatapointArray creates DatapointArray from the raw data - it is assumed that there is only one key for the array (which is an
//assumption that is valid for the major use case of timebatchdb)
func CreateDatapointArray(timestamps []int64, data [][]byte, key string) *DatapointArray {
	arr := make([]Datapoint, len(timestamps))

	for i := 0; i < len(arr); i++ {
		arr[i] = NewDatapoint(timestamps[i], data[i], key)
	}
	return NewDatapointArray(arr)
}

//DatapointArrayFromBytes creates a datapoint array from its associated bytes. Note that the Datapoint array assumes
//that the bytes are correctly sized. The DatapointArray does not actually store its size, so
//this function just keeps trying to read data until there are no Bytes left.
func DatapointArrayFromBytes(data []byte) *DatapointArray {
	if len(data) == 0 {
		return nil
	}
	n := 0
	l := list.New()
	for n < len(data) {
		dp, num := DatapointFromBytes(data[n:])
		l.PushBack(dp)
		n += int(num)
	}

	dp := DatapointArrayFromList(l)
	dp.array = data[:n] //We can set the array too

	return dp
}

//DatapointArrayFromList creates a DatapointArray given a linked list containing Datapoint
func DatapointArrayFromList(l *list.List) *DatapointArray {
	//Now create the points array
	points := make([]Datapoint, l.Len())
	elem := l.Front()
	points[0] = elem.Value.(Datapoint)
	for j := 1; elem.Next() != nil; j++ {
		elem = elem.Next()
		points[j] = elem.Value.(Datapoint)
	}

	return &DatapointArray{points, nil, 0}
}

//DatapointArrayFromCompressedBytes decmpresses the correctly sized byte array for the compressed representation of a DatapointArray
func DatapointArrayFromCompressedBytes(cdata []byte) *DatapointArray {
	r, _ := gzip.NewReader(bytes.NewBuffer(cdata))

	l := list.New()

	d, err := ReadDatapoint(r)
	for err == nil {
		l.PushBack(d)
		d, err = ReadDatapoint(r)
	}
	r.Close()

	return DatapointArrayFromList(l)
}

//DatapointArrayFromDataRange creates a DatapointArray based upon a given DataRange. Closes the DataRange when done
func DatapointArrayFromDataRange(dr DataRange) (*DatapointArray, error) {
	l := list.New()

	d, err := dr.Next()
	for d != nil {
		l.PushBack(*d)
		d, err = dr.Next()
	}
	dr.Close()

	if err != nil {
		return nil, err
	}

	return DatapointArrayFromList(l), nil
}
