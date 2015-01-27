package timebatchdb

import (
    //"bytes"
    //"encoding/binary"
    "os"
    "path"
    )


type IndexReader struct {
    indexfile *os.File    //The file in which the index is stored
    offsetf *os.File     //File where batch offsets and timestamps are stored
    dataf *os.File      //File where data is stored
    pages map[uint64](*IndexCache)  //The cache of the index and associated batches
    pagesize uint64         //The size of each IndexPage

}

//Closes all open files in IndexReader. Must be called to not leak memory
func (ir *IndexReader) Close() {
    ir.indexfile.Close()
    ir.offsetf.Close()
    ir.dataf.Close()
}


//Loads the given page index of the indexfile
func (ir *IndexReader) LoadPage(pagenum uint64) (page *IndexCache,err error) {
    page,err = GetIndexCache(ir.indexfile,pagenum*ir.pagesize,int(ir.pagesize))
    if err != nil {
        return nil,err
    }

    //Add page to the page cache
    ir.pages[pagenum] = page

    return page,nil
}

func (ir *IndexReader) ReadIndex(index uint64) (timestamp uint64, datanum uint64, previndex uint64) {
    //Read the page from page cache
    page, ok := ir.pages[index/ir.pagesize]

    //The page is not loaded yet
    if (ok==false) {
        page,_ = ir.LoadPage(index/ir.pagesize)
    }

    //Now, read the subindex from the page itself
    return page.GetIndexValues(int(index%ir.pagesize))


}

func (ir IndexReader) ReadBatch(index uint64) (timestamps []uint64, data [][]byte) {
    //Read the page from page cache
    page, ok := ir.pages[index/ir.pagesize]

    //The page is not loaded yet
    if (ok==false) {
        page,_ = ir.LoadPage(index/ir.pagesize)
    }

    batch,ok := page.GetBatch(int(index%ir.pagesize))

    if (ok==false) {
        batch = page.LoadBatch(int(index%ir.pagesize),ir.offsetf,ir.dataf)
    }

    return batch.Timestamps,batch.Data
}

//Returns the data and timestamps associated with a data range for the key, given the starting index of the latest batch
//of the correct type.
func (ir *IndexReader) ReadRange(startindex uint64, dataindex1 uint64, dataindex2 uint64) (timestamps []uint64,data [][]byte) {
    datanum := dataindex2-dataindex1

    //First, allocate the timestamp and data arrays, into which all data will be written
    timestamps = make([]uint64,datanum)
    data = make([][]byte,datanum)

    //Next, since we have a batch with and endindex out of our range, we move back batch by batch to find a batch which
    //contains dataindex2.
    _,dataindex,previndex := ir.ReadIndex(startindex)
    for ; dataindex > dataindex2; {
        if (previndex == startindex) {
            //
        }
        //_,dataindex,previndex := ir.ReadIndex(startindex)
    }


    return timestamps, data

}

//Returns the data and timestamps associated with a time range for the key, given the starting index of the latest batch
//of the correct type.
/*
func (ir *IndexReader) ReadTimeRange(startindex uint64, timestamp1 uint64, timestamp2 uint64) (timestamps []uint64,data [][]byte) {

}*/

//Opens the IndexReader given a relative path of the directory containing datafiles
func NewIndexReader(fpath string) (kr *IndexReader, err error){
    if err = MakeDirs(fpath); err!= nil {
        return nil,err
    }

    //Opens offset and data file for append
    offsetf,err := os.OpenFile(path.Join(fpath,"offsets"), os.O_RDONLY, 0666)
    if (err != nil) {
        return nil,err
    }
    dataf,err := os.OpenFile(path.Join(fpath,"data"),os.O_RDONLY, 0666)
    if (err != nil) {
        offsetf.Close()
        return nil,err
    }

    //Open the indexf
    indexf,err := os.OpenFile(path.Join(fpath,"index"),os.O_RDONLY, 0666)
    if (err != nil) {
        dataf.Close()
        offsetf.Close()
        return nil,err
    }

    //The default page size is 10 for testing. This will definitely have to increase a lot.
    return &IndexReader{indexf,offsetf,dataf,make(map[uint64](*IndexCache)),10},nil
}
