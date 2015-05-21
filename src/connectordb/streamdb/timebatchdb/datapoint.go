package timebatchdb

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

//A Datapoint contains a timestamp, the data payload, and an optional key
//The data format is as follows:
//  [timestamp int64][datasize uvarint][data bytes][keysize uvarint][key bytes]
//If key is not given, keysize is set to 0.
type Datapoint struct {
	Buf []byte //The binary bytes associated with a datapoint
}

//Len returns the size in bytes of the datapoint
func (d Datapoint) Len() int {
	return len(d.Buf)
}

//DataLen returns length of the data bytes stored in the datapoint
func (d Datapoint) DataLen() int {
	return len(d.Data())
}

//Bytes returns the byte array representation of the entire datapoint
func (d Datapoint) Bytes() []byte {
	return d.Buf
}

//String returns a nice pretty-printed representation of the datapoint
func (d Datapoint) String() string {
	s := "[TIME=" + time.Unix(0, d.Timestamp()).String() + " DATA=" + string(d.Data())
	if k := d.Key(); k != "" { //The key is optional
		s += " KEY=" + k
	}
	return s + " ]"
}

//Timestamp returns the datapoint's timestamp
func (d Datapoint) Timestamp() (ts int64) {
	binary.Read(bytes.NewBuffer(d.Buf), binary.LittleEndian, &ts)
	return ts
}

//Data returns the data byte array associated with the datapoint
func (d Datapoint) Data() []byte {
	datasize, bytesRead := binary.Uvarint(d.Buf[8:])
	return d.Buf[8+uint64(bytesRead) : 8+uint64(bytesRead)+datasize]
}

//Key returns the string key associated with the datapoint
func (d Datapoint) Key() string {
	datasize, bytesRead := binary.Uvarint(d.Buf[8:])
	startloc := 8 + uint64(bytesRead) + datasize
	_, bytesRead = binary.Uvarint(d.Buf[startloc:])
	return string(d.Buf[startloc+uint64(bytesRead):])
}

//ReadDatapoint reads a datapoint from a file
func ReadDatapoint(r io.Reader) (Datapoint, error) {
	buf := new(bytes.Buffer)
	err := ReadDatapointIntoBuffer(r, buf)
	return Datapoint{buf.Bytes()}, err //The bytes the buffer read are our datapoint

}

//ReadDatapointIntoBuffer reads the datapoint from an io.Reader into the given buffer. Used internally for initializing KeyedDatapoint
func ReadDatapointIntoBuffer(r io.Reader, w *bytes.Buffer) error {
	//First, write the timestamp int64
	_, err := io.CopyN(w, r, 8)
	if err != nil {
		return err
	}

	//Next copy the data length
	n, err := ReadUvarint(r)
	if err != nil {
		return err
	}
	WriteUvarint(w, n)

	//Then copy the actual data
	_, err = io.CopyN(w, r, int64(n))
	if err != nil {
		return err
	}

	//After that, copy the key length
	n, err = ReadUvarint(r)
	if err != nil {
		return err
	}
	WriteUvarint(w, n)

	//And lastly, copy the key data
	_, err = io.CopyN(w, r, int64(n))

	return err
}

//DatapointIntoBuffer writes the datapoint structure into the given buffer
func DatapointIntoBuffer(w *bytes.Buffer, timestamp int64, data []byte, key string) {
	binary.Write(w, binary.LittleEndian, timestamp)
	WriteUvarint(w, uint64(len(data)))
	w.Write(data)
	keyb := []byte(key)
	WriteUvarint(w, uint64(len(keyb)))
	w.Write(keyb)
}

//NewDatapoint creates a datapoint from a timetamp and data byte array
func NewDatapoint(timestamp int64, data []byte, key string) Datapoint {
	buf := new(bytes.Buffer)
	DatapointIntoBuffer(buf, timestamp, data, key)
	return Datapoint{buf.Bytes()} //The bytes the buffer read are our datapoint
}

//DatapointFromBytes takes an arbitrarily sized byte array, and reads one datapoint from it (and uses a slice of the byte array as its internal storage).
//This allows to read a large array of datapoints by using this function repeatedly.
func DatapointFromBytes(buf []byte) (d Datapoint, bytesread uint64) {
	size, bytesRead := binary.Uvarint(buf[8:]) //Get the data length
	bytesread = 8 + uint64(bytesRead) + size
	size, bytesRead = binary.Uvarint(buf[bytesread:]) //Next get the key length
	bytesread += uint64(bytesRead) + size
	return Datapoint{buf[:bytesread]}, bytesread
}
