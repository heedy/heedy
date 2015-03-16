package timebatchdb

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

//A datapoint - contains a timestamp and associated data
//The data format is as follows:
//  [timestamp int64][datasize uvarint][data bytes]
type Datapoint struct {
	Buf []byte //The binary bytes associated with a datapoint
}

//Total size in bytes of the datapoint
func (d Datapoint) Len() int {
	return len(d.Buf)
}

//The length of the data of the databuffer
func (d Datapoint) DataLen() int {
	return len(d.Data())
}

//The byte array of the datapoint
func (d Datapoint) Bytes() []byte {
	return d.Buf
}

//Returns a nice pretty-printed representation of the datapoint
func (d Datapoint) String() string {
	return "[TIME=" + time.Unix(0, int64(d.Timestamp())).String() + " DATA=" + string(d.Data()) + "]"
}

//Returns the datapoint's timestamp
func (d Datapoint) Timestamp() (ts int64) {
	binary.Read(bytes.NewBuffer(d.Buf), binary.LittleEndian, &ts)
	return ts
}

//Returns the data associated with the datapoint
func (d Datapoint) Data() []byte {
	_, bytes_read := binary.Uvarint(d.Buf[8:])
	return d.Buf[8+bytes_read:]
}

//Reads the datapoint from file
func ReadDatapoint(r io.Reader) (Datapoint, error) {
	buf := new(bytes.Buffer)
	err := ReadDatapointIntoBuffer(r, buf)
	return Datapoint{buf.Bytes()}, err //The bytes the buffer read are our datapoint

}

//Given file, reads the datapoint into the given buffer. Used internally for initializing KeyedDatapoint
func ReadDatapointIntoBuffer(r io.Reader, w *bytes.Buffer) error {
	//First, write the timestamp int64
	_, err := io.CopyN(w, r, 8)
	if err != nil {
		return err
	}

	n, err := ReadUvarint(r)
	if err != nil {
		return err
	}
	WriteUvarint(w, n)
	_, err = io.CopyN(w, r, int64(n))
	return err
}

//Writes the datapoint structure into the given buffer
func DatapointIntoBuffer(w *bytes.Buffer, timestamp int64, data []byte) {
	binary.Write(w, binary.LittleEndian, timestamp)
	WriteUvarint(w, uint64(len(data)))
	w.Write(data)
}

//Creates a datapoint from a timetamp and data byte array
func NewDatapoint(timestamp int64, data []byte) Datapoint {
	buf := new(bytes.Buffer)
	DatapointIntoBuffer(buf, timestamp, data)
	return Datapoint{buf.Bytes()} //The bytes the buffer read are our datapoint
}

//Given an arbitrarily sized byte array, reads one datapoint from it (and uses a slice as its internal storage).
//This allows to read a large array of datapoints by using this function repeatedly.
func DatapointFromBytes(buf []byte) (d Datapoint, bytesread uint64) {
	size, bytes_read := binary.Uvarint(buf[8:]) //We just need to know how large the data is
	bytesread = 8 + uint64(bytes_read) + size
	return Datapoint{buf[:bytesread]}, bytesread
}
