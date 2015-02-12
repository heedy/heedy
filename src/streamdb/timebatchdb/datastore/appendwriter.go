package datastore

import (
    "os"
    "path"
    "bytes"
    "strconv"
    "sync"
    "time"
    )

type AppendWriter struct {
    appendf *os.File        //The appendfile is a binary blob dump of the data

    //There are 2 buffers - one is asynchronously written to file, the other is written by messages
    buffer1 *bytes.Buffer
    buffer2 *bytes.Buffer

    buflock sync.Mutex  //The buffer lock is used to make sure that buffers are not switched while things are written
    bufnum int64        //The number of buffer to write
}

//This must be called beofre shutting down not to lose data.
func (a *AppendWriter) Close() {
    a.appendf.Close()
}

//The size of the entire append file
func (a *AppendWriter) Size() int64 {
    st,err := a.appendf.Stat()
    if (err != nil) {
        return 0
    }
    return st.Size()
}

//Returns the length in bytes of the buffer that data is being added to
func (a *AppendWriter) Len() int {
    //We don't need to lock it, since we're not writing the buffer, we're just glancing at its length
    buf := a.buffer1
    if (a.bufnum%2==1) {
        buf = a.buffer2
    }

    return buf.Len()
}

//Flips the buffer, and writes the current buffer to file.
func (a *AppendWriter) FlipBuffer() {
    //Lock the buffer, and flip the bufnum
    a.buflock.Lock()
    a.bufnum++
    a.buflock.Unlock()
}


//Writes the "pong" buffer to file. The pong buffer is the buffer that is currently not being written to
func (a *AppendWriter) WriteFile() error {
    //Choose the correct buffer. The opposite of the buffer used for writebuf
    buf := a.buffer1
    if (a.bufnum%2==0) {
        buf = a.buffer2
    }

    _,err := a.appendf.Write(buf.Bytes())
    buf.Reset() //We want the buffer to be empty again

    return err
}

//Writes the given stuff to the buffer
func (a *AppendWriter) WriteBuffer(d KeyedDatapoint) {
    a.buflock.Lock()
    //Choose the correct buffer. The opposite of the buffer used for writebuf
    buf := a.buffer1
    if (a.bufnum%2==1) {
        buf = a.buffer2
    }

    buf.Write(d.Bytes())

    a.buflock.Unlock()
}

//The function which is to be run asynchronously, which flips the buffer and writes the file. It can optionally
//wait the given number of milliseconds before retrying if there is nothing to write.
func (a *AppendWriter) FlipWrite(waitmillis int64) error {
    if (a.Len()==0) {
        if (waitmillis == 0) {
            return nil
        }

        time.Sleep(time.Duration(waitmillis) * time.Millisecond)
        if (a.Len()==0) {
            return nil //Even after sleeping we don't have anything to write
        }
    }

    a.FlipBuffer()
    err := a.WriteFile()
    return err
}

//New appender writer -
func NewAppendWriter(fpath string,filenum int64) (*AppendWriter,error) {
    if err := MakeDirs(fpath); err!= nil {
        return nil,err
    }

    afile,err := os.OpenFile(path.Join(fpath,"append."+strconv.FormatInt(filenum,10)), os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0666)
    if (err != nil) {
        return nil,err
    }

    return &AppendWriter{afile,new(bytes.Buffer),new(bytes.Buffer),sync.Mutex{},0},nil
}
