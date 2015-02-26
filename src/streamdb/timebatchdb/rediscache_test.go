package timebatchdb

import (
    "testing"
    )

func TestRedisCache(t *testing.T) {

    rc,err := OpenRedisCache("localhost:6379")
    if err!= nil {
        t.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the database
    rc.Clear()

    if (rc.Len("hello")!=0) {
        t.Errorf("Redis gives wrong length %v",rc.Len("hello"))
        return
    }

    rc.InsertOne("hello",NewDatapoint(1337,[]byte("Hello World!")))

    if (rc.Len("hello")!=1) {
        t.Errorf("Redis gives wrong length %v",rc.Len("hello"))
        return
    }
    if (rc.Len("hello/wrld")!=0) {
        t.Errorf("Redis gives wrong length %v",rc.Len("hello"))
        return
    }

    rc.InsertOne("hello2",NewDatapoint(1338,[]byte("test2")))
    rc.InsertOne("hello2",NewDatapoint(1339,[]byte("test3")))
    rc.InsertOne("hello2",NewDatapoint(1340,[]byte("test4")))

    if (rc.Len("hello2")!=3) {
        t.Errorf("Redis gives wrong length %v",rc.Len("hello2"))
        return
    }

    dr := rc.Get("hello2")
    dr.Init()

    dp := dr.Next()
    if (dp==nil || dp.Timestamp()!=1338 || string(dp.Data())!="test2") {
        t.Errorf("DataRange incorrect %v",dp)
    }
    dp = dr.Next()
    if (dp==nil || dp.Timestamp()!=1339 || string(dp.Data())!="test3") {
        t.Errorf("DataRange incorrect %v",dp)
    }
    dp = dr.Next()
    if (dp==nil || dp.Timestamp()!=1340 || string(dp.Data())!="test4") {
        t.Errorf("DataRange incorrect %v",dp)
    }
    dp = dr.Next()
    if (dp!=nil) {
        t.Errorf("DataRange endtime incorrect")
    }

    //Clean the database
    rc.Clear()

}
