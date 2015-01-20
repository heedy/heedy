package timebatchdb

import (
    "bytes"
    "encoding/binary"
    "os"
    "sync"
    )

const indexElementSize uint64 = 8*3

type KeyWriterKey struct {
    bw *BatchWriter         //The writer buffer
    PrevFileIndex uint64    //The index of the most recently written batch of the key
    KeyPoints uint64        //The "index" of the key itself, meaning the number of data points written thus far
}

type KeyWriter struct {
    keyfile *os.File    //The file in which the keys and links to batches are stored
    offsetf *os.File     //File where batch offsets and timestamps are stored
    dataf *os.File      //File where data is stored
    keybuf *bytes.Buffer        //A buffer where keys are written to before dumping to file
    keys map[uint64](*KeyWriterKey)    //The map for all keys
    batchnum uint64                 //The number of batches written to the file
    datafsize uint64        //The size of the data file
    offsetfsize uint64      //The size of the offsetfile

}

func NewKeyWriterKey(previndex uint64, keypoints uint64) {
    return &KeyWriterKey{NewBatchWriter(),previndex,keypoints}
}

//Closes all open files in keyWriter
func (kw *KeyWriter) Close() {
    kw.keyfile.Close()
    kw.offsetf.Close()
    kw.dataf.Close()
}

//Given a key number, write the currently buffered batch to the file, and start a new batch.
func (kw *KeyWriter) Write(uint64 key) (err error) {
    kwk, ok := kw.keys[key]
    if (ok == false) {
        return errors.New("Unrecognized Key")
    }
    return kw.WriteKey(key,kwk)
}

//Given a key number, returns the BatchWriter associated with it.
func (kw *KeyWriter) GetKeyBatch(uint64 key) (batch *BatchWriter,err error) {
    kwk, ok := kw.keys[key]
    if (ok == false) {
        return nil,errors.New("Unrecognized Key")
    }
    return kwk.bw,nil
}

//Allows to use keys not in the KeyWriter's keymap
func (kw *KeyWriter) WriteKey(uint64 keynum, key *KeyWriterKey) (err error) {
    numwrite := key.bw.Len()
    if (numwrite==0) {
        return nil //Don't waste time of empty batches
    }

    key.bw.Lock()

    dataw,err := kw.dataf.Write(key.bw.DataBuffer.Bytes())
    if (err != nil) {
        key.bw.Unlock()
        return err  //The database might now be corrupted
    }

    offsetw,err := kw.offsetf.Write(key.bw.IndexBuffer.Bytes())
    if (err != nil) {
        key.bw.Unlock()
        return err //The database might be corrupted
    }

    key.bw.Clear()
    key.bw.Unlock()

    kw.datafsize += uint64(dataw)
    kw.offsetfsize += uint64(offsetw)

    key.KeyPoints += numwrite

    //Format is (dataloc,batchloc,endtime,endindex,previndex)
    //Except we write 2 offset, since dataloc and batchloc were written by previous iteration
    binary.Write(kw.keybuf,binary.LittleEndian,key.bw.LastTime)
    binary.Write(kw.keybuf,binary.LittleEndian,key.KeyPoints)
    binary.Write(kw.keybuf,binary.LittleEndian,key.PrevFileIndex)
    binary.Write(kw.keybuf,binary.LittleEndian,kw.datafsize)
    binary.Write(kw.keybuf,binary.LittleEndian,kw.offsetfsize)

    //The previndex is the current index in terms of batchnumber
    key.PrevFileIndex = kw.batchnum

    kw.batchnum += 1

}

//Writes the necessary stuff to make the index file consistent with data
func (kw *KeyWriter) Flush() (err error) {
    indexw,err := kw.keyfile.Write(kw.keybuf.Bytes())
    if (err != nil) {
        return err //The database might be corrupted
    }
    kw.keybuf.Reset()
    return nil
}

//Writes all keys and batches that are non-empty to file. This should be used
//  before shutting down the database.
func (kw *KeyWriter) Dump() (err error) {

    //First dump all batches
    for key,kwk := range kw.keys {
        kw.WriteKey(key,kwk)
    }
    //Next, flush the database.
    kw.Flush()
}

func (kw *KeyWriter) AddKey(keyid uint64,key *KeyWriterKey) {
    kw.keys[keyid] = key
}

//Opens the KeyWriter given a relative path of the key index
func NewKeyWriter(path) (kw *KeyWriter, err error){
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

    //Open the keyfile
    keyf,err := os.OpenFile(path + ".index", os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        dataf.Close()
        offsetf.Close()
        return nil,err
    }

    datastat, err := dataf.Stat()
    if (err != nil) {
        keyf.Close()
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    offsetstat, err := offsetf.Stat()
    if (err != nil) {
        keyf.Close()
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    keystat, err := keyf.Stat()
    if (err != nil) {
        keyf.Close()
        offsetf.Close()
        dataf.Close()
        return nil,err
    }

    batchnum := keystat.Size()/indexElementSize

    if (keystat.Size()==0) {
        //The index file is empty - add 2 0s to it which are the dataloc and batchloc of first batch
        binary.Write(keyf,binary.LittleEndian,uint64(0))
        binary.Write(keyf,binary.LittleEndian,uint64(0))
    }

    return &KeyWriter{keyf,offsetf,dataf,new(bytes.Buffer),make(map[uint64](*KeyWriterKey)),batchnum,datastat.Size(),offsetstat.Size()},nil
}
