package timebatchdb

import (
    "bytes"
    "encoding/binary"
    "os"
    )

const indexElementSize uint64 = 8*3

type IndexWriterBatch struct {
    Bw *BatchWriter         //The writer buffer
    PrevFileIndex uint64    //The index of the most recently written batch of this type (key)
    KeyPoints uint64      //The "index" of the key itself, meaning the number of data points written thus far
}

type IndexWriter struct {
    indexf *os.File    //The file in which the keys and links to batches are stored
    offsetf *os.File     //File where batch offsets and timestamps are stored
    dataf *os.File      //File where data is stored
    indexbuf *bytes.Buffer        //A buffer where index is written to before dumping to file
    batchnum uint64                 //The number of batches written to the file
    datafsize uint64        //The size of the data file
    offsetfsize uint64      //The size of the offsetfile

}

//The batchindex of the most recently written batch of this key and the number of datapoints total
//of the batch's key.
func NewIndexWriterBatch(previndex uint64, keypoints uint64) (*IndexWriterBatch) {
    return &IndexWriterBatch{NewBatchWriter(),previndex,keypoints}
}

//Closes all open files in IndexWriter
func (iw *IndexWriter) Close() {
    iw.indexf.Close()
    iw.offsetf.Close()
    iw.dataf.Close()
}

//Given the batch object, writes it to file (the index is not flushed. Use Flush() to flush file)
func (iw *IndexWriter) Write(key *IndexWriterBatch) (err error) {
    numwrite := key.Bw.Len()
    if (numwrite==0) {
        return nil //Don't waste time of empty batches
    }

    key.Bw.Lock()

    dataw,err := iw.dataf.Write(key.Bw.DataBuffer.Bytes())
    if (err != nil) {
        key.Bw.Unlock()
        return err  //The database might now be corrupted
    }

    offsetw,err := iw.offsetf.Write(key.Bw.IndexBuffer.Bytes())
    if (err != nil) {
        key.Bw.Unlock()
        return err //The database might be corrupted
    }

    key.Bw.Clear()
    key.Bw.Unlock()

    iw.datafsize += uint64(dataw)
    iw.offsetfsize += uint64(offsetw)

    key.KeyPoints += uint64(numwrite)

    //Format is (dataloc,batchloc,endtime,endindex,previndex)
    //Except we write 2 offset, since dataloc and batchloc were written by previous iteration
    binary.Write(iw.indexbuf,binary.LittleEndian,key.Bw.LastTime)
    binary.Write(iw.indexbuf,binary.LittleEndian,key.KeyPoints)
    binary.Write(iw.indexbuf,binary.LittleEndian,key.PrevFileIndex)
    binary.Write(iw.indexbuf,binary.LittleEndian,iw.datafsize)
    binary.Write(iw.indexbuf,binary.LittleEndian,iw.offsetfsize)

    //The previndex is the current index in terms of batchnumber
    key.PrevFileIndex = iw.batchnum

    iw.batchnum += 1

    return nil
}

//Writes the necessary stuff to make the index file consistent with data
func (iw *IndexWriter) Flush() (err error) {
    _,err = iw.indexf.Write(iw.indexbuf.Bytes())
    if (err != nil) {
        return err //The database might be corrupted
    }
    iw.indexbuf.Reset()
    return nil
}

//Opens the IndexWriter given a relative path of the key index
func NewIndexWriter(path string) (kw *IndexWriter, err error){
    if err = MakeParentDirs(path); err!= nil {
        return nil,err
    }

    //Opens offset and data file for append
    offsetf,err := os.OpenFile(path + ".offsets", os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        return nil,err
    }
    dataf,err := os.OpenFile(path + ".data", os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        offsetf.Close()
        return nil,err
    }

    //Open the indexf
    indexf,err := os.OpenFile(path + ".index", os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        dataf.Close()
        offsetf.Close()
        return nil,err
    }

    datastat, err := dataf.Stat()
    if (err != nil) {
        indexf.Close()
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    offsetstat, err := offsetf.Stat()
    if (err != nil) {
        indexf.Close()
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    indexstat, err := indexf.Stat()
    if (err != nil) {
        indexf.Close()
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    batchnum := uint64(indexstat.Size())/indexElementSize

    if (indexstat.Size()==0) {
        //The index file is empty - add 2 0s to it which are the dataloc and offsetloc of first batch
        binary.Write(indexf,binary.LittleEndian,uint64(0))
        binary.Write(indexf,binary.LittleEndian,uint64(0))
    }

    return &IndexWriter{indexf,offsetf,dataf,new(bytes.Buffer),batchnum,uint64(datastat.Size()),uint64(offsetstat.Size())},nil
}
