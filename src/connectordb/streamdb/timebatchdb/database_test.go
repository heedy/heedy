package timebatchdb

import (
	"database/sql"
	"errors"

	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestDatabasError(t *testing.T) {
	err := errors.New("OSHIT")
	_, err2 := Open(nil, "", "", 100, err)
	if err != err2 {
		t.Errorf("Error chain fail")
		return
	}
}

func openTestingEnvironment(t testing.TB) (*Database, *RedisCache, *SqlStore, error) {
	rc, err := OpenRedisCache("localhost:6379", nil)
	require.Nil(t, err, "Couldn't open redis %v", err)

	// Clear the cache before testing
	rc.Clear()

	sdb, err := sql.Open("postgres", "postgres://127.0.0.1:52592/connectordb?sslmode=disable")
	require.Nil(t, err, "Couldn't open database %v", err)
	TableMakerTestCreate(sdb)

	db, err := Open(sdb, "postgres", "localhost:6379", 3, nil)
	require.Nil(t, err, "Couldn't call open: %v", err)

	s, err := OpenSqlStore(sdb, "postgres", nil)
	require.Nil(t, err, "Couldn't create sqlstore: %v", err)

	return db, rc, s, err
}

func TestDatabaseInsert(t *testing.T) {

	db, rc, s, err := openTestingEnvironment(t)
	require.Nil(t, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	//First, test unordered input
	timestamps := []int64{1000, 1500, 2000, 2100, 2200, 2500, 3000, 3100, 3100}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	if db.Insert("hello", CreateDatapointArray(timestamps, data, "")) != ErrorUnordered {
		t.Errorf("Wrong error on insert: %v", err)
		return
	}
	err = db.Insert("hello", CreateDatapointArray(timestamps[:1], data[:1], ""))
	if err != nil {
		t.Errorf("error on insert: %v", err)
		return
	}
	l, err := db.Len("hello")
	if 1 != l || err != nil {
		t.Errorf("Data length not correct %v %v", l, err)
	}
	err = db.Insert("hello", CreateDatapointArray(timestamps[:1], data[:1], ""))
	if err != ErrorTimestamp {
		t.Errorf("wrong error on insert: %v", err)
		return
	}

	err = db.Insert("hello", CreateDatapointArray(timestamps[1:8], data[1:8], ""))
	if err != nil {
		t.Errorf("error on insert: %v", err)
		return
	}

	//Now make sure that the key was pushed to the batch queue
	rc.BatchPush("END")

	for i := 0; i < 2; i++ {
		k, err := rc.BatchWait()
		if err != nil || k != "hello" {
			t.Errorf("Error in batch queue: %v, %s %v", err, k, i)
			return
		}
	}

	k, err := rc.BatchWait()
	if err != nil || k != "END" {
		t.Errorf("Error in batch queue: %v, %s", err, k)
		return
	}
}

func TestDatabaseWrite(t *testing.T) {
	db, rc, s, err := openTestingEnvironment(t)
	require.Nil(t, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	timestamps := []int64{1000, 1500, 2000, 2100, 2200, 2500, 3000, 3100}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7")}
	err = db.Insert("hello", CreateDatapointArray(timestamps, data, ""))
	if err != nil {
		t.Errorf("error on insert: %v", err)
		return
	}
	for i := 0; i < 2; i++ {
		err = db.WriteDatabaseIteration()
		if err != nil {
			t.Errorf("error on write: %v %v", i, err)
			return
		}
	}
	if i, _ := rc.CacheLength("hello"); i != 2 {
		t.Errorf("cache length wrong: %v", i)
		return
	}

	rc.BatchPush("NOTAKEY")
	err = db.WriteDatabaseIteration() //Should just ignore the bad key
	if err != nil {
		t.Errorf("error on write: %v", err)
		return
	}
	dr, ei, err := s.GetByIndex("hello", 0)
	if err != nil || ei != 0 {
		t.Errorf("Error getting from sql database: %v %v", err, ei)
		return
	}
	//Now make sure that the data actually exists in the sql database
	if !ensureValidityTest(t, timestamps[:6], data[:6], dr) {
		return
	}
}

func TestDatabaseRead(t *testing.T) {
	db, rc, s, err := openTestingEnvironment(t)
	require.Nil(t, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	timestamps := []int64{1000, 1500, 2000, 2100, 2200, 2500, 3000}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6")}
	err = db.Insert("hello", CreateDatapointArray(timestamps, data, ""))
	if err != nil {
		t.Errorf("error on insert: %v", err)
		return
	}
	//Write to the sql database
	err = db.WriteDatabaseIteration()
	if err != nil {
		t.Errorf("error on write: %v", err)
		return
	}
	_, err = db.GetIndexRange("hello", 2, 1)
	if err != ErrorUserFail {
		t.Errorf("Get by index range failure: %v", err)
		return
	}
	_, err = db.GetTimeRange("hello", 3, 2)
	if err != ErrorUserFail {
		t.Errorf("Get by index range failure: %v", err)
		return
	}
	dr, err := db.GetIndexRange("hello", 0, 6)
	dr.Init()
	if err != nil {
		t.Errorf("Get by index range failure: %v", err)
		return
	}
	//Now make sure that the data actually exists in the sql database
	if !ensureValidityTest(t, timestamps[:6], data[:6], dr) {
		return
	}
	dr.Close()

	dr, err = db.GetIndexRange("hello", 4, 10)
	dr.Init()
	if err != nil {
		t.Errorf("Get by index range failure: %v", err)
		return
	}
	//Now make sure that the data actually exists in the sql database
	if !ensureValidityTest(t, timestamps[4:], data[4:], dr) {
		return
	}
	dr.Close()

	dr, err = db.GetTimeRange("hello", 100, 2500)
	dr.Init()
	if err != nil {
		t.Errorf("Get by time range failure: %v", err)
		return
	}
	//Now make sure that the data actually exists in the sql database
	if !ensureValidityTest(t, timestamps[:6], data[:6], dr) {
		return
	}
	dr.Close()
	dr, err = db.GetTimeRange("hello", 2200, 3900)
	dr.Init()
	if err != nil {
		t.Errorf("Get by time range failure: %v", err)
		return
	}
	//Now make sure that the data actually exists in the sql database
	if !ensureValidityTest(t, timestamps[5:], data[5:], dr) {
		return
	}
	dr.Close()

	dr, err = db.GetTimeRange("hello", 3800, 3900)
	dr.Init()
	if v, _ := dr.Next(); err != nil || v != nil {
		t.Errorf("Get by time range failure: %v %v", err, v)
		return
	}
	i, err := db.GetTimeIndex("hello", 2530)
	require.NoError(t, err)
	require.Equal(t, 6, int(i))
	i, err = db.GetTimeIndex("hello", 10)
	require.NoError(t, err)
	require.Equal(t, 0, int(i))
}

func TestDatabaseDelete(t *testing.T) {
	db, rc, s, err := openTestingEnvironment(t)
	require.Nil(t, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	timestamps := []int64{1000, 1500, 2000, 2100, 2200, 2500, 3000}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6")}
	err = db.Insert("hello", CreateDatapointArray(timestamps, data, ""))
	if err != nil {
		t.Errorf("error on insert: %v", err)
		return
	}
	//Write to the sql database
	err = db.WriteDatabaseIteration()
	if err != nil {
		t.Errorf("error on write: %v", err)
		return
	}
	l, err := db.Len("hello")
	if l != 7 || err != nil {
		t.Errorf("wrong length: %v %v", l, err)
		return
	}
	err = db.Delete("hello")
	if err != nil {
		t.Errorf("Failed to delete %v", err)
	}
	l, err = db.Len("hello")
	if l != 0 || err != nil {
		t.Errorf("wrong length: %v %v", l, err)
		return
	}
	dr, err := db.GetIndexRange("hello", 0, 6)
	dr.Init()
	if err != nil {
		t.Errorf("Get deleted by index range failure: %v", err)
		return
	}
	dp, err := dr.Next()
	if dp != nil || err != nil {
		t.Errorf("Next on deleted: %v %v", err, dp)
		return
	}

	err = db.Insert("hello/world", CreateDatapointArray(timestamps, data, ""))
	if err != nil {
		t.Errorf("error on insert: %v", err)
		return
	}
	//Write to the sql database
	err = db.WriteDatabaseIteration()
	if err != nil {
		t.Errorf("error on write: %v", err)
		return
	}
	l, err = db.Len("hello/world")
	if l != 7 || err != nil {
		t.Errorf("wrong length: %v %v", l, err)
		return
	}

	err = db.DeletePrefix("hello/")
	if err != nil {
		t.Errorf("Failed to delete %v", err)
	}
	l, err = db.Len("hello/world")
	if l != 0 || err != nil {
		t.Errorf("wrong length: %v %v", l, err)
		return
	}
	dr, err = db.GetIndexRange("hello/world", 0, 6)
	dr.Init()
	if err != nil {
		t.Errorf("Get deleted by index range failure: %v", err)
		return
	}

	dp, err = dr.Next()
	if dp != nil || err != nil {
		t.Errorf("Next on deleted: %v %v", err, dp)
		return
	}
}

//This is a benchmark of how fast we can read out a thousand-datapoint range from postgres in chunks of 100
func BenchmarkThousandS_P(b *testing.B) {
	db, rc, s, err := openTestingEnvironment(b)
	require.Nil(b, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	for n := int64(0); n < 100; n++ {
		timestamps := []int64{0 + n*10, 1 + n*10, 2 + n*10, 3 + n*10, 4 + n*10, 5 + n*10, 6 + n*10, 7 + n*10, 8 + n*10, 9 + n*10}
		db.Insert("testkey", CreateDatapointArray(timestamps, data, ""))

	}
	for i := 0; i < 10; i++ {
		db.WriteDatabaseIteration()
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _ := db.GetIndexRange("testkey", 0, 1000)

		r.Init()
		for dp, _ := r.Next(); dp != nil; dp, _ = r.Next() {
			dp.Timestamp()
			dp.Data()
		}
		r.Close()
	}
}

//This is a benchmark of how fast we can read out a thousand-datapoint range from redis
func BenchmarkThousandR(b *testing.B) {
	db, rc, s, err := openTestingEnvironment(b)
	require.Nil(b, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	for n := int64(0); n < 100; n++ {
		timestamps := []int64{0 + n*10, 1 + n*10, 2 + n*10, 3 + n*10, 4 + n*10, 5 + n*10, 6 + n*10, 7 + n*10, 8 + n*10, 9 + n*10}
		db.Insert("testkey", CreateDatapointArray(timestamps, data, ""))
		//db.WriteDatabaseIteration()
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _ := db.GetIndexRange("testkey", 0, 1000)

		r.Init()
		for dp, _ := r.Next(); dp != nil; dp, _ = r.Next() {
			dp.Timestamp()
			dp.Data()
		}
		r.Close()
	}
}

func BenchmarkInsert(b *testing.B) {
	db, rc, s, err := openTestingEnvironment(b)
	require.Nil(b, err, "Could not setup testing.")
	defer rc.Close()
	defer db.Close()
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}

	b.ResetTimer()
	for n := int64(0); n < int64(b.N); n++ {
		timestamps := []int64{0 + n*10, 1 + n*10, 2 + n*10, 3 + n*10, 4 + n*10, 5 + n*10, 6 + n*10, 7 + n*10, 8 + n*10, 9 + n*10}
		db.Insert("testkey", CreateDatapointArray(timestamps, data, ""))
	}

}
