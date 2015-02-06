package timebatchdb

import (
    "testing"
    )

func TestMongoStore(t *testing.T) {
    m,err := OpenMongoStore("localhost","testdb1")
    if (err!=nil) {
       t.Errorf("Couldn't open MongoStore")
       return
    }
    defer m.Close()

    //First drop the collection - so that tests are fresh
    m.DropCollection("0")

    //First check returning empties

    i,err := m.GetEndIndex("hello/world")
    if err!=nil || i!=0 {
        t.Errorf("EndIndex of empty failed: %d %s",i,err)
        return
    }

    r := m.GetTime("hello/world",0)
    defer r.Close()
    if (r.Next()!=nil) {
        t.Errorf("Empty datapoint returned for time")
        return
    }

    r = m.GetIndex("hello/world",0)
    defer r.Close()
    if (r.Next()!=nil) {
        t.Errorf("Empty datapoint returned for index")
        return
    }

    //Now insert some data
    timestamps := []uint64{1,2,3,4,5,6,6,7,8}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    err = m.Append("hello/world",CreateDatapointArray(timestamps[0:1],data[0:1]))
    if err != nil {
        t.Errorf("Error in append: %s",err)
        return
    }

    i,err = m.GetEndIndex("hello/world")
    if err!=nil || i!=1 {
        t.Errorf("EndIndex of nonempty failed: %d %s",i,err)
        return
    }
    i,err = m.GetEndIndex("hello/world2")
    if err!=nil || i!=0 {
        t.Errorf("EndIndex of empty failed: %d %s",i,err)
        return
    }

    //TIME

    r = m.GetTime("hello/world",0)
    defer r.Close()
    dp := r.Next()
    if (dp==nil || dp.Timestamp()!=1) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for time")
        return
    }

    r = m.GetTime("hello/world",1)
    defer r.Close()
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for time")
        return
    }

    //INDEX

    r = m.GetIndex("hello/world",0)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=1) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for index")
        return
    }

    r = m.GetIndex("hello/world",1)
    defer r.Close()
    if (r.Next()!=nil) {
        t.Errorf("incorrect data for index")
        return
    }

    err = m.Append("hello/world",CreateDatapointArray(timestamps[1:3],data[1:3]))
    if err != nil {
        t.Errorf("Error in append: %s",err)
        return
    }
    i,err = m.GetEndIndex("hello/world")
    if err!=nil || i!=3 {
        t.Errorf("EndIndex of nonempty failed: %d %s",i,err)
        return
    }

    r = m.GetTime("hello/world",0)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=1) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=2) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for time")
        return
    }

    r = m.GetTime("hello/world",2)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for time")
        return
    }

    r = m.GetIndex("hello/world",0)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=1) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=2) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for index")
        return
    }

    r = m.GetIndex("hello/world",2)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("incorrect data for index")
        return
    }

    err = m.Append("hello/world",CreateDatapointArray(timestamps[3:],data[3:]))
    if err != nil {
        t.Errorf("Error in append: %s",err)
        return
    }
    i,err = m.GetEndIndex("hello/world")
    if err!=nil || i!=9 {
        t.Errorf("EndIndex of nonempty failed: %d %s",i,err)
        return
    }

    r = m.GetTime("hello/world",4)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("incorrect data for time")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("incorrect data for time")
        return
    }
    r.Close()

    r = m.GetIndex("hello/world",4)
    defer r.Close()
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("incorrect data for index")
        return
    }
    dp = r.Next()
    if (dp==nil || dp.Timestamp()!=6) {
        t.Errorf("incorrect data for index")
        return
    }
    r.Close()
}
