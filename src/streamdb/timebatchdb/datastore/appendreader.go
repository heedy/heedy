package datastore

import (
    "os"
    "path"
    "strconv"
    )

type AppendReader struct {
    appendf *os.File        //The appendfile is a binary blob dump of the data
}

func (a *AppendReader) Close() {
    a.appendf.Close()
}

//The size of the entire append file
func (a *AppendReader) Size() int64 {
    st,err := a.appendf.Stat()
    if (err != nil) {
        return 0
    }
    return st.Size()
}

//Returns the next datapoint from the append file.
func (a *AppendReader) Next() (d KeyedDatapoint,err error) {
    d,err = ReadKeyedDatapoint(a.appendf)

    return d,err

}

//Resets back to the beginning of the file
func (a *AppendReader) Reset() {
    a.appendf.Seek(0,0)
}

//Sets the cursor at the end of the file (ie, at the most recent datapoint)
//Presumably because there will be further appends that are to be read
func (a *AppendReader) ToEnd() {
    a.appendf.Seek(0,2)
}


//New append reader. It slurps the file, and can play back all messages
func NewAppendReader(fpath string,filenum int64) (*AppendReader,error) {
    if err := MakeDirs(fpath); err!= nil {
        return nil,err
    }

    afile,err := os.OpenFile(path.Join(fpath,"append."+strconv.FormatInt(filenum,10)), os.O_RDONLY|os.O_CREATE, 0666)
    if (err != nil) {
        return nil,err
    }

    return &AppendReader{afile},nil
}
