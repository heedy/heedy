package datastore

import (
    "os"
    "path/filepath"
    "errors"
    "io"
)

//Checks if the given path exists
func PathExists(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true
    }
    return false
}

//Given the path to a file, it checks if the parent directories exist, and if they don't, creates them.
func MakeParentDirs(path string) (err error) {
    //Check if the directory exists
    parentdir := filepath.Dir(path)
    return MakeDirs(parentdir)
}

//Given the path of a directory, makes sure the entire path exists, and creates it if not.
func MakeDirs(parentdir string) (err error) {
    if PathExists(parentdir) == false {
        err = os.MkdirAll(parentdir,0777)
        if (err != nil) {
            return err
        }
    }
    return nil
}

//Returns the size of the file pointed to by path
func DataSize(path string) (int64,error) {
    s, err := os.Stat(path)
    if (err != nil) {
        return 0,err
    }
    return s.Size(),nil
}


//
//For Fuck's sake, go, why do you make it so freaking annoying to read/write varints
//Most of the stuff here is copied straight from go source code and modified not to be fail

var overflow = errors.New("binary: varint overflows a 64-bit integer")

// ReadUvarint reads an encoded unsigned integer from r and returns it as a uint64.
func ReadUvarint(r io.Reader) (uint64, error) {
    var x uint64
    var s uint
    var b byte
    barr := make([]byte,1)

    for i := 0; ; i++ {
        _, err := r.Read(barr)
        if err != nil {
            return x, err
        }
        b = barr[0]
        if b < 0x80 {
            if i > 9 || i == 9 && b > 1 {
                return x, overflow
            }
            return x | uint64(b)<<s, nil
        }
        x |= uint64(b&0x7f) << s
        s += 7
    }
}


// WriteUvarint encodes a uint64 into writer and returns the number of bytes written.
// If the buffer is too small, PutUvarint will panic.
func WriteUvarint(buf io.Writer, x uint64) int {
    i := int(0)
    for x >= 0x80 {
        _,err := buf.Write([]byte{ byte(x) | 0x80})
        if err!=nil {
            return -1
        }
        x >>= 7
        i++
    }
    buf.Write([]byte{byte(x)})
    return i + 1
}
