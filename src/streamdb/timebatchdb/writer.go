package timebatchdb

import (
    "os"
    "path"
    "bytes"
    //"encoding/binary"
    "strconv"
    "sync"
    )

type AppendWriter struct {
    //The two files to write the datapoints
    appendf *os.File

    keymap *KeyMap  //The keymap for the append file

    //There are 2 buffers - one is asynchronously written to file, the other is written by messages
    buffer1 *bytes.Buffer
    buffer2 *bytes.Buffer

    buflock sync.Mutex  //The buffer lock is used to make sure that buffers are not switched while things are written

    filenum int64           //The number of file being written (strictly increasing)

    flipsize int64      //The size that the indexfile + datafile reach before switching write files
    waittime float32      //The time file writer sleeps before checking if new stuff to write

    fpath string            //The path of the database folder
}

//Switches the files
func (a *AppendWriter) switchfiles() error {

    //Opens index and data file for append
    appendf,err := os.OpenFile(path.Join(a.fpath,"append."+strconv.FormatInt(a.filenum,10)), os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        return err
    }
    a.appendf = appendf

    //Increment the filenum for the next file switch
    a.filenum++
    return nil
}


func (a *AppendWriter) Close() {
    a.appendf.Close()
}

func NewAppendWriter(fpath string,flipsize int64,waittime float32,filenum int64) (*AppendWriter,error) {
    if err := MakeDirs(fpath); err!= nil {
        return nil,err
    }

    return nil,nil
}
