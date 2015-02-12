package timebatchdb

import (
    "testing"
    "streamdb/timebatchdb/datastore"
    )

func TestDataStore(t *testing.T) {
    //Turn on the DataStore writer
    go datastore.DataStoreWriter("localhost:4222","localhost","testdb", "testing/>")

    m,err := datastore.OpenMongoStore("localhost","testdb")
    if (err!=nil) {
       t.Errorf("Couldn't open MongoStore")
       return
    }
    defer m.Close()

    //First drop the collection - so that tests are fresh
    m.DropCollection("0")

    db,err := Open("localhost:4222","localhost","testdb")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer db.Close()



}
