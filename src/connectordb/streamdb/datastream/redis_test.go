package datastream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisBasics(t *testing.T) {

	require.NoError(t, rc.Clear())

	require.NoError(t, rc.DeleteStream("mystream"))

	i, err := rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	i, err = rc.Insert("mystream", "", dpa6, false)
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	require.NoError(t, rc.DeleteStream("mystream"))

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
}

func TestRedisInsert(t *testing.T) {

	require.NoError(t, rc.Clear())

	_, err := rc.Insert("mystream", "", dpa6, false)
	require.NoError(t, err)

	dpatest, err := rc.Get("mystream", "")
	require.NoError(t, err)
	require.True(t, dpa6.IsEqual(dpatest))

	_, err = rc.Insert("mystream", "", dpa1, false)
	require.EqualError(t, err, ErrTimestamp.Error())

	dpz := DatapointArray{Datapoint{5.0, "helloWorld", "me"}, Datapoint{6.0, "helloWorld2", "me2"}}
	i, err := rc.Insert("mystream", "", dpz, false)
	require.NoError(t, err)
	require.Equal(t, int64(7), i)

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(7), i)

	//Now we must test an internal quirk in the redis lua code: inserting more than
	// 5k chunks.
	dpz = make(DatapointArray, 1, 6000)
	dpz[0] = Datapoint{9.0, "ol", ""}
	for iter := 1; iter < 6000; iter++ {
		dpz = append(dpz, Datapoint{10.0 + float64(iter), true, ""})
	}
	i, err = rc.Insert("mystream", "", dpz, false)
	require.NoError(t, err)
	require.Equal(t, int64(6007), i)

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(6007), i)
}

func TestRedisRestamp(t *testing.T) {

	require.NoError(t, rc.Clear())

	_, err := rc.Insert("mystream", "", dpa6, false)
	require.NoError(t, err)
	_, err = rc.Insert("mystream", "", dpa1, true)
	require.NoError(t, err)

	restampedDpa1 := make(DatapointArray, 2)
	copy(restampedDpa1, dpa1)

	restampedDpa1[0].Timestamp = 5.0
	restampedDpa1[1].Timestamp = 5.0

	dpatest, err := rc.Get("mystream", "")
	require.NoError(t, err)
	require.Equal(t, restampedDpa1.String(), dpatest[5:].String())
}

func TestRedisSubstream(t *testing.T) {

	require.NoError(t, rc.Clear())

	i, err := rc.Insert("mystream", "s1", dpa6, false)
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	_, err = rc.Insert("mystream", "s1", dpa1, false)
	require.EqualError(t, err, ErrTimestamp.Error())

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	i, err = rc.StreamLength("mystream", "s1")
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	i, err = rc.Insert("mystream", "", dpa1, false)
	require.NoError(t, err)
	require.Equal(t, int64(2), i)

	require.NoError(t, rc.DeleteSubstream("mystream", "s1"))
	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(2), i)
	i, err = rc.StreamLength("mystream", "s1")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

}

func TestRedisTrim(t *testing.T) {
	require.NoError(t, rc.Clear())
	i, err := rc.Insert("mystream", "", dpa7, false)
	require.NoError(t, err)
	require.EqualValues(t, 9, i)

	require.NoError(t, rc.TrimStream("mystream", "", 2))

	dpa, err := rc.Get("mystream", "")
	require.NoError(t, err)
	require.Equal(t, dpa7[2:].String(), dpa.String())

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.EqualValues(t, 9, i)

	require.NoError(t, rc.TrimStream("mystream", "", 1))

	dpa, err = rc.Get("mystream", "")
	require.NoError(t, err)
	require.Equal(t, dpa7[2:].String(), dpa.String())

	require.NoError(t, rc.TrimStream("mystream", "", 2))

	dpa, err = rc.Get("mystream", "")
	require.NoError(t, err)
	require.Equal(t, dpa7[2:].String(), dpa.String())

	require.NoError(t, rc.TrimStream("mystream", "", 3))

	dpa, err = rc.Get("mystream", "")
	require.NoError(t, err)
	require.Equal(t, dpa7[3:].String(), dpa.String())
}

func TestRedisRange(t *testing.T) {
	require.NoError(t, rc.Clear())
	i, err := rc.Insert("mystream", "", dpa7, false)
	require.NoError(t, err)
	require.EqualValues(t, 9, i)

	dpa, i1, i2, err := rc.Range("systream", "s1", 1, 8)
	require.Error(t, err)

	dpa, i1, i2, err = rc.Range("systream", "s1", 0, 8)
	require.NoError(t, err)
	require.EqualValues(t, 0, i1)
	require.EqualValues(t, 0, i2)

	dpa, i1, i2, err = rc.Range("mystream", "", 2, 8)
	require.NoError(t, err)
	require.EqualValues(t, 2, i1)
	require.EqualValues(t, 8, i2)
	require.Equal(t, dpa7[2:8].String(), dpa.String())

	dpa, i1, i2, err = rc.Range("mystream", "", 0, 0)
	require.NoError(t, err)
	require.EqualValues(t, 0, i1)
	require.EqualValues(t, 9, i2)
	require.Equal(t, dpa7.String(), dpa.String())

	dpa, i1, i2, err = rc.Range("mystream", "", -2, -1)
	require.NoError(t, err)
	require.EqualValues(t, 7, i1)
	require.EqualValues(t, 8, i2)
	require.Equal(t, dpa7[7:8].String(), dpa.String())

	dpa, i1, i2, err = rc.Range("mystream", "", -2, 20)
	require.NoError(t, err)
	require.EqualValues(t, 7, i1)
	require.EqualValues(t, 9, i2)
	require.Equal(t, dpa7[7:].String(), dpa.String())

	dpa, i1, i2, err = rc.Range("mystream", "", -20, 0)
	require.Error(t, err)

	//Now trim the range, to make sure that correct values
	//are returned if not all data is in redis
	require.NoError(t, rc.TrimStream("mystream", "", 3))
	dpa, i1, i2, err = rc.Range("mystream", "", 3, 0)
	require.NoError(t, err)
	require.EqualValues(t, 3, i1)
	require.EqualValues(t, 9, i2)
	require.Equal(t, dpa7[3:].String(), dpa.String())

	dpa, i1, i2, err = rc.Range("mystream", "", 2, 0)
	require.NoError(t, err)
	require.EqualValues(t, 2, i1)
	require.EqualValues(t, 9, i2)
	require.Nil(t, dpa)
}

func BenchmarkRedis1Insert(b *testing.B) {
	rc.Clear()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Insert("mystream", "", DatapointArray{Datapoint{float64(n), true, ""}}, false)
	}
}

func BenchmarkRedis1InsertRestamp(b *testing.B) {
	rc.Clear()

	rc.Insert("mystream", "", DatapointArray{Datapoint{2.0, true, ""}}, false)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Insert("mystream", "", DatapointArray{Datapoint{1.0, true, ""}}, true)
	}
}

func BenchmarkRedis1InsertParallel(b *testing.B) {

	rc.Clear()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rc.Insert("mystream", "", DatapointArray{Datapoint{1.0, true, ""}}, false)
		}
	})
}

func BenchmarkRedis1000Insert(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 1000)
	for i := 0; i < 1000; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Insert("mystream", "", dpa, false)
	}
}

func BenchmarkRedis1000InsertParallel(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 1000)
	for i := 0; i < 1000; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rc.Insert("mystream", "", dpa, false)
		}
	})
}

func BenchmarkRedis1000InsertRestamp(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 1000)
	for i := 0; i < 1000; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}

	rc.Insert("mystream", "", DatapointArray{Datapoint{9000000.0, true, ""}}, false)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Insert("mystream", "", dpa, true)
	}
}

func BenchmarkRedisStreamLength(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 1000)
	for i := 0; i < 1000; i++ {
		dpa[i] = Datapoint{float64(i), true, ""}
	}
	rc.Insert("mystream", "", dpa, false)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		rc.StreamLength("mystream", "")
	}
}

func BenchmarkRedis1000Get(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 1000)
	for i := 0; i < 1000; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	rc.Insert("mystream", "", dpa, false)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Get("mystream", "")
	}
}

func BenchmarkRedis250Get(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	rc.Insert("mystream", "", dpa, false)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Get("mystream", "")
	}
}

func BenchmarkRedis250Range(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	rc.Insert("mystream", "", dpa, false)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Range("mystream", "", 0, 0)
	}

}

func BenchmarkRedis250RangeMiss(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	rc.Insert("mystream", "", dpa, false)

	rc.TrimStream("mystream", "", 4)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Range("mystream", "", 0, 0)
	}

}

func BenchmarkRedis10Range(b *testing.B) {
	rc.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	rc.Insert("mystream", "", dpa, false)
	rc.TrimStream("mystream", "", 4)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Range("mystream", "", -10, 0)
	}

}
