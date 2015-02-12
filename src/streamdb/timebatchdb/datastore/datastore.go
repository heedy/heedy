//TimeBatchDB's Datastore is a time series DataStore built to handle extremely fast messaging as well as
//enormous quantities of data.
package datastore

//This is the object which handles all querying/inserting of data into the DataStore
type DataStore struct {
    //hc HotCache     //The cache of the most recent datapoints
    ws WarmStore    //The intermediate storage of the DataStore
    ms *Messenger    //The messaging system
}

func (d *DataStore) Close() {
    //d.hc.Close()
    d.ws.Close()
    d.ms.Close()
}

//Returns the DataRange associated with the given time range
func (d *DataStore) GetTimeRange(key string, starttime int64, endtime int64) DataRange {
    drl := NewRangeList()
    if (endtime <=starttime) {
        return drl  //The RangeList is empty on invalid params
    }
    drl.Append(d.ws.GetTime(key,starttime))
    //drl.Append(hc.Get(key))
    return NewTimeRange(drl,starttime,endtime)
}

//Returns the DataRange associated with the given index range
func (d *DataStore) GetIndexRange(key string, startindex uint64, endindex uint64) DataRange {
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

//Inserts the given data into the DataStore, and uses the given routing address for data
func (d *DataStore) Insert(key string, timestamp int64, data []byte,routing string) error {
    return d.ms.Publish(NewKeyedDatapoint(key,timestamp,data),routing)
}

//Opens the DataStore.
func Open(msgurl string, mongourl string, mongoname string) (*DataStore,error) {
    ms,err := ConnectMessenger(msgurl)
    if err!=nil {
        return nil,err
    }
    ws,err := OpenMongoStore(mongourl,mongoname)
    if err!=nil {
        ms.Close()
        return nil,err
    }

    return &DataStore{ws,ms},nil
}
