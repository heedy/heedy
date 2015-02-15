package dtypes

import (
    "time"
    "testing"
    "streamdb/timebatchdb"
    )


func TestTypedRange(t *testing.T) {
    timestamps := []int64{1,2}
    datab := [][]byte{[]byte("test0"),[]byte("test1")}

    da := timebatchdb.CreateDatapointArray(timestamps,datab)

    dtype,ok := GetType("text")
    if (!ok) {
        t.Errorf("Type text not found")
        return
    }
    da.Init()
    tr := TypedRange{da,dtype}

    v := tr.Next()
    ts,_ := v.Timestamp()
    if v==nil || v.(*TextDatapoint).D!="test0" || ts!=1 {
        t.Errorf("Failed unloading next: %v",v)
        return
    }
    v = tr.Next()
    ts,_ = v.Timestamp()
    if v==nil || v.(*TextDatapoint).D!="test1" || ts!=2 {
        t.Errorf("Failed unloading next: %v",v)
    }
    v = tr.Next()
    if v!=nil {
        t.Errorf("Bad return: %v",v)
        return
    }

    tr.Close()

    tr = TypedRange{timebatchdb.EmptyRange{},NilType{}}

    if tr.Next()!=nil {
        t.Errorf("Empty return nonempty")
        return
    }
}

func TestTypedDatabase(t *testing.T) {
    m,err := timebatchdb.OpenMongoStore("localhost","testdb")
    if (err!=nil) {
       t.Errorf("Couldn't open MongoStore")
       return
    }
    defer m.Close()

    //First drop the collection - so that tests are fresh
    m.DropCollection("0")

    //Turn on the DataStore writer
    go timebatchdb.DatabaseWriter("localhost:4222","localhost","testdb", "testing/>")

    db,err := Open("localhost:4222","localhost","testdb")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer db.Close()

    //Wait for the DataStoreWriter to initialize
    time.Sleep(500 * time.Millisecond)

    //Now we test the database
    dt,ok := GetType("text")
    if !ok {
        t.Errorf("Bad type")
        return
    }
    dpoint := dt.New().(*TextDatapoint)

    dpoint.K = "testing/key1"
    dpoint.T = int64(1)
    dpoint.D = "Hello World!"
    err = db.Insert(dpoint,"")
    if err!=nil {
        t.Errorf("Insert failed: %s",err)
        return
    }
    dpoint.T=int64(2)
    dpoint.K="key2"
    dpoint.D = "hi"
    err = db.InsertKey("testing/key1",dpoint,"")
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

    v := r.Next()
    if v==nil || v.(*TextDatapoint).D !="Hello World!" {
        t.Errorf("Get incorrect datapoint %v",v)
        return
    }
    v = r.Next()
    if v==nil || v.(*TextDatapoint).D !="hi" {
        t.Errorf("Get incorrect datapoint %v",v)
        return
    }
    if r.Next()!=nil {
        t.Errorf("Got more than I bargained for")
        return
    }

    //Now test getting unknown datatypes
    r = db.GetTimeRange("testing/key1","badtype",0,505785867)
    if r.Next()!=nil {
        t.Errorf("Get nonexisting type failed")
        return
    }
    //Now test getting unknown datatypes
    r = db.GetIndexRange("testing/key1","badtype",0,505785867)
    if r.Next()!=nil {
        t.Errorf("Get nonexisting type failed")
        return
    }
}
