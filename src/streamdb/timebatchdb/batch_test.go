package timebatchdb

import (
    "testing"
    )

func TestBatch(t *testing.T) {
    bw := NewBatchWriter()

    if (bw.Len()!= 0 || bw.Size() != 0) {
        t.Errorf("Incorrect Length")
        return
    }

    bw.Insert(1000,[]byte("Hello World!"))
    if (bw.Len()!= 1 || bw.Size() != 12) {
        t.Errorf("Incorrect Length")
        return
    }

    bw.Insert(2000,[]byte("Hello World!2"))
    if (bw.Len()!= 2 || bw.Size() != 25) {
        t.Errorf("Incorrect Length")
        return
    }

    //Now read the write buffer right into a BatchReader
    br := NewBatchReader(bw.IndexBuffer.Bytes(),bw.DataBuffer.Bytes())

    if (br.Len()!= 2 || br.Size() != 25) {
        t.Errorf("Incorrect Lengths")
        return
    }

    if (string(br.Data[0])!="Hello World!") {
        t.Errorf("Data decode failure: %s",string(br.Data[0]))
    }
    if (string(br.Data[1])!="Hello World!2") {
        t.Errorf("Data decode failure: %s",string(br.Data[1]))
    }

    i,err := br.FindTime(500)
    if (i!=0 || err != nil) {
        t.Errorf("FindTime Failed: %s %d",err,i)
    }

    i,err = br.FindTime(1000)
    if (i!=1 || err != nil) {
        t.Errorf("FindTime Failed: %s %d",err,i)
    }

    i,err = br.FindTime(2000)
    if (i!=2 || err == nil) {
        t.Errorf("FindTime Failed: %s %d",err,i)
    }


}

func TestBatchRange(t *testing.T) {
    bw := NewBatchWriter()

    bw.Insert(1000,[]byte("test"))
    bw.Insert(1500,[]byte("test"))
    bw.Insert(2000,[]byte("test0"))
    bw.Insert(2000,[]byte("test1"))
    bw.Insert(2000,[]byte("test2"))
    bw.Insert(2500,[]byte("test3"))
    bw.Insert(3000,[]byte("test"))
    bw.Insert(3000,[]byte("test"))
    bw.Insert(3000,[]byte("test"))


    reader := NewBatchReader(bw.IndexBuffer.Bytes(),bw.DataBuffer.Bytes())

    if (reader.Timestamps[0]!=1000 || len(reader.Timestamps)!=9) {
        t.Errorf("error reading\n")
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

    ts,d,err := reader.GetRange(1900,2800)

    if (err != nil || len(ts) != 4 || len(d) != 4 || ts[0] != 2000 || ts[3]!=2500 || string(d[0]) != "test0" || string(d[3]) != "test3") {
        t.Errorf("Fails to get array range")
    }

    ts,d,err = reader.GetRange(500,1500)
    if (err != nil || len(ts) != 2 || len(d) != 2 || ts[0] != 1000 || string(d[0]) != "test") {
        t.Errorf("Fails on start array")
    }

    ts,d,err = reader.GetRange(2900,3000)
    if (err != nil || len(ts) != 3 || len(d) != 3 || ts[0] != 3000 || string(d[0]) != "test") {
        t.Errorf("Fails on start array")
    }

    ts,d,err = reader.GetRange(3000,3100)
    if (err == nil) {
        t.Errorf("Fails on out of range array")
    }
    ts,d,err = reader.GetRange(300,800)
    if (err == nil) {
        t.Errorf("Fails on out of range array")
    }

}
