package storagedb

import (
    "os"
    )

func PathExists(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true
    }
    return false
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
