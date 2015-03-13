package timebatchdb

import (
    "testing"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    _ "github.com/lib/pq"
    "os"
    "errors"
    )

func TestDatabasError(t *testing.T) {
    err := errors.New("OSHIT")
    _,err2 := Open(nil,"","",100,err)
    if err!=err2 {
        t.Errorf("Error chain fail")
        return
    }
}

func TestDatabaseInsert(t *testing.T) {
    os.Remove("TESTING_timebatch.db")
    sdb,err := sql.Open("sqlite3","TESTING_timebatch.db")
    if err!=nil {
        t.Errorf("Couldn't open database: %v",err)
        return
    }
    defer sdb.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        t.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"sqlite3","localhost:6379",3,nil)
    if err!=nil {
        t.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    //First, test unordered input
    timestamps := []int64{1000,1500,2000,2100,2200,2500,3000,3100,3100}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8")}

    if db.Insert("hello",CreateDatapointArray(timestamps,data))!=ERROR_UNORDERED {
        t.Errorf("Wrong error on insert: %v",err)
        return
    }
    err = db.Insert("hello",CreateDatapointArray(timestamps[:1],data[:1]))
    if err!=nil {
        t.Errorf("error on insert: %v",err)
        return
    }
    l,err := db.Len("hello")
    if 1!=l || err!=nil {
        t.Errorf("Data length not correct %v %v",l,err)
    }
    err = db.Insert("hello",CreateDatapointArray(timestamps[:1],data[:1]))
    if err!=ERROR_TIMESTAMP {
        t.Errorf("wrong error on insert: %v",err)
        return
    }

    err = db.Insert("hello",CreateDatapointArray(timestamps[1:8],data[1:8]))
    if err!=nil {
        t.Errorf("error on insert: %v",err)
        return
    }

    //Now make sure that the key was pushed twice to the batch queue
    k,err := rc.BatchWait()
    if err!=nil || k!= "hello" {
        t.Errorf("Error in batch queue: %v, %s",err, k)
        return
    }
    k,err = rc.BatchWait()
    if err!=nil || k!= "hello" {
        t.Errorf("Error in batch queue: %v, %s",err, k)
        return
    }
    rc.BatchPush("END")
    k,err = rc.BatchWait()
    if err!=nil || k!= "END" {
        t.Errorf("Error in batch queue: %v, %s",err, k)
        return
    }
}


func TestDatabaseWrite(t *testing.T) {
    os.Remove("TESTING_timebatch.db")
    sdb,err := sql.Open("sqlite3","TESTING_timebatch.db")
    if err!=nil {
        t.Errorf("Couldn't open database: %v",err)
        return
    }
    defer sdb.Close()
    s,err := OpenSqlStore(sdb,"sqlite3",nil)
    if err!=nil {
        t.Errorf("Couldn't create SQLiteStore: %v",err)
        return
    }
    defer s.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        t.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"sqlite3","localhost:6379",3,nil)
    if err!=nil {
        t.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    timestamps := []int64{1000,1500,2000,2100,2200,2500,3000,3100}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7")}
    err = db.Insert("hello",CreateDatapointArray(timestamps,data))
    if err!=nil {
        t.Errorf("error on insert: %v",err)
        return
    }

    err = db.WriteDatabaseIteration()
    if err!=nil {
        t.Errorf("error on write: %v",err)
        return
    }
    err = db.WriteDatabaseIteration()
    if err!=nil {
        t.Errorf("error on write: %v",err)
        return
    }
    if i,_ := rc.CacheLength("hello"); i!=2 {
        t.Errorf("cache length wrong: %v",i)
        return
    }

    rc.BatchPush("NOTAKEY")
    err = db.WriteDatabaseIteration()   //Should repush the key and give an error
    if err!=nil {
        t.Errorf("error on write: %v",err)
        return
    }
    dr,ei,err := s.GetByIndex("hello",0)
    if err!=nil || ei!=0 {
        t.Errorf("Error getting from sql database: %v %v",err,ei)
        return
    }
    //Now make sure that the data actually exists in the sql database
    if !ensureValidityTest(t,timestamps[:6],data[:6],dr) {
        return
    }
}


func TestDatabaseRead(t *testing.T) {
    os.Remove("TESTING_timebatch.db")
    sdb,err := sql.Open("sqlite3","TESTING_timebatch.db")
    if err!=nil {
        t.Errorf("Couldn't open database: %v",err)
        return
    }
    defer sdb.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        t.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"sqlite3","localhost:6379",4,nil)
    if err!=nil {
        t.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    timestamps := []int64{1000,1500,2000,2100,2200,2500,3000}
    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6")}
    err = db.Insert("hello",CreateDatapointArray(timestamps,data))
    if err!=nil {
        t.Errorf("error on insert: %v",err)
        return
    }
    //Write to the sql database
    err = db.WriteDatabaseIteration()
    if err!=nil {
        t.Errorf("error on write: %v",err)
        return
    }
    _,err = db.GetIndexRange("hello",0,0)
    if err!=ERROR_USERFAIL {
        t.Errorf("Get by index range failure: %v",err)
        return
    }
    _,err = db.GetTimeRange("hello",0,0)
    if err!=ERROR_USERFAIL {
        t.Errorf("Get by index range failure: %v",err)
        return
    }
    dr,err := db.GetIndexRange("hello",0,6)
    dr.Init()
    if err!=nil {
        t.Errorf("Get by index range failure: %v",err)
        return
    }
    //Now make sure that the data actually exists in the sql database
    if !ensureValidityTest(t,timestamps[:6],data[:6],dr) {
        return
    }
    dr.Close()

    dr,err = db.GetIndexRange("hello",4,10)
    dr.Init()
    if err!=nil {
        t.Errorf("Get by index range failure: %v",err)
        return
    }
    //Now make sure that the data actually exists in the sql database
    if !ensureValidityTest(t,timestamps[4:],data[4:],dr) {
        return
    }
    dr.Close()

    dr,err = db.GetTimeRange("hello",100,2500)
    dr.Init()
    if err!=nil {
        t.Errorf("Get by time range failure: %v",err)
        return
    }
    //Now make sure that the data actually exists in the sql database
    if !ensureValidityTest(t,timestamps[:6],data[:6],dr) {
        return
    }
    dr.Close()
    dr,err = db.GetTimeRange("hello",2200,3900)
    dr.Init()
    if err!=nil {
        t.Errorf("Get by time range failure: %v",err)
        return
    }
    //Now make sure that the data actually exists in the sql database
    if !ensureValidityTest(t,timestamps[5:],data[5:],dr) {
        return
    }
    dr.Close()

    dr,err = db.GetTimeRange("hello",3800,3900)
    dr.Init()
    if v,_:= dr.Next(); err!=nil || v!=nil {
        t.Errorf("Get by time range failure: %v %v",err,v)
        return
    }
}

//This is a benchmark of how fast we can read out a thousand-datapoint range from sqlite in chunks of 100
func BenchmarkThousandS_S(b *testing.B) {
    os.Remove("TESTING_timebatch.db")
    sdb,err := sql.Open("sqlite3","TESTING_timebatch.db")
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer sdb.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        b.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"sqlite3","localhost:6379",100,nil)
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8"),[]byte("test9")}
    for n := int64(0);n<100;n++ {
        timestamps := []int64{0+n*10,1+n*10,2+n*10,3+n*10,4+n*10,5+n*10,6+n*10,7+n*10,8+n*10,9+n*10}
        db.Insert("testkey",CreateDatapointArray(timestamps,data))
        //db.WriteDatabaseIteration()
    }
    for i:=0;i<10;i++ {
        db.WriteDatabaseIteration()
    }
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        r,_ := db.GetIndexRange("testkey",0,1000)

        r.Init()
        for dp,_ := r.Next(); dp!= nil ;dp,_ = r.Next() {
            dp.Timestamp()
            dp.Data()
        }
        r.Close()
    }
}

//This is a benchmark of how fast we can read out a thousand-datapoint range from postgres in chunks of 100
func BenchmarkThousandS_P(b *testing.B) {
    sdb,err := sql.Open("postgres",TEST_postgresString)
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    sdb.Exec("DROP TABLE IF EXISTS timebatchtable")
    defer sdb.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        b.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"postgres","localhost:6379",100,nil)
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8"),[]byte("test9")}
    for n := int64(0);n<100;n++ {
        timestamps := []int64{0+n*10,1+n*10,2+n*10,3+n*10,4+n*10,5+n*10,6+n*10,7+n*10,8+n*10,9+n*10}
        db.Insert("testkey",CreateDatapointArray(timestamps,data))
        //db.WriteDatabaseIteration()
    }
    for i:=0;i<10;i++ {
        db.WriteDatabaseIteration()
    }
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        r,_ := db.GetIndexRange("testkey",0,1000)

        r.Init()
        for dp,_ := r.Next(); dp!= nil ;dp,_ = r.Next() {
            dp.Timestamp()
            dp.Data()
        }
        r.Close()
    }
}

//This is a benchmark of how fast we can read out a thousand-datapoint range from redis
func BenchmarkThousandR(b *testing.B) {
    os.Remove("TESTING_timebatch.db")
    sdb,err := sql.Open("sqlite3","TESTING_timebatch.db")
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer sdb.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        b.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"sqlite3","localhost:6379",100,nil)
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8"),[]byte("test9")}
    for n := int64(0);n<100;n++ {
        timestamps := []int64{0+n*10,1+n*10,2+n*10,3+n*10,4+n*10,5+n*10,6+n*10,7+n*10,8+n*10,9+n*10}
        db.Insert("testkey",CreateDatapointArray(timestamps,data))
        //db.WriteDatabaseIteration()
    }
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        r,_ := db.GetIndexRange("testkey",0,1000)

        r.Init()
        for dp,_ := r.Next(); dp!= nil ;dp,_ = r.Next() {
            dp.Timestamp()
            dp.Data()
        }
        r.Close()
    }
}

func BenchmarkInsert(b *testing.B) {
    os.Remove("TESTING_timebatch.db")
    sdb,err := sql.Open("sqlite3","TESTING_timebatch.db")
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer sdb.Close()

    rc,err := OpenRedisCache("localhost:6379",nil)
    if err!= nil {
        b.Errorf("Open Redis error %v",err)
        return
    }
    defer rc.Close()

    //Cleans the cache
    rc.Clear()

    db,err := Open(sdb,"sqlite3","localhost:6379",4,nil)
    if err!=nil {
        b.Errorf("Couldn't open database: %v",err)
        return
    }
    defer db.Close()

    data := [][]byte{[]byte("test0"),[]byte("test1"),[]byte("test2"),[]byte("test3"),
        []byte("test4"),[]byte("test5"),[]byte("test6"),[]byte("test7"),[]byte("test8"),[]byte("test9")}

    b.ResetTimer()
    for n := int64(0); n < int64(b.N); n++ {
        timestamps := []int64{0+n*10,1+n*10,2+n*10,3+n*10,4+n*10,5+n*10,6+n*10,7+n*10,8+n*10,9+n*10}
        db.Insert("testkey",CreateDatapointArray(timestamps,data))
    }


}
