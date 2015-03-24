package timebatchdb

import (
	"bytes"
	"errors"
	"math"
	"testing"
)

func ensureValidityTest(t *testing.T, timestamps []int64, data [][]byte, dr DataRange) bool {
	err := dr.Init()
	if err != nil {
		t.Errorf("Failed to initialize DataRange: %v", err)
		return false
	}
	for i := 0; i < len(timestamps); i++ {
		dp, err := dr.Next()
		if dp == nil || err != nil {
			t.Errorf("DataRange terminated too early: %v", err)
			return false
		}
		if dp.Timestamp() != timestamps[i] || bytes.Compare(dp.Data(), data[i]) != 0 {
			t.Errorf("Datapoint incorrect data: %d %v %v", i, dp.Timestamp(), dp.Data())
			return false
		}
	}
	//Now make sure there are no extra datapoints
	dp, err := dr.Next()
	if dp != nil || err != nil {
		t.Errorf("DataRange terminated too early: %v", err)
		return false
	}
	return true
}

func TestRedisCache(t *testing.T) {
	err2 := errors.New("FAILTEST")
	//First try dialing an invalid redis cache
	rc, err := OpenRedisCache("", err2)
	if err != err2 {
		t.Errorf("OpenFail", err)
		return
	}
	//First try dialing an invalid redis cache
	rc, err = OpenRedisCache("localhost:12324", nil)
	if err == nil {
		rc.Close()
		t.Errorf("Open invalid Redis error %v", err)
		return
	}

	rc, err = OpenRedisCache("localhost:6379", nil)
	if err != nil {
		t.Errorf("Open Redis error %v", err)
		return
	}
	defer rc.Close()

	//Cleans the cache
	rc.Clear()

	dp, err := rc.GetMostRecent("hello/world")
	if err != ErrorRedisDNE {
		t.Errorf("Get most recent failed: %v %v", err, dp)
		return
	}
	dp, err = rc.GetOldest("hello/world")
	if err != ErrorRedisDNE {
		t.Errorf("Get oldest failed: %v %v", err, dp)
		return
	}
	tme, err := rc.GetEndTime("hello/world")
	if err != nil || tme != math.MinInt64 {
		t.Errorf("Get most recent failed: %v %v", err, tme)
		return
	}
	tme, err = rc.GetStartTime("hello/world")
	if err != nil || tme != math.MaxInt64 {
		t.Errorf("Get oldest failed: %v %v", err, tme)
		return
	}

	idx, err := rc.EndIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.StartIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.CacheLength("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("get cache length failed %d %v", idx, err)
		return
	}

	//Now insert some data
	timestamps := []int64{1, 2, 3, 4, 5, 6, 6, 7, 8}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	//Test empty key
	dpa, idx, err := rc.Get("hello/world")
	if err != nil || dpa == nil || dpa.Len() != 0 || idx != 0 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}

	keysize, err := rc.Insert("hello/world", CreateDatapointArray(timestamps[:4], data[:4], ""))
	if keysize != 4 || err != nil {
		t.Errorf("Insert error %d %v", keysize, err)
		return
	}

	dpa, idx, err = rc.Get("hello/world")
	if err != nil || dpa == nil || dpa.Len() != 4 || idx != 0 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps[:4], data[:4], dpa) {
		return
	}

	keysize, err = rc.Insert("hello/world", CreateDatapointArray(timestamps[4:], data[4:], ""))
	if keysize != 9 || err != nil {
		t.Errorf("Insert error %d %v", keysize, err)
		return
	}

	dpa, idx, err = rc.Get("hello/world")
	if err != nil || dpa == nil || dpa.Len() != 9 || idx != 0 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps, data, dpa) {
		return
	}

	dpa, idx, err = rc.BatchGet("hello/world", 5)
	if err != nil || dpa == nil || dpa.Len() != 5 || idx != 0 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps[:5], data[:5], dpa) {
		return
	}

	dpa, idx, err = rc.BatchGet("hello/world", 20)
	if err != ErrorRedisWrongSize || dpa == nil || dpa.Len() != 9 || idx != 0 {
		t.Errorf("Get too big error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps, data, dpa) {
		return
	}

	idx, err = rc.EndIndex("hello/world")
	if idx != 9 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.StartIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.CacheLength("hello/world")
	if idx != 9 || err != nil {
		t.Errorf("get cache length failed %d %v", idx, err)
		return
	}

	err = rc.BatchRemove("hello/world", 5)
	if err != nil {
		t.Errorf("Error removing batch: %v", err)
		return
	}

	dpa, idx, err = rc.Get("hello/world")
	if err != nil || dpa == nil || dpa.Len() != 4 || idx != 5 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps[5:], data[5:], dpa) {
		return
	}
	dr, idx, err := rc.GetByIndex("hello/world", 5)
	if err != nil || idx != 5 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps[5:], data[5:], dr) {
		return
	}
	dr, idx, err = rc.GetByIndex("hello/world", 6)
	if err != nil || idx != 6 {
		t.Errorf("Get error %d %v", idx, err)
		return
	}
	if !ensureValidityTest(t, timestamps[6:], data[6:], dr) {
		return
	}
	dr, idx, err = rc.GetByIndex("hello/world", 20)
	if val, _ := dr.Next(); err != nil || idx != 20 || val != nil {
		t.Errorf("Get error %d %v", idx, err)
		return
	}

	idx, err = rc.EndIndex("hello/world")
	if idx != 9 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.StartIndex("hello/world")
	if idx != 5 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.CacheLength("hello/world")
	if idx != 4 || err != nil {
		t.Errorf("get cache length failed %d %v", idx, err)
		return
	}

	dp, err = rc.GetMostRecent("hello/world")
	if err != nil || dp.Timestamp() != 8 {
		t.Errorf("Get most recent failed: %v %v", err, dp)
		return
	}
	tme, err = rc.GetEndTime("hello/world")
	if err != nil || tme != 8 {
		t.Errorf("Get most recent failed: %v %v", err, tme)
		return
	}
	dp, err = rc.GetOldest("hello/world")
	if err != nil || dp.Timestamp() != 6 {
		t.Errorf("Get most recent failed: %v %v", err, dp)
		return
	}
	tme, err = rc.GetStartTime("hello/world")
	if err != nil || tme != 6 {
		t.Errorf("Get most recent failed: %v %v", err, tme)
		return
	}

	//Now check if the queue of batches works
	err = rc.BatchPush("hello/world")
	if err != nil {
		t.Errorf("Error pushing batch: %v", err)
		return
	}

	//Now check if the queue of batches works
	k, err := rc.BatchWait()
	if err != nil || k != "hello/world" {
		t.Errorf("Error in batch queue: %v, %s", err, k)
		return
	}

	//Lastly, make sure that deleting works
	rc.Delete("hello/world")

	dp, err = rc.GetMostRecent("hello/world")
	if err != ErrorRedisDNE {
		t.Errorf("Get most recent failed: %v %v", err, dp)
		return
	}
	dp, err = rc.GetOldest("hello/world")
	if err != ErrorRedisDNE {
		t.Errorf("Get oldest failed: %v %v", err, dp)
		return
	}
	tme, err = rc.GetEndTime("hello/world")
	if err != nil || tme != math.MinInt64 {
		t.Errorf("Get most recent failed: %v %v", err, tme)
		return
	}
	tme, err = rc.GetStartTime("hello/world")
	if err != nil || tme != math.MaxInt64 {
		t.Errorf("Get oldest failed: %v %v", err, tme)
		return
	}

	idx, err = rc.EndIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.StartIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.CacheLength("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("get cache length failed %d %v", idx, err)
		return
	}

	keysize, err = rc.Insert("hello/world", CreateDatapointArray(timestamps, data, ""))
	if keysize != 9 || err != nil {
		t.Errorf("Insert error %d %v", keysize, err)
		return
	}

	rc.DeletePrefix("hi")
	dp, err = rc.GetMostRecent("hello/world")
	if err != nil {
		t.Errorf("Get most recent failed: %v", err)
		return
	}

	//Lastly, make sure that deleting works
	rc.DeletePrefix("hello")

	dp, err = rc.GetMostRecent("hello/world")
	if err != ErrorRedisDNE {
		t.Errorf("Get most recent failed: %v %v", err, dp)
		return
	}
	dp, err = rc.GetOldest("hello/world")
	if err != ErrorRedisDNE {
		t.Errorf("Get oldest failed: %v %v", err, dp)
		return
	}
	tme, err = rc.GetEndTime("hello/world")
	if err != nil || tme != math.MinInt64 {
		t.Errorf("Get most recent failed: %v %v", err, tme)
		return
	}
	tme, err = rc.GetStartTime("hello/world")
	if err != nil || tme != math.MaxInt64 {
		t.Errorf("Get oldest failed: %v %v", err, tme)
		return
	}

	idx, err = rc.EndIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.StartIndex("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("Index get failed %d %v", idx, err)
		return
	}
	idx, err = rc.CacheLength("hello/world")
	if idx != 0 || err != nil {
		t.Errorf("get cache length failed %d %v", idx, err)
		return
	}
}

//This is a benchmark of how fast we can read out a thousand-datapoint range in chunks of 10 datapoints.
func BenchmarkThousandRedis(b *testing.B) {
	rc, err := OpenRedisCache("localhost:6379", nil)
	if err != nil {
		b.Errorf("Open Redis error %v", err)
		return
	}
	defer rc.Close()

	//Cleans the cache
	rc.Clear()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := int64(0); i < 100; i++ {
		_, err = rc.Insert("testkey", CreateDatapointArray(timestamps, data, ""))
		if err != nil {
			b.Errorf("Insert Error: %v", err)
			return
		}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _, _ := rc.Get("testkey")

		r.Init()
		for dp, _ := r.Next(); dp != nil; dp, _ = r.Next() {
			dp.Timestamp()
			dp.Data()
		}
		r.Close()
	}
}

func BenchmarkRedisInsert(b *testing.B) {
	rc, err := OpenRedisCache("localhost:6379", nil)
	if err != nil {
		b.Errorf("Open Redis error %v", err)
		return
	}
	defer rc.Close()

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if n%10000 == 0 {
			rc.Clear() //Make sure we don't overflow the ram
		}
		rc.Insert("testkey", CreateDatapointArray(timestamps, data, ""))

	}
}
