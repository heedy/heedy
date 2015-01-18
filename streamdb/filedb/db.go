package filedb

import (
    "os"
    "path"
    "strings"
    "errors"
    )

type FileDB struct {
    startpath string   //The root path to the database
}

func FileDatabase(dbpath string) (db *FileDB, err error) {
    //Clean the path
    dbpath = path.Clean(dbpath)
    //Make sure the database path exists
    if PathExists(dbpath) == false {
        err = os.MkdirAll(dbpath,0777)
        if (err != nil) {
            return nil,err
        }
    }

    return &FileDB{dbpath},nil
}

//Check if the path is valid for the database. In particular, .data files are strictly prohibited.
//  also clean the path
func CheckPath(dbpath string) bool {
    if (strings.LastIndex(dbpath,".data") != -1) {
        return false
    }
    //We're not going to mess around with relative paths. If the clean version is different,
    //  then the user was trying to be tricky, and we don't like tricky.
    if (path.Clean(dbpath)!= dbpath) {
        return false
    }
    return true
}

func (db *FileDB) Writer(wpath string) (*DataWriter,error) {
    if (CheckPath(wpath)==false) {
        return nil,errors.New("Invalid data path")
    }
    return GetWriter(path.Join(db.startpath,wpath))
}

func (db *FileDB) Reader(rpath string) (*DataReader,error) {
    if (CheckPath(rpath)==false) {
        return nil,errors.New("Invalid data path")
    }
    return GetReader(path.Join(db.startpath,rpath))
}

//Whether or not the given path exists as a datastream
func (db *FileDB) Exists(rpath string) (bool) {
    if (CheckPath(rpath)==false) {
        return false
    }
    //The .data extension guarantees that there won't be any weirdness
    return PathExists(path.Join(db.startpath,rpath+".data"))
}
