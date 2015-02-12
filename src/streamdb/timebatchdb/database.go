//TimeBatchDB is a time series Database built to handle extremely fast messaging as well as
//enormous quantities of data.
package timebatchdb

import (
    "streamdb/timebatchdb/datastore"
    )

//This is the object which handles all querying/inserting of data into the DataStore
type Database struct {
    ds *datastore.DataStore
}

func (d *Database) Close() {
    d.ds.Close()
}

//Returns the DataRange associated with the given time range
func (d *Database) GetTimeRange(key string, starttime int64, endtime int64) datastore.DataRange {
    return d.ds.GetTimeRange(key,starttime,endtime)
}

//Returns the DataRange associated with the given index range
func (d *Database) GetIndexRange(key string, startindex uint64, endindex uint64) datastore.DataRange {
    return d.ds.GetIndexRange(key,startindex,endindex)
}

//Inserts the given data into the DataStore, and uses the given routing address for data
func (d *Database) Insert(key string, timestamp int64, data []byte,routing string) error {
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
