package main

import (
    "os"
    "encoding/binary"
    "bytes"
    "time"
    "fmt"
    "errors"
)

type DataReader struct {
    offsetf *os.File            //The offset file (contains time stamps)
    dataf *os.File              //The data storage file (blob)
    size int64                  //The number of entries written when last checked
}

func (dr *DataReader) Close() {
    dr.offsetf.Close()
    dr.dataf.Close()
}

func (dr *DataReader) Len() (int64) {
    //Check the underlying file size - maybe it changed
    ostat,err := dr.offsetf.Stat()
    if ( err != nil) {
        return dr.size
    }
    dr.size = (ostat.Size()-8)/16   //16 bytes are written for each entry

    return dr.size
}

func GetReader(path string) (dr *DataReader, err error) {
    //Opens the offset and data files for append
    offsetf,err := os.OpenFile(path,os.O_RDONLY, 0666)
    if (err != nil) {
        return nil,err
    }
    dataf,err := os.OpenFile(path + ".data",os.O_RDONLY, 0666)
    if (err != nil) {
        offsetf.Close()
        return nil,err
    }

    dr = &DataReader{offsetf,dataf,0}
    dr.Len()    //Find the size of the file - update internal size

    return dr,nil
}

func (dr *DataReader) Read(index int64) (timestamp int64, data []byte, err error) {
    //Makes sure that the length is within range
    if (index >= dr.size) {
        if (index >= dr.Len()) {
            return 0,nil,errors.New("Index out of bounds")
        }
    }

    //The index is within bounds - read the offsetfile. The offsetfile is written: (startloc,timestamp,endloc)
    offsetbuffer := make([]byte, 8*3)
    dr.offsetf.ReadAt(offsetbuffer,2*8*index)

    //Decode the item
    var startloc int64
    var endloc int64
    buf := bytes.NewReader(offsetbuffer)
    binary.Read(buf,binary.LittleEndian,&startloc)
    binary.Read(buf,binary.LittleEndian,&timestamp)
    binary.Read(buf,binary.LittleEndian,&endloc)

    if (startloc > endloc) {
        return 0,nil,errors.New("File Corrupted")
    }
    if (startloc == endloc) {
        return timestamp,[]byte{},nil //If there is nothing to read, return empty bytes
    }

    databuffer := make([]byte,endloc-startloc)
    dr.dataf.ReadAt(databuffer,startloc)

    return timestamp,databuffer,nil
}

func (dr *DataReader) ReadBatch(startindex int64,endindex int64) (timestamp []int64, data [][]byte, err error) {
    //Makes sure that the length is within range
    if (endindex > dr.size) {
        if (endindex > dr.Len()) {
            return nil,nil,errors.New("Index out of bounds")
        }
    }
    if (endindex <= startindex) {
        return nil,nil,errors.New("startindex and end index set incorrectly")
    }

    numread := endindex - startindex
    timestamp = make([]int64,numread)
    data = make([][]byte,numread)
    locs := make([]int64,numread+1) //The +1 is because there is one extra start location

    //The index is within bounds - read the offsetfile. The offsetfile is written: (startloc,timestamp,endloc)
    offsetbuffer := make([]byte, 16*numread+8)
    dr.offsetf.ReadAt(offsetbuffer,2*8*startindex)
    buf := bytes.NewReader(offsetbuffer)

    //Decode the offsetfile chunk
    for i := int64(0); i < numread; i++ {
        binary.Read(buf,binary.LittleEndian,&locs[i])
        binary.Read(buf,binary.LittleEndian,&timestamp[i])
    }
    binary.Read(buf,binary.LittleEndian,&locs[numread])

    if (locs[0] > locs[numread]) {
        return nil,nil,errors.New("File Corrupted")
    }

    //Read the data into the byte arrays
    databuffer := make([]byte,locs[numread]-locs[0])
    dr.dataf.ReadAt(databuffer,locs[0])
    for i := int64(0); i < numread; i++ {
        data[i] = databuffer[locs[i]-locs[0]:locs[i+1]-locs[0]]
    }

    return timestamp,data,nil
}


func main() {
    x,err := GetReader("./lol")

    if (err!=nil) {
        fmt.Printf("get err: %s\n",err)
        panic(0)
    }
    defer x.Close()

    timestamp,data,err := x.Read(1)
    if (err != nil) {
        fmt.Printf("Read error\n")
    } else {
        fmt.Printf("2nd element: %s (%s)\n",string(data),time.Unix(0,timestamp))
    }

    timestamps,datas,err := x.ReadBatch(0,3)
    if (err != nil) {
        fmt.Printf("ReadBatch error\n")
    } else {
        for i := 0; i < len(timestamps); i++ {
            fmt.Printf("%d: %s (%s)\n",i,string(datas[i]),time.Unix(0,timestamps[i]))
        }
    }

    fmt.Printf("Size: %d\n",x.Len())
    time.Sleep(5*time.Second)

    fmt.Printf("Size: %d\n",x.Len())
    time.Sleep(5*time.Second)
    fmt.Printf("Size: %d\n",x.Len())


}
