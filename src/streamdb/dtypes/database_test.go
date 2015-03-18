package dtypes

import (
	"database/sql"
	"os"
	"streamdb/timebatchdb"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestTypedRange(t *testing.T) {
	timestamps := []int64{1, 2}
	datab := [][]byte{[]byte("test0"), []byte("test1")}

	da := timebatchdb.CreateDatapointArray(timestamps, datab, "")

	dtype, ok := GetType("s")
	if !ok {
		t.Errorf("Type s not found")
		return
	}
	da.Init()
	tr := TypedRange{da, dtype}

	v := tr.Next()
	ts, _ := v.Timestamp()
	if v == nil || v.(*TextDatapoint).D != "test0" || ts != 1 {
		t.Errorf("Failed unloading next: %v", v)
		return
	}
	v = tr.Next()
	ts, _ = v.Timestamp()
	if v == nil || v.(*TextDatapoint).D != "test1" || ts != 2 {
		t.Errorf("Failed unloading next: %v", v)
	}
	v = tr.Next()
	if v != nil {
		t.Errorf("Bad return: %v", v)
		return
	}

	tr.Close()

	tr = TypedRange{timebatchdb.EmptyRange{}, NilType{}}

	if tr.Next() != nil {
		t.Errorf("Empty return nonempty")
		return
	}
}

func TestTypedDatabase(t *testing.T) {
	os.Remove("TESTING_timebatch.db")
	sdb, err := sql.Open("sqlite3", "TESTING_timebatch.db")
	if err != nil {
		t.Errorf("Couldn't open database: %v", err)
		return
	}
	defer sdb.Close()

	rc, err := timebatchdb.OpenRedisCache("localhost:6379", nil)
	if err != nil {
		t.Errorf("Open Redis error %v", err)
		return
	}
	defer rc.Close()

	//Cleans the cache
	rc.Clear()

	db, err := Open(sdb, "sqlite3", "localhost:6379", 10, nil)
	if err != nil {
		t.Errorf("Couldn't connect: %s", err)
		return
	}
	defer db.Close()

	//Now we test the database
	dt, ok := GetType("s")
	if !ok {
		t.Errorf("Bad type")
		return
	}
	dpoint := dt.New().(*TextDatapoint)

	dpoint.K = "testing/key1"
	dpoint.T = int64(1)
	dpoint.D = "Hello World!"
	err = db.Insert(dpoint)
	if err != nil {
		t.Errorf("Insert failed: %s", err)
		return
	}
	dpoint.T = int64(2)
	dpoint.K = "key2"
	dpoint.D = "hi"
	err = db.InsertKey("testing/key1", dpoint)
	if err != nil {
		t.Errorf("Insert failed: %s", err)
		return
	}

	r := db.GetTimeRange("testing/randomkey", "s", 0, 505785867)

	if r.Next() != nil {
		t.Errorf("Get nonexisting failed")
		return
	}

	r = db.GetIndexRange("testing/key1", "s", 0, 50)
	defer r.Close()

	v := r.Next()
	if v == nil || v.(*TextDatapoint).D != "Hello World!" {
		t.Errorf("Get incorrect datapoint %v", v)
		return
	}
	v = r.Next()
	if v == nil || v.(*TextDatapoint).D != "hi" {
		t.Errorf("Get incorrect datapoint %v", v)
		return
	}
	if r.Next() != nil {
		t.Errorf("Got more than I bargained for")
		return
	}

	//Now test getting unknown datatypes
	r = db.GetTimeRange("testing/key1", "badtype", 0, 505785867)
	if r.Next() != nil {
		t.Errorf("Get nonexisting type failed")
		return
	}
	//Now test getting unknown datatypes
	r = db.GetIndexRange("testing/key1", "badtype", 0, 505785867)
	if r.Next() != nil {
		t.Errorf("Get nonexisting type failed")
		return
	}

	rc.Clear()
}
