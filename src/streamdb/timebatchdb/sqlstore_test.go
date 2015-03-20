package timebatchdb

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var (
	TEST_postgresString = "sslmode=disable dbname=connectordb port=52592"
)

func SqlStoreTest(s *SqlStore, t *testing.T) {

	//First check returning empties
	i, err := s.GetEndIndex("hello/world")
	if err != nil || i != 0 {
		t.Errorf("EndIndex of empty failed: %d %s", i, err)
		return
	}

	r, si, err := s.GetByTime("hello/world", 0)
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	if r.Init() != nil {
		t.Errorf("Error in DataRange init")
		return
	}
	defer r.Close()
	dp, err := r.Next()
	if dp != nil || err != nil {
		t.Errorf("Failed to get by time: %v", err)
		return
	}

	r, si, err = s.GetByIndex("hello/world", 0)
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	defer r.Close()
	dp, err = r.Next()
	if dp != nil || err != nil {
		t.Errorf("Failed to get by index: %v", err)
		return
	}
	dp, err = r.Next()
	if dp != nil || err != nil {
		t.Errorf("Failed to get by index: %v", err)
		return
	}

	//Now insert some data
	timestamps := []int64{1, 2, 3, 4, 5, 6, 6, 7, 8}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	err = s.Append("hello/world", CreateDatapointArray(timestamps[0:1], data[0:1], ""))
	if err != nil {
		t.Errorf("Error in append: %s", err)
		return
	}

	i, err = s.GetEndIndex("hello/world")
	if err != nil || i != 1 {
		t.Errorf("EndIndex of nonempty failed: %d %s", i, err)
		return
	}
	i, err = s.GetEndIndex("hello/world2")
	if err != nil || i != 0 {
		t.Errorf("EndIndex of empty failed: %d %s", i, err)
		return
	}

	//TIME

	r, si, err = s.GetByTime("hello/world", 0)
	if r.Init() != nil {
		t.Errorf("Error in DataRange init")
		return
	}
	defer r.Close()
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if err != nil || dp == nil || dp.Timestamp() != 1 {
		t.Errorf("incorrect data for time: %v", err)
		return
	}
	dp, err = r.Next()
	if err != nil || dp != nil {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if err != nil || dp != nil {
		t.Errorf("incorrect data for time")
		return
	}

	r, si, err = s.GetByTime("hello/world", 1)
	defer r.Close()
	if si != 1 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if err != nil || dp != nil {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if err != nil || dp != nil {
		t.Errorf("incorrect data for time")
		return
	}

	//INDEX

	r, si, err = s.GetByIndex("hello/world", 0)
	defer r.Close()
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if err != nil || dp == nil || dp.Timestamp() != 1 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if err != nil || dp != nil {
		t.Errorf("incorrect data for index")
		return
	}

	r, si, err = s.GetByIndex("hello/world", 1)
	defer r.Close()
	if si != 1 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if err != nil || dp != nil || err != nil {
		t.Errorf("incorrect data for index")
		return
	}

	err = s.Append("hello/world", CreateDatapointArray(timestamps[1:3], data[1:3], ""))
	if err != nil {
		t.Errorf("Error in append: %s", err)
		return
	}
	i, err = s.GetEndIndex("hello/world")
	if err != nil || i != 3 {
		t.Errorf("EndIndex of nonempty failed: %d %s", i, err)
		return
	}

	r, si, err = s.GetByTime("hello/world", 0)
	defer r.Close()
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 1 {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 2 {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 3 {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if dp != nil || err != nil {
		t.Errorf("incorrect data for time")
		return
	}

	r, si, err = s.GetByTime("hello/world", 2)
	defer r.Close()
	if si != 2 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 3 {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if dp != nil || err != nil {
		t.Errorf("incorrect data for time")
		return
	}

	r, si, err = s.GetByIndex("hello/world", 0)
	defer r.Close()
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 1 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 2 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 3 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if dp != nil || err != nil {
		t.Errorf("incorrect data for index")
		return
	}

	r, si, err = s.GetByIndex("hello/world", 2)
	if si != 2 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	defer r.Close()
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 3 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if dp != nil || err != nil {
		t.Errorf("incorrect data for index")
		return
	}

	err = s.Append("hello/world", CreateDatapointArray(timestamps[3:], data[3:], ""))
	if err != nil {
		t.Errorf("Error in append: %s", err)
		return
	}
	i, err = s.GetEndIndex("hello/world")
	if err != nil || i != 9 {
		t.Errorf("EndIndex of nonempty failed: %d %s", i, err)
		return
	}

	r, si, err = s.GetByTime("hello/world", 4)
	if si != 4 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	defer r.Close()
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 5 {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 6 {
		t.Errorf("incorrect data for time")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 6 {
		t.Errorf("incorrect data for time")
		return
	}
	r.Close() //Test doulbe-closing

	r, si, err = s.GetByIndex("hello/world", 4)
	defer r.Close()
	if si != 4 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 5 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 6 {
		t.Errorf("incorrect data for index")
		return
	}
	dp, err = r.Next()
	if dp == nil || err != nil || dp.Timestamp() != 6 {
		t.Errorf("incorrect data for index")
		return
	}
	r.Close()

	err = s.Delete("hello/world")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	//Check returning empties
	i, err = s.GetEndIndex("hello/world")
	if err != nil || i != 0 {
		t.Errorf("EndIndex of deleted failed: %d %s", i, err)
		return
	}

	r, si, err = s.GetByTime("hello/world", 0)
	if si != 0 || err != nil {
		t.Errorf("get failed %v %v", si, err)
		return
	}
}

func TestNoDriver(t *testing.T) {
	err2 := errors.New("FAILTEST")
	_, err := OpenSqlStore(nil, "", err2)
	if err != err2 {
		t.Errorf("Fail error chain")
		return
	}
	_, err = OpenSqlStore(nil, "notavaliddriver", nil)
	if err != ErrorDatabaseDriver {
		t.Errorf("Bad database driver reaction")
		return
	}
}

func TestSQLiteStore(t *testing.T) {
	os.Remove("TESTING_timebatch.db")
	db, err := sql.Open("sqlite3", "TESTING_timebatch.db")
	if err != nil {
		t.Errorf("Couldn't open database: %v", err)
		return
	}
	defer db.Close()
	s, err := OpenSqlStore(db, "sqlite3", nil)
	if err != nil {
		t.Errorf("Couldn't create SQLiteStore: %v", err)
		return
	}
	defer s.Close()

	SqlStoreTest(s, t)

}

func TestPostgresStore(t *testing.T) {
	db, err := sql.Open("postgres", TEST_postgresString)
	if err != nil {
		t.Errorf("Couldn't open database: %v", err)
		return
	}
	db.Exec("DROP TABLE IF EXISTS timebatchtable")
	defer db.Close()
	s, err := OpenSqlStore(db, "postgres", nil)
	if err != nil {
		t.Errorf("Couldn't create PostgresStore: %v", err)
		return
	}
	defer s.Close()

	SqlStoreTest(s, t)

}

//This is a benchmark of how fast we can read out a thousand-datapoint range in chunks of 10 datapoints.
func BenchmarkThousandSQLite(b *testing.B) {
	os.Remove("TESTING_timebatch.db")
	db, err := sql.Open("sqlite3", "TESTING_timebatch.db")
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	defer db.Close()
	s, err := OpenSQLiteStore(db)
	if err != nil {
		b.Errorf("Couldn't create SQLiteStore: %v", err)
		return
	}
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := int64(0); i < 100; i++ {
		s.Append("testkey", CreateDatapointArray(timestamps, data, ""))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _, _ := s.GetByIndex("testkey", 0)

		r.Init()
		for dp, _ := r.Next(); dp != nil; dp, _ = r.Next() {
			dp.Timestamp()
			dp.Data()
		}
		r.Close()
	}
}

func BenchmarkSQLiteInsert(b *testing.B) {
	os.Remove("TESTING_timebatch.db")
	db, err := sql.Open("sqlite3", "TESTING_timebatch.db")
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	defer db.Close()
	s, err := OpenSQLiteStore(db)
	if err != nil {
		b.Errorf("Couldn't create SQLiteStore: %v", err)
		return
	}
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.Append("testkey", CreateDatapointArray(timestamps, data, ""))
	}

}

//This is a benchmark of how fast we can read out a thousand-datapoint range in chunks of 10 datapoints.
func BenchmarkThousandPostgres(b *testing.B) {
	db, err := sql.Open("postgres", TEST_postgresString)
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	db.Exec("DROP TABLE IF EXISTS timebatchtable")
	defer db.Close()
	s, err := OpenPostgresStore(db)
	if err != nil {
		b.Errorf("Couldn't create PostgresStore: %v", err)
		return
	}
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := int64(0); i < 100; i++ {
		s.Append("testkey", CreateDatapointArray(timestamps, data, ""))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _, _ := s.GetByIndex("testkey", 0)

		r.Init()
		for dp, _ := r.Next(); dp != nil; dp, _ = r.Next() {
			dp.Timestamp()
			dp.Data()
		}
		r.Close()
	}
}

func BenchmarkPostgresInsert(b *testing.B) {
	db, err := sql.Open("postgres", TEST_postgresString)
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	db.Exec("DROP TABLE IF EXISTS timebatchtable")
	defer db.Close()
	s, err := OpenPostgresStore(db)
	if err != nil {
		b.Errorf("Couldn't create PostgresStore: %v", err)
		return
	}
	defer s.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.Append("testkey", CreateDatapointArray(timestamps, data, ""))
	}

}
