package filedb

import (
    "os"
    "bytes"
    "encoding/binary"
    "sync"
    )

type KeyWriterBatch struct {
    indexb *bytes.Buffer    //A buffer which stores the batchfile elements
    datab *bytes.Buffer     //A buffer which stores the blob data to be written
    previndex uint64        //The location of the key's previous batch
    dataindex uint64        //The databatch's data index
    lasttime uint64         //The timestamp of the most recent datapoint
    writelock sync.Mutex    //The writeLock - when writelock is on, the batch is being written
}

func (kcwb *KeyWriterBatch) Write(indexf *os.File,dataf *os.File) (dataw int64, indexw int64, err error) {
    kcwb.writelock.Lock()

    dataw,err = dataf.Write(kcwb.datab.Bytes())
    if (err != nil) {
        return dataw,0,err
    }

    indexw,err = indexf.Write(kcwb.indexb.Bytes())
    if (err != nil) {
        return dataw,indexw,err
    }

    kcwb.Clear()

    kcwb.writelock.Unlock()
}

func (kcwb *KeyWriterBatch) Insert(timestamp uint64, data []byte) {
    kcwb.writelock.Lock()

    kcwb.datab.Write(data)

    binary.Write(kcwb.indexb,binary.LittleEndian,timestamp)
    binary.Write(kcwb.indexb,binary.LittleEndian,int64(kcwb.datab.Len()))

    kcwb.lasttime = timestamp


    kcwb.writelock.Unlock()
}

func (kcwb *KeyWriterBatch) Clear() {
    kcwb.datab.Reset()
    kcwb.indexb.Reset()

    binary.Write(kcwb.indexb,binary.LittleEndian,int64(0))
}

func NewKeyWriterBatch(previndex uint64,dataindex uint64) (*KeyCacheWriterBatch) {
    x := &KeyWriterBatch{new(bytes.Buffer),new(bytes.Buffer),previndex,dataindex,sync.Mutex{}}
    x.Clear()   //Set it up so it is ready to write next batch
    return x
}
