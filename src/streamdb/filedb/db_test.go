package filedb_test

import (
    "connectordb/streamdb/filedb"
    "time"
    "testing"
    )

func TestFileDB(t *testing.T) {
    if (filedb.PathExists("./testdb") == true) {
        filedb.Delete("./testdb")
    }

    db,err := filedb.FileDatabase("./testdb")

    if err != nil {
        t.Errorf("Database Error: %s\n",err)
        return
    }

    if (db.Exists("user1/stream1")) {
        t.Errorf("Existence of nonexisting\n")
    }


    writer,err := db.Writer("user1/stream1")
    if (err!=nil) {
        t.Errorf("Writer error: %s\n",err)
        return
    }
    defer writer.Close()

    if (db.Exists("/user1/stream1") == false) {
        t.Errorf("Nonexistence of existing\n")
    }

    reader,err := db.Reader("user1/stream1")
    if (err!=nil) {
        t.Errorf("Reader error: %s\n",err)
        return
    }

    defer reader.Close()

    if (reader.Len()!=0) {
        t.Errorf("Reader length nonzero\n")
    }


    _,_,err = reader.Read(0)
    if (err == nil) {
        t.Errorf("No error reading empty\n")
    }
    _,_,err = reader.ReadBatch(0,1)
    if (err == nil) {
        t.Errorf("No error reading empty batch\n")
    }

    writer.BatchInsert(time.Now().UnixNano(),[]byte("Hello0"))

    _,_,err = reader.Read(0)
    if (err == nil) {
        t.Errorf("No error reading uncommitted\n")
    }
    _,_,err = reader.ReadBatch(0,1)
    if (err == nil) {
        t.Errorf("No error reading uncommitted batch\n")
    }

    writer.BatchInsertNow([]byte("HelloThere1"))
    writer.BatchInsertNow([]byte("HelloWrld2"))
    writer.BatchInsertNow([]byte("HelloWorld3"))
    writer.BatchWrite()

    _,data,err := reader.Read(1)
    if (err!=nil || string(data)!="HelloThere1") {
        t.Errorf("Error reading\n")
    }
    timestamps,datas,err := reader.ReadBatch(1,3)
    if (err != nil || len(timestamps) != 2 || len(datas)!=2) {
        t.Errorf("Error reading batch\n")
    }
    if (string(datas[0])!="HelloThere1" || string(datas[1])!="HelloWrld2") {
        t.Errorf("Data batch read incorrectly\n")
    }

    if (reader.Len()!=4) {
        t.Errorf("Reader length incorrect\n")
    }

}

func BenchmarkWrite(b *testing.B) {
    writer,err := filedb.GetWriter("./testdb/database")
    if (err!=nil) {
        b.Errorf("Writer error: %s\n",err)
        return
    }
    defer writer.Close()

    for i:=0; i< b.N;i++ {
        writer.BatchInsertNow([]byte("Hello World! Testing testing 1 2 3 Blah blah"))
        writer.BatchWrite()
    }
}

func BenchmarkRead(b *testing.B) {
    writer,err := filedb.GetWriter("./testdb/database")
    if (err!=nil) {
        b.Errorf("Writer error: %s\n",err)
        return
    }
    defer writer.Close()

    for i:= 0; i<b.N; i++ {
        writer.BatchInsertNow([]byte("Hello World! Testing testing 1 2 3 Blah blah"))
    }
    err = writer.BatchWrite()
    if (err!= nil) {
        b.Errorf("BatchWrite error: %s\n",err)
    }


    reader,err := filedb.GetReader("./testdb/database")
    if (err!=nil) {
        b.Errorf("Reader error: %s\n",err)
        return
    }
    defer reader.Close()

    b.ResetTimer()

    for i:=0; i< b.N;i++ {
        _,_,err = reader.Read(int64(i))
        if (err!= nil) {
            b.Errorf("Reader error: %s\n",err)
        }
    }
}
