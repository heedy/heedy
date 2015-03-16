package timebatchdb

import (
	"testing"
)

func TestRangeList(t *testing.T) {
	//DataRange can't handle same-timestamp values
	//timestamps := []int64{1000,1500,2000,2000,2000,2500,3000,3000,3000}
	timestamps := []int64{1, 2, 3, 4, 5, 6, 3000, 3100, 3200}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	da := CreateDatapointArray(timestamps[:5], data[:5])
	db := CreateDatapointArray(timestamps[5:], data[5:])

	rl := NewRangeList()
	rl.Append(da)
	rl.Append(db)
	err := rl.Init()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	da, err = DatapointArrayFromDataRange(rl)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if da.Len() != 9 {
		t.Errorf(" DatapointArray length: %d", da.Len())
		return
	}

	timestamps = da.Timestamps()

	if timestamps[0] != 1 || timestamps[1] != 2 || timestamps[8] != 3200 {
		t.Errorf("Datarange timestamp fail1: %d %d", timestamps[0], timestamps[8])
		return
	}

	//The Close method was not tested at all
	rl3 := NewRangeList()

	rl3.Init()
	dpt, err := rl3.Next()
	if dpt != nil || err != nil {
		t.Errorf("Next value of empty list had weird result")
		return
	}
	rl2 := NewRangeList()
	rl2.Append(da)
	rl2.Append(db)
	err = rl2.Init()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	rl2.Close()

}

func TestTimeRange(t *testing.T) {
	timestamps := []int64{1, 2, 3, 4, 5, 6, 6, 7, 8}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	da := CreateDatapointArray(timestamps, data)

	tr := NewTimeRange(da, 3, 6)
	defer tr.Close()
	err := tr.Init()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	dp, err := tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 4 {
		t.Errorf("TimeRange start time incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 5 {
		t.Errorf("TimeRange incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 6 {
		t.Errorf("TimeRange incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 6 {
		t.Errorf("TimeRange incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp != nil {
		t.Errorf("TimeRange endtime incorrect")
	}
}

func TestNumRange(t *testing.T) {
	timestamps := []int64{1, 2, 3, 4, 5, 6, 6, 7, 8}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	da := CreateDatapointArray(timestamps, data)

	tr := NewNumRange(da, 5)
	defer tr.Close()
	err := tr.Init()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	dp, err := tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 1 {
		t.Errorf("NumRange start time incorrect")
	}
	err = tr.Skip(2)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 4 {
		t.Errorf("NumRange start time incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 5 {
		t.Errorf("NumRange incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 6 {
		t.Errorf("NumRange incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp == nil || dp.Timestamp() != 6 {
		t.Errorf("NumRange incorrect")
	}
	dp, err = tr.Next()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dp != nil {
		t.Errorf("NumRange endtime incorrect")
	}
}

//This is just to increase test coverage...
func TestEmptyRange(t *testing.T) {
	er := EmptyRange{}
	err := er.Init()
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	dp, err := er.Next()
	if err != nil || dp != nil {
		t.Errorf("EmptyRange datapoint next error: %s %v", err, dp)
		return
	}
	er.Close()
}

//This is a baseline of how fast we can read out a thousand-datapoint range in chunks of 10 datapoints.
//it isn't a perfect test, since we have to create the data in the mean time, but it still gives an idea of what to expect.
func BenchmarkThousandRangeList(b *testing.B) {

	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8"), []byte("test9")}

	for n := 0; n < b.N; n++ {
		rl := NewRangeList()
		for i := int64(0); i < 100; i++ {
			timestamps := []int64{i, i + 1, i + 1, i + 3, i + 4, i + 5, i + 6, i + 7, i + 8, i + 9}
			rl.Append(CreateDatapointArray(timestamps, data))
		}
		rl.Init()
		for dp, _ := rl.Next(); dp != nil; dp, _ = rl.Next() {
			dp.Timestamp()
			dp.Data()
		}
		rl.Close()
	}
}
