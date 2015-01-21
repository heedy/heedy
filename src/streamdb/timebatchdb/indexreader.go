package timebatchdb

import (
    //"bytes"
    //"encoding/binary"
    "os"
    )

type KeyReaderKey struct {
    prevfileindex uint64    //The index of the most recently written batch of the key
    keypoints uint64        //The "index" of the key itself, meaning the number of data points written thus far in total
}

type KeyReader struct {
    keyfile *os.File    //The file in which the keys and links to batches are stored
    offsetf *os.File     //File where batch offsets and timestamps are stored
    dataf *os.File      //File where data is stored
    keys map[uint64](*KeyReaderKey)  //The map for all keys
}

//Closes all open files in keyReader
func (kw *KeyReader) Close() {
    kw.keyfile.Close()
    kw.offsetf.Close()
    kw.dataf.Close()
}


//The previndex is the batch number of the most recent batch with the same key. keypoints is
//the total number of datapoints written with this key. Ie, if there were an array of all the
//datapoints of the given key written thus far, keypoints would be the size of this array
func NewKeyReaderKey(previndex uint64, keypoints uint64) (*KeyReaderKey) {
    return &KeyReaderKey{previndex,keypoints}
}

//Opens the KeyWriter given a relative path of the datafiles (without extensions)
func NewKeyReader(path string) (kr *KeyReader, err error){
    if err = MakeParentDirs(path); err!= nil {
        return nil,err
    }

    //Opens offset and data file for append
    offsetf,err := os.OpenFile(path + ".offsets", os.O_RDONLY, 0666)
    if (err != nil) {
        return nil,err
    }
    dataf,err := os.OpenFile(path + ".data",os.O_RDONLY, 0666)
    if (err != nil) {
        offsetf.Close()
        return nil,err
    }

    //Open the keyfile
    keyf,err := os.OpenFile(path + ".index",os.O_RDONLY, 0666)
    if (err != nil) {
        dataf.Close()
        offsetf.Close()
        return nil,err
    }

    return &KeyReader{keyf,offsetf,dataf,make(map[uint64](*KeyReaderKey))},nil
}
