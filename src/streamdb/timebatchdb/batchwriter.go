package timebatchdb

import (
    "bytes"
    "encoding/binary"
    "sync"
    "time"
    )

type BatchWriter struct {
    IndexBuffer *bytes.Buffer    //A buffer which stores the batchfile elements
    DataBuffer *bytes.Buffer     //A buffer which stores the blob data to be written
    LastTime uint64         //The timestamp of the most recent datapoint
    writelock sync.Mutex    //The writeLock - when writelock is on, the batch is being written
}

//Returns the number of data points buffered in the BatchWriter.
func (bw *BatchWriter) Len() int {
    return (bw.IndexBuffer.Len()-8)/16
}

//Returns the size in bytes of the byte array of data buffered thus far. This does
//not include the timestamps or other metadata which will also be written to file.
func (bw *BatchWriter) Size() int {
    return bw.DataBuffer.Len()
}

//Unlocks the BatchWriter write mutex.
func (bw *BatchWriter) Unlock() {
    bw.writelock.Unlock()
}

//Locks the batchWriter such that no inserts are allowed. This is used while writing to file
func (bw *BatchWriter) Lock() () {
    bw.writelock.Lock()
}

//Given a timestamp (Use Time.UnixNano()), and data bytes, adds them to the batch buffer
func (bw *BatchWriter) Insert(timestamp uint64, data []byte) {
    bw.writelock.Lock()

    bw.DataBuffer.Write(data)

    binary.Write(bw.IndexBuffer,binary.LittleEndian,timestamp)
    binary.Write(bw.IndexBuffer,binary.LittleEndian,int64(bw.DataBuffer.Len()))

    bw.LastTime = timestamp


    bw.writelock.Unlock()
}

//Inserts the given bytes with the current time as timestamp
func (bw *BatchWriter) InsertNow(data []byte) {
    bw.Insert(uint64(time.Now().UnixNano()),data)
}

func (bw *BatchWriter) Clear() {
    bw.DataBuffer.Reset()
    bw.IndexBuffer.Reset()

    binary.Write(bw.IndexBuffer,binary.LittleEndian,int64(0))
}

func NewBatchWriter() (*BatchWriter) {
    x := &BatchWriter{new(bytes.Buffer),new(bytes.Buffer),0,sync.Mutex{}}
    x.Clear()   //Set it up so it is ready to write next batch
    return x
}
