package main

import (
    "os"
    "bytes"
    "encoding/binary"
    "time"
    "fmt"
)

type DataWriter struct {
    offsetf *os.File           //The offset file (contains time stamps)
    dataf *os.File             //The data storage file (blob)
    offsetb *bytes.Buffer    //A byte buffer of the offsets in this batch
    datab *bytes.Buffer      //A byte buffer of the data in this batch
    curloc int64            //The current location of the file's end including buffer
}

func (dw *DataWriter) Close() {
    dw.offsetf.Close()
    dw.dataf.Close()
}

func GetWriter(path string) (dw *DataWriter, err error) {
    //Opens the offset and data files for append
    offsetf,err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        return nil,err
    }
    dataf,err := os.OpenFile(path + ".data", os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        offsetf.Close()
        return nil,err
    }
    datastat, err := dataf.Stat()
    if (err != nil) {
        offsetf.Close()
        dataf.Close()
        return nil,err
    }
    offsetstat, err := offsetf.Stat()
    if (err != nil) {
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    //If the offset file is empty, write the size of datafile as location, which will be the location
    //  of the first write
    if (offsetstat.Size() == 0) {
        binary.Write(offsetf,binary.LittleEndian, datastat.Size())
    }

    dw = &DataWriter{offsetf,dataf,new(bytes.Buffer),new(bytes.Buffer),datastat.Size()}
    return dw,nil
}

func (dw *DataWriter) BatchWrite() (err error) {
    _,err = dw.dataf.Write(dw.datab.Bytes())   //Write the data byte buffer first, then write the offsets
    if (err != nil) {
        //TODO: I don't even know how to handle this, since we are append only.
        return err
    }
    _,err = dw.offsetf.Write(dw.offsetb.Bytes())

    //Clear the two buffers
    dw.datab.Reset()
    dw.offsetb.Reset()

    return err
}

func (dw *DataWriter) BatchInsert(timestamp int64, data []byte) {
    dw.datab.Write(data)

    //Each element of the batch has a timestamp, and location in the file
    binary.Write(dw.offsetb,binary.LittleEndian, timestamp)
    dw.curloc = dw.curloc + int64(len(data))
    binary.Write(dw.offsetb,binary.LittleEndian, dw.curloc)
}

func (dw *DataWriter) BatchInsertNow(data []byte) {
    dw.BatchInsert(time.Now().UnixNano(),data)
}

func (dw *DataWriter) Len() (int) {
    return dw.offsetb.Len()/16   //Each write to the offset buffer is 2 64 bit integers, which are 8 bytes each
}

func main() {
    x,err := GetWriter("./lol")

    if (err!=nil) {
        fmt.Printf("get err: %s\n",err)
        panic(0)
    }
    defer x.Close()

    fmt.Printf("BatchSize: %d\n",x.Len())
    x.BatchInsertNow([]byte("Hello World!"))
    fmt.Printf("BatchSize: %d\n",x.Len())
    x.BatchWrite()
    fmt.Printf("BatchSize: %d\n",x.Len())


}
