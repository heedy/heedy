package timebatchdb

import (
    "os"
    )

type IndexPage struct {
    dataloc []uint64
    batchloc []uint64
    endtime []uint64
    endindex []uint64
    previndex []uint64
}


func GetIndexPage(file *os.File, startindex uint64, datasize int) (*IndexPage){

    return &IndexPage{nil,nil,nil,nil,nil}
}
