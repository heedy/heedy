package storagedb_test

import (
    "connectordb/storagedb"
    "time"
    "testing"
    )

func TestReadWrite(t *testing.T) {
    if (storagedb.PathExists("./data_test") == true) {
        storagedb.Delete("./data_test")
        storagedb.Delete("./data_test.data")
    }


    writer,err := storagedb.GetWriter("./data_test")
    if (err!=nil) {
        t.Errorf("Writer error: %s\n",err)
        return
    }
    defer writer.Close()

    reader,err := storagedb.GetReader("./data_test")
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

    if (storagedb.PathExists("./data_test") != true) {
        t.Errorf("The path DOES exist, foo\n")
    }

}

func BenchmarkWrite(b *testing.B) {
    writer,err := storagedb.GetWriter("./data_test")
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
    writer,err := storagedb.GetWriter("./data_test")
    if (err!=nil) {
        b.Errorf("Writer error: %s\n",err)
        return
    }
    defer writer.Close()

    for i:= 0; i<b.N; i++ {
        writer.BatchInsertNow([]byte("Hello World! Testing testing 1 2 3 Blah blah"))
    }
    writer.BatchWrite()

    reader,err := storagedb.GetReader("./data_test")
    if (err!=nil) {
        b.Errorf("Reader error: %s\n",err)
        return
    }
    defer reader.Close()

    b.ResetTimer()

    for i:=0; i< b.N;i++ {
        reader.Read(int64(i))
    }
}
