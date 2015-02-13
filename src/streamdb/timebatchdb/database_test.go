package timebatchdb

import (
    "time"
    "testing"
    "streamdb/timebatchdb/datastore"
    )

func TestDatabase(t *testing.T) {
    m,err := datastore.OpenMongoStore("localhost","testdb")
    if (err!=nil) {
       t.Errorf("Couldn't open MongoStore")
       return
    }
    defer m.Close()

    //First drop the collection - so that tests are fresh
    m.DropCollection("0")

    //Turn on the DataStore writer
    go datastore.DataStoreWriter("localhost:4222","localhost","testdb", "testing/>")

    db,err := Open("localhost:4222","localhost","testdb")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer db.Close()

    //Wait for the DataStoreWriter to initialize
    time.Sleep(500 * time.Millisecond)

    //Now we test the database
    v := struct {K string; T int64; D string}{"testing/key1",1,"Hello World!"}

    err = db.Insert(v,"text","")
    if err!=nil {
        t.Errorf("Insert failed: %s",err)
        return
    }
    v.T=2
    v.D = "hi"
    v.K = "key2"
    err = db.InsertKey("testing/key1",v,"text","")
    if err!=nil {
        t.Errorf("Insert failed: %s",err)
        return
    }

    r := db.GetTimeRange("testing/randomkey","text",0,505785867)

    if r.Next()!=nil {
        t.Errorf("Get nonexisting failed")
        return
    }

    //Wait for the DataStoreWriter to write the earlier data
    time.Sleep(200 * time.Millisecond)

    r = db.GetIndexRange("testing/key1","text",0,50)
    defer r.Close()
    ok := r.UnmarshalNext(&v)
    if !ok || v.T!=1 || v.D!="Hello World!" {
        t.Errorf("Get incorrect datapoint %v %v",v,ok)
        return
    }
    ok = r.UnmarshalNext(&v)
    if !ok || v.T!=2 || v.D!="hi" {
        t.Errorf("Get incorrect datapoint %v %v",v,ok)
        return
    }
    if r.UnmarshalNext(&v) {
        t.Errorf("Got more than I bargained for")
        return
    }

}
