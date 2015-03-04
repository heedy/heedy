//Dtypes is a package that handles data types for timebatchdb. It allows saving and converting between different data
//types transparently to timebatchdb's byte arrays
package dtypes

import (
    "streamdb/timebatchdb"
    "log"
    "errors"
    "database/sql"
    )

var (
    ERROR_KEYNOTFOUND = errors.New("Key not found in datapoint")
    //ERROR_UNKNOWNDTYPE = errors.New("Unrecognized data type")
    )

//A simple wrapper for DataRange which returns marshalled data
type TypedRange struct {
    dr timebatchdb.DataRange
    dtype DataType
}

func (tr TypedRange) Close() {
    tr.dr.Close()
}
func (tr TypedRange) Next() TypedDatapoint {
    d,_ := tr.dr.Next()
    if d==nil {
        return nil
    }
    dp := tr.dtype.New()
    err := dp.Load(*d)
    if err!= nil {
        return nil
    }
    return dp
}



//Simple type wrapper for timebatchdb's Database
type TypedDatabase struct {
    db *timebatchdb.Database
}

func (d *TypedDatabase) Close() {
    d.db.Close()
}

//Returns the DataRange associated with the given time range
func (d *TypedDatabase) GetTimeRange(key string, dtype string, starttime int64, endtime int64) TypedRange {
    t,ok := GetType(dtype)
    if (!ok) {
        log.Printf("TypedDatabase.Get: Unrecognized type '%s'\n",dtype)
        return TypedRange{timebatchdb.EmptyRange{},NilType{}}
    }
    val,_:=d.db.GetTimeRange(key,starttime,endtime)
    return TypedRange{val,t}
}

//Returns the DataRange associated with the given index range
func (d *TypedDatabase) GetIndexRange(key string, dtype string, startindex uint64, endindex uint64) TypedRange {
    t,ok := GetType(dtype)
    if (!ok) {
        log.Printf("TypedDatabase.Get: Unrecognized type '%s'\n",dtype)
        return TypedRange{timebatchdb.EmptyRange{},NilType{}}
    }
    val,_:=d.db.GetIndexRange(key,startindex,endindex)
    return TypedRange{val,t}
}

//Inserts the given data into the DataStore, and uses the given routing address for data
func (d *TypedDatabase) Insert(datapoint TypedDatapoint) error {
    s := datapoint.Key()
    if (s=="") {
        return ERROR_KEYNOTFOUND
    }
    return d.InsertKey(s,datapoint)
}
func (d *TypedDatabase) InsertKey(key string, datapoint TypedDatapoint) error {
    timestamp,err := datapoint.Timestamp()
    data := datapoint.Data()
    if err!=nil {
        return err
    }
    return d.db.Insert(key,timebatchdb.CreateDatapointArray([]int64{timestamp},[][]byte{data}))
}

//Opens the DataStore.
func Open(sdb *sql.DB, sqlstring string, redisurl string) (*TypedDatabase,error) {

    var td  TypedDatabase
    err := td.InitTypedDB(sdb,sqlstring,redisurl)

    if err != nil {
        return nil, err
    }

    return &td, nil
    /**
    TODO removeme when all tests check out.

    ds,err := timebatchdb.Open(msgurl,mongourl,mongoname)
    if err!=nil {
        return nil,err
    }
    return &TypedDatabase{ds},nil
    **/
}

// Initializes a Typed Database that already exists.
func (d *TypedDatabase) InitTypedDB(sdb *sql.DB, sqlstring string, redisurl string) (error) {
    ds, err := timebatchdb.Open(sdb,sqlstring,redisurl,10,nil)
    if err != nil {
        return err
    }
    d.db = ds
    return nil
}
