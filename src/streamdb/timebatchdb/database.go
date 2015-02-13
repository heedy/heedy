//TimeBatchDB is a time series Database built to handle extremely fast messaging as well as
//enormous quantities of data.
package timebatchdb

import (
    "streamdb/timebatchdb/datastore"
    "log"
    "errors"
    )

//This is the object which handles all querying/inserting of data into the DataStore
type Database struct {
    ds *datastore.DataStore
}

func (d *Database) Close() {
    d.ds.Close()
}

//Returns the DataRange associated with the given time range
func (d *Database) GetTimeRange(key string, dtype string, starttime int64, endtime int64) TypedRange {
    t,ok := GetType(dtype)
    if (!ok) {
        log.Printf("TimeBatchDB.Get: Unrecognized type '%s'\n",dtype)
        return TypedRange{datastore.EmptyRange{},NilType{}}
    }
    return TypedRange{d.ds.GetTimeRange(key,starttime,endtime),t}
}

//Returns the DataRange associated with the given index range
func (d *Database) GetIndexRange(key string, dtype string, startindex uint64, endindex uint64) TypedRange {
    t,ok := GetType(dtype)
    if (!ok) {
        log.Printf("TimeBatchDB.Get: Unrecognized type '%s'\n",dtype)
        return TypedRange{datastore.EmptyRange{},NilType{}}
    }
    return TypedRange{d.ds.GetIndexRange(key,startindex,endindex),t}
}

//Inserts the given data into the DataStore, and uses the given routing address for data
func (d *Database) Insert(datapoint interface{}, dtype string,routing string) error {
    s := ExtractKey(datapoint)
    if (s=="") {
        return errors.New("Key not found in datapoint")
    }
    return d.InsertKey(s,datapoint,dtype,routing)
}
func (d *Database) InsertKey(key string, datapoint interface{}, dtype string,routing string) error {
    t,ok := GetType(dtype)
    if (!ok) {
        log.Printf("TimeBatchDB.Insert: Unrecognized type '%s'\n",dtype)
        return errors.New("Unrecognized data type")
    }
    timestamp,data,err := t.Unload(datapoint)
    if err!=nil {
        return err
    }
    return d.ds.Insert(key,timestamp,data,routing)
}

//Opens the DataStore.
func Open(msgurl string, mongourl string, mongoname string) (*Database,error) {
    ds,err := datastore.Open(msgurl,mongourl,mongoname)
    if err!=nil {
        return nil,err
    }
    return &Database{ds},nil
}
