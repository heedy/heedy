package timebatchdb

import (
    "os"
    "path/filepath"
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
