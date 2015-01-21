package timebatchdb

import (
    "bytes"
    "encoding/binary"
    "errors"
    )

type BatchReader struct {
    Timestamps []uint64  //The array of Timestamps in the batch
    Data [][]byte     //The byte arrays of Data associated with each timestamp
}

//Returns the number of datapoints in the batch
func (br *BatchReader) Len() int {
    return len(br.Timestamps)
}

//Returns the size in bytes of the data within the batch. This does not include timestamps.
func (br *BatchReader) Size() int {
    size := 0
    for i:= 0; i < len(br.Timestamps); i++ {
        size += len(br.Data[i])
    }
    return size
}

//Given a timestamp, finds the index within the batch of the element with timestamp greater than
//the given time. Note that times are assumed to be fuzzy.
func (br *BatchReader) FindTime(timestamp uint64) (index int, err error) {
    leftbound := 0

    //If the timestamp is earlier than the earliest Datapoint
    if (br.Timestamps[0] > timestamp) {
        return 0,nil
    }

    rightbound := br.Len()-1

    if (br.Timestamps[rightbound] <= timestamp) {
        return br.Len(),errors.New("Out of Range")
    }

    //Find the time in logn
    for (rightbound - leftbound > 1) {
        midpoint := (leftbound + rightbound)/2
        if (br.Timestamps[midpoint] <= timestamp) {
            leftbound = midpoint
        } else {
            rightbound = midpoint
        }
    }
    return rightbound, nil
}

func (br *BatchReader) GetRange(timestamp1 uint64, timestamp2 uint64) (timestamps []uint64, data [][]byte,err error) {
    index1,err := br.FindTime(timestamp1)
    if (err!=nil) {
        return nil,nil,err
    }
    index2,_ := br.FindTime(timestamp2)

    if (index2 == 0) {
        return nil,nil,errors.New("Out of Range")
    }

    return br.Timestamps[index1:index2],br.Data[index1:index2],nil
}

//Decodes the byte arrays of a batch into a BatchReader.
func NewBatchReader(index []byte, databytes []byte) (*BatchReader) {
    numread := (len(index)-8)/16

    timestamp := make([]uint64,numread)
    data := make([][]byte,numread)
    locs := make([]int64,numread+1) //The +1 is because there is one extra start location

    buf := bytes.NewReader(index)
    //Decode the index
    for i := 0; i < numread; i++ {
        binary.Read(buf,binary.LittleEndian,&locs[i])
        binary.Read(buf,binary.LittleEndian,&timestamp[i])
    }
    binary.Read(buf,binary.LittleEndian,&locs[numread])

    //Read data into byte arrays
    for i := 0; i < numread; i++ {
        data[i] = databytes[locs[i]:locs[i+1]]
    }

    return &BatchReader{timestamp,data}
}
