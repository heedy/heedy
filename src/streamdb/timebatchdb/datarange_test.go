package timebatchdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRangeList(t *testing.T) {
	//DataRange can't handle same-timestamp values
	//timestamps := []int64{1000,1500,2000,2000,2000,2500,3000,3000,3000}
	timestamps := []int64{1, 2, 3, 4, 5, 6, 3000, 3100, 3200}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	da := CreateDatapointArray(timestamps[:5], data[:5], "")
	db := CreateDatapointArray(timestamps[5:], data[5:], "")

	rl := NewRangeList()
	rl.Append(da)
	rl.Append(db)
	require.NoError(t, rl.Init())

	da, err := DatapointArrayFromDataRange(rl)
	require.NoError(t, err)
	require.Equal(t, 9, da.Len())

	timestamps = da.Timestamps()
	require.Equal(t, int64(1), timestamps[0])
	require.Equal(t, int64(2), timestamps[1])
	require.Equal(t, int64(3200), timestamps[8])

	//The Close method was not tested at all
	rl3 := NewRangeList()

	rl3.Init()
	dpt, err := rl3.Next()
	require.NoError(t, err)
	require.Nil(t, dpt)

	rl2 := NewRangeList()
	rl2.Append(da)
	rl2.Append(db)
	require.NoError(t, rl2.Init())
	rl2.Close()

}

func TestTimeRange(t *testing.T) {
	timestamps := []int64{1, 2, 3, 4, 5, 6, 6, 7, 8}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	da := CreateDatapointArray(timestamps, data, "")

	tr := NewTimeRange(da, 3, 6)
	defer tr.Close()
	require.NoError(t, tr.Init())

	dp, err := tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(4), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(5), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(6), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(6), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
}

func TestNumRange(t *testing.T) {
	timestamps := []int64{1, 2, 3, 4, 5, 6, 6, 7, 8}
	data := [][]byte{[]byte("test0"), []byte("test1"), []byte("test2"), []byte("test3"),
		[]byte("test4"), []byte("test5"), []byte("test6"), []byte("test7"), []byte("test8")}

	da := CreateDatapointArray(timestamps, data, "")

	tr := NewNumRange(da, 5)
	defer tr.Close()
	require.NoError(t, tr.Init())

	dp, err := tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1), dp.Timestamp())

	err = tr.Skip(2)
	require.NoError(t, err)

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(4), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(5), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(6), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(6), dp.Timestamp())

	dp, err = tr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
}

//This is just to increase test coverage...
func TestEmptyRange(t *testing.T) {
	er := EmptyRange{}
	require.NoError(t, er.Init())

	dp, err := er.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

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
			rl.Append(CreateDatapointArray(timestamps, data, ""))
		}
		rl.Init()
		for dp, _ := rl.Next(); dp != nil; dp, _ = rl.Next() {
			dp.Timestamp()
			dp.Data()
		}
		rl.Close()
	}
}
