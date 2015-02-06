package timebatchdb

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "log"
    )

//The WarmStore interface - it allows to insert a DataArray to the storage, and to query a DataRange
//either by time or by index
type WarmStore interface {
    Append(key string,da *DatapointArray) error
    GetTime(key string, starttime uint64) DataRange
    GetIndex(key string, startindex uint64) DataRange
    Close()
}

//This is the struct which defines how stuff is stored in MongoDB
type mongostruct struct {
    Key string              //The key of the data array
    EndTime uint64        //The final timestamp of the array
    EndIndex uint64       //The index of the last element in the array
    Data []byte             //The byte representation of a datapoint
}




//This is a datarange built upon a mongoDB iterator.
type MongoRange struct {
    mdata *mgo.Iter     //The iterator for MongoDB
    da *DatapointArray  //The array of datapoints which is currently being processed
}

func (r *MongoRange) Close() {
    //Close might be called multiple times
    if (r.mdata!=nil) {
        r.mdata.Close()
        r.mdata = nil
    }
}

//Placeholder - does nothing
func (r *MongoRange) Init() {
}

func (r *MongoRange) Next() *Datapoint {
    d := r.da.Next()
    if d!=nil {
        return d
    }

    //The DatapointArray is now empty - check that the iterator is functional
    if r.mdata== nil {
        return nil
    }
    dp := mongostruct{}
    if (!r.mdata.Next(&dp)) {
        //Crap, there was an error. We therefore return nil - we can't do anything.
        //it is probably because the iterator is finished. Close the database connection.
        r.Close()
        return nil
    }

    //Okay, the next thing was loaded. We convert the data to a DatapointArray
    r.da = DatapointArrayFromBytes(dp.Data)

    //Now, we repeat the procedure
    return r.Next()

}


type MongoStore struct {
    session *mgo.Session    //The mongoDB session
    db *mgo.Database        //The database
    c  *mgo.Collection     //Temporary: The collection used for all data
}



func (s *MongoStore) Close() {
    s.session.Close()
}

//Inserts the datapoint array into the warmstore - only needing to know its startindex
func (s *MongoStore) Insert(key string,startindex uint64, da *DatapointArray) error {
    if (da.Len()!=0) {
        return s.c.Insert(&mongostruct{key,da.Datapoints[da.Len()-1].Timestamp(),
            startindex+uint64(da.Len()),da.Bytes()})
    }
    return nil
}


//The microstruct allows to fetch the endIndex quickly
type micromongostruct struct {
    EndIndex uint64
}
//Returns the first index point outside of the most recent datapointarray stored within the database.
//In effect, if the datapontis in a key were all in one huge array, returns array.length
//(not including the datapoints which are not yet committed to warmstore)
func (s *MongoStore) GetEndIndex(key string) (uint64,error) {
    result := micromongostruct{}
    err := s.c.Find(bson.M{"key": key}).Sort("-endindex").Select(bson.M{"endindex":1}).One(&result)
    if err != nil {
        if (err==mgo.ErrNotFound) {
            return 0,nil
        }
        return 0,err
    }
    return result.EndIndex,nil
}

//Appends the given DatapointArray to the data stream for key
func (s *MongoStore) Append(key string, dp *DatapointArray) error {
    i,err := s.GetEndIndex(key)
    if (err!=nil) {
        return err
    }
    return s.Insert(key,i,dp)
}

func (s *MongoStore) GetTime(key string, starttime uint64) DataRange {
    i := s.c.Find(bson.M{"key": key,"endtime": bson.M{"$gt": starttime}}).Sort("endtime").Iter()
    dp := mongostruct{}
    if (!i.Next(&dp)) {
        //Looks like there are no documents that fit the criteria.
        //We therefore close the Iterator, and return an empty RangeList (which has correct behavior
        //on empty)
        i.Close()
        return EmptyRange{}
    }

    da := DatapointArrayFromBytes(dp.Data)
    da = da.TStart(starttime)
    if (da==nil) {
        //This is a legit error. This means that the database is corrupted!
        log.Println("MongoStore: Key=%s T=%d Corrupted!",key,starttime)
        i.Close()
        return EmptyRange{}
    }

    //And return the result
    return &MongoRange{i,da}
}

func (s *MongoStore) GetIndex(key string, startindex uint64) DataRange {
    i := s.c.Find(bson.M{"key": key,"endindex": bson.M{"$gt": startindex}}).Sort("endindex").Iter()
    dp := mongostruct{}
    if (!i.Next(&dp)) {
        //Looks like there are no documents that fit the criteria.
        i.Close()
        return EmptyRange{}
    }

    //Okay, now we convert the datapoints
    da := DatapointArrayFromBytes(dp.Data)

    //Lastly, we start the DatapointArray from the correct index
    //This is guaranteed to work on uint, since query requires $gt
    fromend := dp.EndIndex-startindex

    //Lastly, make sure that the given index is within range
    if (fromend > uint64(da.Len())) {
        log.Println("MongoStore: Key=%s I=%d Corrupted!",key,startindex)
        i.Close()
        return EmptyRange{}
    }

    //And return the result
    return &MongoRange{i,NewDatapointArray(da.Datapoints[da.Len()-int(fromend):])}
}

func (s *MongoStore) DropCollection(name string) {
    s.db.C(name).DropCollection()
}

//Opens the database at the given URL
func OpenMongoStore(dburl string,dbname string) (*MongoStore,error) {
    session, err := mgo.Dial(dburl)
    if (err!=nil) {
        return nil,err
    }
    db := session.DB(dbname)
    c := db.C("0")

    //Make sure that the index exists
    index := mgo.Index{
        Key: []string{"key", "endtime"},
        Unique: false,
        DropDups: false,
        Background: true,
        Sparse: true,
    }
    err = c.EnsureIndex(index)

    index = mgo.Index{
        Key: []string{"-key", "endindex"},
        Unique: false,
        DropDups: false,
        Background: true,
        Sparse: true,
    }
    err = c.EnsureIndex(index)

    return &MongoStore{session,db,c},nil
}
