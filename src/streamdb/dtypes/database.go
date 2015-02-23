//Dtypes is a package that handles data types for timebatchdb. It allows saving and converting between different data
//types transparently to timebatchdb's byte arrays
package dtypes

import (
    "streamdb/timebatchdb"
    "log"
    "errors"
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
    d := tr.dr.Next()
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
    return TypedRange{d.db.GetTimeRange(key,starttime,endtime),t}
}

//Returns the DataRange associated with the given index range
func (d *TypedDatabase) GetIndexRange(key string, dtype string, startindex uint64, endindex uint64) TypedRange {
    t,ok := GetType(dtype)
    if (!ok) {
        log.Printf("TypedDatabase.Get: Unrecognized type '%s'\n",dtype)
        return TypedRange{timebatchdb.EmptyRange{},NilType{}}
    }
    return TypedRange{d.db.GetIndexRange(key,startindex,endindex),t}
}

//Inserts the given data into the DataStore, and uses the given routing address for data
func (d *TypedDatabase) Insert(datapoint TypedDatapoint,routing string) error {
    s := datapoint.Key()
    if (s=="") {
        return ERROR_KEYNOTFOUND
    }
    return d.InsertKey(s,datapoint,routing)
}
func (d *TypedDatabase) InsertKey(key string, datapoint TypedDatapoint,routing string) error {
    timestamp,err := datapoint.Timestamp()
    data := datapoint.Data()
    if err!=nil {
        return err
    }
    return d.db.Insert(key,timestamp,data,routing)
}

//Opens the DataStore.
func Open(msgurl string, mongourl string, mongoname string) (*TypedDatabase,error) {

    var td  TypedDatabase
    err := td.InitTypedDB(msgurl, mongourl, mongoname)

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
func (d *TypedDatabase) InitTypedDB(msgurl string, mongourl string, mongoname string) (error) {
    ds, err := timebatchdb.Open(msgurl, mongourl, mongoname)
    if err != nil {
        return err
    }
    d.db = ds
    return nil
}
