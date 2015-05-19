package timebatchdb

import (
	"testing"

	"github.com/stretchr/testify/require"
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
		t.Errorf("%s wrong range returned %d %d", try, len(timestamps), len(data))
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

	require.True(t, CreateDatapointArray(timestamps[0:1], data[0:1], "").IsTimestampOrdered())
	require.True(t, CreateDatapointArray(timestamps[0:3], data[0:3], "").IsTimestampOrdered())

	da := CreateDatapointArray(timestamps, data, "test") //This internally tests fromlist

	require.True(t, assertData(t, da, "creation"))
	require.False(t, da.IsTimestampOrdered(), "Timestamp ordering should fail due to duplicate stamps")

	dplen := da.Size()
	//It looks like the basics are working. Now let's test going to bytes and back
	da.Bytes()
	require.Equal(t, dplen, da.Size())

	//da was reloaded when Bytes() was called. Make sure things are fine
	require.True(t, assertData(t, da, "nochangebytes"))
	require.True(t, assertData(t, DatapointArrayFromBytes(da.Bytes()), "frombytes"))
	require.True(t, assertData(t, DatapointArrayFromCompressedBytes(da.CompressedBytes()), "compressed"))

	//Now check getting by time
	require.Equal(t, 1, da.FindTimeIndex(1200))
	require.Equal(t, 5, da.FindTimeIndex(2000))
	require.Equal(t, -1, da.FindTimeIndex(3000))

	require.Equal(t, 4, da.DatapointTRange(1200, 2000).Len())
	require.Equal(t, 4, len(da.DataTRange(1200, 2000)))
	require.Equal(t, 4, len(da.TimestampTRange(1200, 2000)))

	require.Equal(t, 8, da.DatapointTRange(1200, 3500).Len())
	require.Equal(t, 8, len(da.DataTRange(1200, 3500)))
	require.Equal(t, 8, len(da.TimestampTRange(1200, 3500)))

	require.Equal(t, da.Len(), len(da.DataIRange(0, 50)))
	require.Equal(t, 0, len(da.DataIRange(40, 50)))

	require.Equal(t, da.Len(), len(da.TimestampIRange(0, 50)))
	require.Equal(t, 0, len(da.TimestampIRange(40, 50)))

	datat, datad := da.GetTRange(1200, 3500)
	require.Equal(t, 8, len(datat))
	require.Equal(t, 8, len(datad))

	require.Equal(t, 4, da.TStart(2000).Len())

	dp, err := da.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1000), dp.Timestamp())

	dp, err = da.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1500), dp.Timestamp())

	da.Next()
	da.Next()
	da.Next()
	da.Next()
	da.Next()
	da.Next()
	dp, err = da.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(3000), dp.Timestamp())

	dp, err = da.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	da.Close()
	dp, err = da.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1000), dp.Timestamp())
	da.Close()

	// make sure loading from DataRange is functional
	da2, err := DatapointArrayFromDataRange(da)
	require.NoError(t, err)
	require.True(t, assertData(t, da2, "fromdatarange"))

	// make sure loading from ByteDatapoints is functional
	da3 := DatapointArrayFromByteDatapoints(da.ByteDatapoints())
	require.True(t, assertData(t, da3, "bytedatapoints"))
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
