package filedb_test

import (
    "connectordb/streamdb/filedb"
    "testing"
    )

func TestTimestamp(t *testing.T) {
    if (filedb.PathExists("./testdb") == true) {
        filedb.Delete("./testdb")
    }

    db,err := filedb.FileDatabase("./testdb")

    if err != nil {
        t.Errorf("Database Error: %s\n",err)
        return
    }

    writer,err := db.Writer("user1/stream1")
    if (err!=nil) {
        t.Errorf("Writer error: %s\n",err)
        return
    }
    defer writer.Close()

    reader,err := db.Reader("user1/stream1")
    if (err!=nil) {
        t.Errorf("Reader error: %s\n",err)
        return
    }

    defer reader.Close()

    writer.BatchInsert(1000,[]byte("test"))
    writer.BatchInsert(1500,[]byte("test"))
    writer.BatchInsert(2000,[]byte("test0"))
    writer.BatchInsert(2000,[]byte("test1"))
    writer.BatchInsert(2000,[]byte("test0"))
    writer.BatchInsert(2500,[]byte("test2"))
    writer.BatchInsert(3000,[]byte("test"))
    writer.BatchInsert(3000,[]byte("test"))
    writer.BatchInsert(3000,[]byte("test"))
    writer.BatchWrite()

    ts,err := reader.ReadTimestamp(0)
    if (err != nil || ts != 1000) {
        t.Errorf("error reading\n")
    }

    ts,err = reader.ReadTimestamp(9)
    if (err == nil) {
        t.Errorf("no error reading past boundary\n")
    }

    i,err := reader.FindTime(1200)
    if (err != nil || i != 1) {
        t.Errorf("Error in findtime: %s %d",err,i)
    }

    i,err = reader.FindTime(2000)
    if (err != nil || i != 5) {
        t.Errorf("Error in findtime: %s %d",err,i)
    }

    i,err = reader.FindTime(3000)
    if (err==nil || i != 9) {
        t.Errorf("Does not return 'out of bounds': %s %d",err,i)
    }

    i1,i2,err := reader.FindTimeRange(1300,2900)
    if (i1 != 1 || i2 != 6 || err != nil) {
        t.Errorf("Incorrect range: %s %d %d",err,i1,i2)
    }

    i1,i2,err = reader.FindTimeRange(1300,3000)
    if (i1 != 1 || i2 != 9 || err == nil) {
        t.Errorf("Incorrect range: %s %d %d",err,i1,i2)
    }
    i1,i2,err = reader.FindTimeRange(3000,3500)
    if (i1 != 9 || i2 != -1 || err == nil) {
        t.Errorf("Incorrect range: %s %d %d",err,i1,i2)
    }
}
