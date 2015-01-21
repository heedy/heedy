package timebatchdb

import (
    "os"
    "errors"
    "bytes"
    "encoding/binary"
    )



type IndexPage struct {
    Dataloc []uint64
    Batchloc []uint64
    Timestamp []uint64
    Datanum []uint64
    Previndex []uint64
}

//Returns the number of elements that the IndexPage holds
func (ip *IndexPage) Len() int {
    return len(ip.Datanum)
}

func (ip *IndexPage) GetLocs(i int) (dataindex uint64,datasize int,offsetindex uint64, offsetsize int) {
    return ip.Dataloc[i],int(ip.Dataloc[i+1]-ip.Dataloc[i]),ip.Batchloc[i],int(ip.Batchloc[i+1]-ip.Batchloc[i])
}

//Given the index file, an index at which to start, and the number of entries to read,
//returns the IndexPage with the decoded data
func GetIndexPage(file *os.File, startindex int64, pagesize int) (ip *IndexPage,err error){

    //There are 5 elements to each, and the dataloc and batchloc is shared with next one
    pagebuf := make([]byte,int(IndexElementSize)*pagesize+2*8)
    numread,_ := file.ReadAt(pagebuf,startindex*int64(IndexElementSize))

    //The file might end before
    pagesize = numread/int(IndexElementSize)

    if (pagesize <= 0) {
        return nil,errors.New("Error reading page")
    }

    buf := bytes.NewReader(pagebuf)

    dataloc := make([]uint64,pagesize+1)
    batchloc := make([]uint64,pagesize+1)
    endtime := make([]uint64,pagesize)
    endindex := make([]uint64,pagesize)
    previndex := make([]uint64,pagesize)

    //Now decode the buffer into memory
    for i := 0; i < pagesize; i++ {
        binary.Read(buf,binary.LittleEndian,&dataloc[i])
        binary.Read(buf,binary.LittleEndian,&batchloc[i])
        binary.Read(buf,binary.LittleEndian,&endtime[i])
        binary.Read(buf,binary.LittleEndian,&endindex[i])
        binary.Read(buf,binary.LittleEndian,&previndex[i])
    }
    binary.Read(buf,binary.LittleEndian,&dataloc[pagesize])
    binary.Read(buf,binary.LittleEndian,&batchloc[pagesize])

    return &IndexPage{dataloc,batchloc,endtime,endindex,previndex},nil
}
