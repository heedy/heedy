package timebatchdb

import (
    "testing"
    )

func TestDatabase(t *testing.T) {
    m,err := OpenMongoStore("localhost","testdb")
    if (err!=nil) {
       t.Errorf("Couldn't open MongoStore")
       return
    }
    defer m.Close()

    //First drop the collection - so that tests are fresh
    m.DropCollection("0")

    rc,err := OpenRedisCache("localhost:6379")
    if (err!=nil) {
       t.Errorf("Couldn't open RedisCache")
       return
    }
    defer rc.Close()

    rc.Clear()

    db,err := Open("localhost:6379","localhost","testdb")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer db.Close()

    timestamps := []int64{1,2,3,4,5,6,3000,3100,3200}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    for i:=0;i<len(timestamps);i++ {
        err = db.Insert("user1/device1/stream1",timestamps[i],data[i])
        if err!=nil {
            t.Errorf("Insert Failed: %s",err)
        }
    }

    //Now check a data range by index, and then by timestamp
    r := db.GetTimeRange("user1/device1/stream2",0,1000)
    defer r.Close()
    dp:= r.Next()
    if (dp!=nil) {
        t.Errorf("Insert wrong key")
        return
    }

    //Now check a data range by index, and then by timestamp
    r = db.GetIndexRange("user1/device1/stream2",0,1000)
    defer r.Close()
    dp = r.Next()
    if (dp!=nil) {
        t.Errorf("Insert wrong key")
        return
    }

    r = db.GetTimeRange("user1/device1/stream1",2,5)
    defer r.Close()
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("time range get failed")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=4) {
        t.Errorf("time range get failed")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("time range get failed")
        return
    }
    dp= r.Next()
    if (dp!=nil) {
        t.Errorf("time range get failed")
        return
    }

    r = db.GetIndexRange("user1/device1/stream1",2,5)
    defer r.Close()
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=3) {
        t.Errorf("Index get failed")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=4) {
        t.Errorf("Index get failed")
        return
    }
    dp= r.Next()
    if (dp==nil || dp.Timestamp()!=5) {
        t.Errorf("Index get failed")
        return
    }
    dp= r.Next()
    if (dp!=nil) {
        t.Errorf("Index get failed")
        return
    }
    rc.Clear()
}
