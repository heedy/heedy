package timebatchdb

import (
    "os"
    "path/filepath"
)

func PathExists(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true
    }
    return false
}

func MakeParentDirs(path string) (err error) {
    //Check if the directory exists
    parentdir := filepath.Dir(path)
    if PathExists(parentdir) == false {
        err = os.MkdirAll(parentdir,0777)
        if (err != nil) {
            return err
        }
    }
    return nil
}

func DataSize(path string) (int64,error) {
    s, err := os.Stat(path)
    if (err != nil) {
        return 0,err
    }
    return s.Size(),nil
}

func Delete(path string) error {
    return os.RemoveAll(path)
}
