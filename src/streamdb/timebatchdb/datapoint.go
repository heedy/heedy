package timebatchdb

import (
    "time"
    "bytes"
    "encoding/binary"
    "io"
    )


//A datapoint - contains a timestamp and associated data
//The data format is as follows:
//  [timestamp uint64][datasize uvarint][data bytes]
type Datapoint struct {
    Buf []byte                 //The binary bytes associated with a datapoint
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

func (d Datapoint) String() string {
    return "[TIME="+time.Unix(0,int64(d.Timestamp())).String()+" DATA="+string(d.Data())+"]"
}

//Returns the datapoint's timestamp
func (d Datapoint) Timestamp() (ts uint64) {
    binary.Read(bytes.NewBuffer(d.Buf),binary.LittleEndian,&ts)
    return ts
}

//Returns the data associated with the datapoint
func (d Datapoint) Data() ([]byte) {
    _, bytes_read := binary.Uvarint(d.Buf[8:])
    return d.Buf[8+bytes_read:]
}

//Gets the datapoint
func (d Datapoint) Get() (timestamp uint64,data []byte) {
    return d.Timestamp(),d.Data()   //The return here is simple, since we don't do any real processing
}

//Reads the datapoint from file
func ReadDatapoint(r io.Reader) (Datapoint,error) {
    buf := new(bytes.Buffer)
    err := ReadDatapointIntoBuffer(r,buf)
    return Datapoint{buf.Bytes()},err  //The bytes the buffer read are our datapoint

}

//Given file, reads the datapoint into the given buffer. Used internally for initializing KeyedDatapoint
func ReadDatapointIntoBuffer(r io.Reader,w *bytes.Buffer) error {
    //First, write the timestamp uint64
    _,err  := io.CopyN(w,r,8)
    if err!=nil {
        return err
    }

    n,err := ReadUvarint(r)
    if err!=nil {
        return err
    }
    WriteUvarint(w,n)
    _,err  = io.CopyN(w,r,int64(n))
    return err
}

//Writes the datapoint structure into the given buffer
func DatapointIntoBuffer(w *bytes.Buffer, timestamp uint64,data []byte) {
    binary.Write(w,binary.LittleEndian,timestamp)
    WriteUvarint(w,uint64(len(data)))
    w.Write(data)
}

func NewDatapoint(timestamp uint64, data []byte) Datapoint {
    buf := new(bytes.Buffer)
    DatapointIntoBuffer(buf,timestamp,data)
    return Datapoint{buf.Bytes()}  //The bytes the buffer read are our datapoint
}

//Given an arbitrarily sized byte array, reads one datapoint from it (and uses a slice as its internal storage).
//This allows to read a large array of datapoints by using this function repeatedly.
func DatapointFromBytes(buf [] byte) (d Datapoint, bytesread uint64) {
    size,bytes_read := binary.Uvarint(buf[8:]) //We just need to know how large the data is
    bytesread = 8+uint64(bytes_read)+size
    return Datapoint{buf[:bytesread]},bytesread
}


//Same as datapoint, but also contains its key.
//The format is as follows:
//  [keylen uvarint][key bytes][datapoint]
type KeyedDatapoint struct {
    Buf []byte                 //The binary bytes associated with a datapoint (along with its key)
}

//Returns the byte array associated with the keyed datapoint
func (d KeyedDatapoint) Bytes() []byte {
    return d.Buf
}

//Gets the datapoint from a keyed-datapoint
func (d KeyedDatapoint) Datapoint() Datapoint {
    //Find the location of the datapoint
    size,bytes_read := binary.Uvarint(d.Buf)
    return Datapoint{d.Buf[bytes_read+int(size):]}
}

func (d KeyedDatapoint) Key() string {
    size,bytes_read := binary.Uvarint(d.Buf)
    return string(d.Buf[bytes_read:bytes_read+int(size)])
}

func (d KeyedDatapoint) Timestamp() uint64 {
    return d.Datapoint().Timestamp()
}
func (d KeyedDatapoint) Data() []byte {
    return d.Datapoint().Data()
}

//A standard way to create a datapoint
func NewKeyedDatapoint(key string, timestamp uint64, data []byte) KeyedDatapoint {
    buf := new(bytes.Buffer)
    bytekey := []byte(key)
    WriteUvarint(buf,uint64(len(bytekey)))
    buf.Write(bytekey)
    DatapointIntoBuffer(buf,timestamp,data)
    return KeyedDatapoint{buf.Bytes()}

}

//Reads the datapoint from file
func ReadKeyedDatapoint(r io.Reader) (KeyedDatapoint,error) {
    w := new(bytes.Buffer)
    n,err := ReadUvarint(r)
    if err!=nil {
        return KeyedDatapoint{nil},err
    }
    WriteUvarint(w,n)
    _,err  = io.CopyN(w,r,int64(n))
    if err!=nil {
        return KeyedDatapoint{nil},err
    }
    err = ReadDatapointIntoBuffer(r,w)
    return KeyedDatapoint{w.Bytes()},err //The bytes the buffer read are our datapoint

}

func (d KeyedDatapoint) String() string {
    return "[KEY="+d.Key()+" TIME="+time.Unix(0,int64(d.Timestamp())).String()+" DATA="+string(d.Data())+"]"
}

//Total size in bytes of the datapoint
func (d KeyedDatapoint) Len() int {
    return len(d.Buf)
}

//The length of the data of the databuffer
func (d KeyedDatapoint) DataLen() int {
    return len(d.Data())
}
