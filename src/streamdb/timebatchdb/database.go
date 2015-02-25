//TimeBatchDB is a time series Database built to handle extremely fast messaging as well as
//enormous quantities of data.
package timebatchdb

//This is the object which handles all querying/inserting of data into the Database
type Database struct {
    //hc HotCache     //The cache of the most recent datapoints
    ws WarmStore    //The intermediate storage of the Database
}

func (d *Database) Close() {
    //d.hc.Close()
    d.ws.Close()
}

//Returns the DataRange associated with the given time range
func (d *Database) GetTimeRange(key string, starttime int64, endtime int64) DataRange {
    drl := NewRangeList()
    if (endtime <=starttime) {
        return drl  //The RangeList is empty on invalid params
    }
    drl.Append(d.ws.GetTime(key,starttime))
    //drl.Append(hc.Get(key))
    return NewTimeRange(drl,starttime,endtime)
}

//Returns the DataRange associated with the given index range
func (d *Database) GetIndexRange(key string, startindex uint64, endindex uint64) DataRange {
    //BUG(daniel): Getting ranges makes the critical assumption that each element in the stream
    //has a STRICTLY increasing timestamp. That means that no two elements share the same time stamp.
    //This allows us to make the range-getting code incredibly simple

    drl := NewRangeList()
    if (endindex <=startindex) {
        return drl  //The RangeList is empty on invalid params
    }
    drl.Append(d.ws.GetIndex(key,startindex))
    //drl.Append(hc.Get(key))
    return NewNumRange(drl,endindex-startindex)
}

//Inserts the given data into the Database
func (d *Database) Insert(key string, timestamp int64, data []byte) error {
    return d.ws.Append(key,NewDatapointArray([]Datapoint{NewDatapoint(timestamp,data)}))
}

//Opens the Database.
func Open(mongourl string, mongoname string) (*Database,error) {
    ws,err := OpenMongoStore(mongourl,mongoname)
    if err!=nil {
        return nil,err
    }

    return &Database{ws},nil
}
