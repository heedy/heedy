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
func (kw *IndexReader) Close() {
    kw.indexfile.Close()
    kw.offsetf.Close()
    kw.dataf.Close()
}


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
