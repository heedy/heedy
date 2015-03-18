package timebatchdb

import (
	"testing"
)

//WARNING: This function is used in tests of datarange also
func assertData(t *testing.T, da *DatapointArray, try string) bool {
	if da.Len() != 9 {
		t.Errorf("%s: DatapointArray length: %d", try, da.Len())
		return false
	}

	timestamps := da.Timestamps()

	if timestamps[0] != 1000 || timestamps[1] != 1500 || timestamps[8] != 3000 {
		t.Errorf("%s: DatapointArray timestamp fail1: %d %d", try, timestamps[0], timestamps[8])
		return false
	}

	timestamps, data := da.Get()
	if len(timestamps) != 9 || len(data) != 9 {
		t.Errorf("%s wrong range returned %d %d", len(timestamps), len(data))
		return false
	}
	if timestamps[0] != 1000 || timestamps[1] != 1500 || timestamps[8] != 3000 {
		t.Errorf("%s: DatapointArray timestamp fail: %d %d", try, timestamps[0], timestamps[8])
		return false
	}

	if string(data[0]) != "test0" || string(data[1]) != "test1" || string(data[8]) != "test8" || len(data) != 9 {
		t.Errorf("%s: DatapointArray timestamp fail: %d", try, len(timestamps))
		return false
	}

	if da.Datapoints[0].Key() != "test" {
		t.Errorf("%s: DatapointArray data key fail: %d", try, len(timestamps))
		return false
	}
	return true
}

func TestDatapointArray(t *testing.T) {
	timestamps := []int64{1000, 1500, 2000, 2000, 2000, 2500, 3000, 3000, 3000}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	if !CreateDatapointArray(timestamps[0:1], data[0:1], "").IsTimestampOrdered() {
		t.Errorf("Timestamp ordering failure")
	}
	if !CreateDatapointArray(timestamps[0:3], data[0:3], "").IsTimestampOrdered() {
		t.Errorf("Timestamp ordering failure")
	}

	da := CreateDatapointArray(timestamps, data, "test") //This internally tests fromlist

	if !assertData(t, da, "creation") {
		return
	}

	if da.IsTimestampOrdered() {
		t.Errorf("Timestamp ordering should fail due to duplicate stamps")
	}

	dplen := da.Size()
	//It looks like the basics are working. Now let's test going to bytes and back
	da.Bytes()
	if dplen != da.Size() {
		t.Errorf("Error finding size of dataArray")
	}

	//da was reloaded when Bytes() was called. Make sure things are fine
	if !assertData(t, da, "nochangebytes") {
		return
	}

	//Now test da2
	if !assertData(t, DatapointArrayFromBytes(da.Bytes()), "frombytes") {
		return
	}

	if !assertData(t, DatapointArrayFromCompressedBytes(da.CompressedBytes()), "compressed") {
		return
	}

	//Now check getting by time
	i := da.FindTimeIndex(1200)
	if i != 1 {
		t.Errorf("Error in findtimeindex: %d", i)
	}

	i = da.FindTimeIndex(2000)
	if i != 5 {
		t.Errorf("Error in findtimeindex: %d", i)
	}

	i = da.FindTimeIndex(3000)
	if i != -1 {
		t.Errorf("Error in findtimeindex: %d", i)
		return
	}

	if da.DatapointTRange(1200, 2000).Len() != 4 {
		t.Errorf("Wrong TRange")
		return
	}
	if len(da.DataTRange(1200, 2000)) != 4 {
		t.Errorf("Wrong TRange")
		return
	}
	if len(da.TimestampTRange(1200, 2000)) != 4 {
		t.Errorf("Wrong TRange")
		return
	}
	if da.DatapointTRange(1200, 3500).Len() != 8 {
		t.Errorf("Wrong TRange")
		return
	}
	if len(da.DataTRange(1200, 3500)) != 8 {
		t.Errorf("Wrong TRange")
		return
	}
	if len(da.TimestampTRange(1200, 3500)) != 8 {
		t.Errorf("Wrong TRange")
		return
	}

	if len(da.DataIRange(0, 50)) != da.Len() {
		t.Errorf("DataIRange doesn't return correct number of values")
		return
	}
	if len(da.DataIRange(40, 50)) != 0 {
		t.Errorf("DataIRange doesn't return correct number of values")
		return
	}
	if len(da.TimestampIRange(0, 50)) != da.Len() {
		t.Errorf("DataIRange doesn't return correct number of values")
		return
	}
	if len(da.TimestampIRange(40, 50)) != 0 {
		t.Errorf("DataIRange doesn't return correct number of values")
		return
	}

	datat, datad := da.GetTRange(1200, 3500)
	if len(datat) != 8 || len(datad) != 8 {
		t.Errorf("Wrong TRange")
		return
	}

	if da.TStart(2000).Len() != 4 {
		t.Errorf("Wrong TStart")
		return
	}

	dp, err := da.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 1000 {
		t.Errorf("Iterator wrong")
	}
	dp, err = da.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 1500 {
		t.Errorf("Iterator wrong")
	}
	da.Next()
	da.Next()
	da.Next()
	da.Next()
	da.Next()
	da.Next()
	dp, err = da.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 3000 {
		t.Errorf("Iterator wrong")
	}
	dp, err = da.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp != nil {
		t.Errorf("Iterator wrong")
	}
	da.Close()
	dp, err = da.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 1000 {
		t.Errorf("Iterator wrong")
	}
	da.Close()
	//Lastly, make sure loading from DataRange is functional
	da2, err := DatapointArrayFromDataRange(da)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if !assertData(t, da2, "fromdatarange") {
		return
	}
}

func BenchmarkDatapointArrayRange(b *testing.B) {
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}

	da := CreateDatapointArray(timestamps, data, "")

	for n := 0; n < b.N; n++ {
		da.Init()
		for dp, _ := da.Next(); dp != nil; dp, _ = da.Next() {
			dp.Timestamp()
			dp.Data()
		}
		da.Close()
	}
}

func BenchmarkDatapointArrayByteConversion(b *testing.B) {
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}

	for n := 0; n < b.N; n++ {
		da := CreateDatapointArray(timestamps, data, "")
		da.Bytes()
	}
}

func BenchmarkDatapointArrayCompress(b *testing.B) {
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}

	for n := 0; n < b.N; n++ {
		CreateDatapointArray(timestamps, data, "test").CompressedBytes()
	}
}

func BenchmarkDatapointArrayUncompress(b *testing.B) {
	timestamps := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}
	dpb := CreateDatapointArray(timestamps, data, "test").CompressedBytes()
	for n := 0; n < b.N; n++ {
		DatapointArrayFromCompressedBytes(dpb)
	}
}
