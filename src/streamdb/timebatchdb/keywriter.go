package timebatchdb

/*

import (
    "os"
    "encoding/binary"
    "bytes"
    "errors"
    )

type KeyCacheWriter struct {
    kcf *os.File            //The file where the KeyCache is stored (append only)
    kcindexf *os.File       //The keyCache index file
    indexguarantee int64    //The index that is guaranteed to have been taken care of.
}


func (kcw *KeyCacheWriter) Close() {
    kcw.kcf.Close()
    kcw.kcindexf.Close()
}

func (kcw *KeyCacheWriter) LoadKeys() error {
    //If the index is non-empty, that means that we should load the most recent data from the keyCache
    kcindexstat, err := kcw.kcindexf.Stat()
    if (err!= nil) {
        return err
    }

    size := (kcindexstat.Size()-8)/16 //16 bytes are written each entry
    if (size == 0) {    //If there is no keycache yet, don't worry about it
        return nil
    }

    indexbuffer := make([]byte, 8*3)
    kcw.kcindexf.ReadAt(indexbuffer,2*8*(size-1))

    var startindex int64
    var endindex int64
    buf := bytes.NewReader(offsetbuffer)
    binary.Read(buf,binary.LittleEndian,&startindex)
    binary.Read(buf,binary.LittleEndian,&kcw.indexguarantee)
    binary.Read(buf,binary.LittleEndian,&endindex)

    if (startindex > endindex) {
        return errors.New("File Corrupted")
    }
    if (startindex == endindex) {   //There were no keys in this iteration
        return nil
    }

    //Now read the keycache
    numread := (endindex-startindex)
    keybuffer := make([]byte,numread*8*3)
    kcw.kcf.ReadAt(keybuffer,3*8*startindex)
    buf := bytes.NewReader(keybuffer)
    var key int64
    var keyindex int64
    var timestamp int64
    for i := int64(0); i< numread; i++ {
        binary.Read(buf,binary.LittleEndian,&key)
        binary.Read(buf,binary.LittleEndian,&keyindex)
        binary.Read(buf,binary.LittleEndian,&timestamp)

    }

}

func GetKeyCacheWriter(path string) (kcw *KeyCacheWriter, err error) {
    err = MakeParentDirs(path)
    if (err != nil) {
        return nil,err
    }

    //Opens the KeyCache for append
    kcfile,err := os.OpenFile(path+".keyCache", os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        return nil,err
    }

    //Opens the KeyCache Index for append
    kcindexf,err := os.OpenFile(path + ".keyIndex", os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        kcfile.Close()
        return nil,err
    }

    //If the keyCacheIndex is empty, write 0 as the location index (location,indexguarantee)
    if (offsetstat.Size() == 0) {
        binary.Write(offsetf,binary.LittleEndian, datastat.Size())
    }

    //Now, create the keycachewriter object
    kcw := &KeyCacheWriter{kcfile,kcindexf}

    err = kcw.LoadKeys()
    if (err != nil) {
        kcw.Close()
        return nil,err
    }

    return kcw,nil
}
*/
