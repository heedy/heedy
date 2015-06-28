package datastream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisBasics(t *testing.T) {
	rc, err := NewRedisConnection(&testOptions)
	require.NoError(t, err)
	defer rc.Close()

	require.NoError(t, rc.Clear())

	require.NoError(t, rc.DeleteStream("mystream"))

	i, err := rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	require.NoError(t, rc.Insert("mystream", "", dpa6, false))

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	require.NoError(t, rc.DeleteStream("mystream"))

	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
}

func TestRedisInsert(t *testing.T) {
	rc, err := NewRedisConnection(&testOptions)
	require.NoError(t, err)
	defer rc.Close()

	require.NoError(t, rc.Clear())

	require.NoError(t, rc.Insert("mystream", "", dpa6, false))

	dpatest, err := rc.Get("mystream", "")
	require.NoError(t, err)
	require.True(t, dpa6.IsEqual(dpatest))

	require.EqualError(t, rc.Insert("mystream", "", dpa1, false), ErrTimestamp.Error())

	dpz := DatapointArray{Datapoint{5.0, "helloWorld", "me"}, Datapoint{6.0, "helloWorld2", "me2"}}
	require.NoError(t, rc.Insert("mystream", "", dpz, false))

	i, err := rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(7), i)

	//Now we must test an internal quirk in the redis lua code: inserting more than
	// 5k chunks.
	dpz = make(DatapointArray, 1, 6000)
	dpz[0] = Datapoint{9.0, "ol", ""}
	for iter := 1; iter < 6000; iter++ {
		dpz = append(dpz, Datapoint{10.0 + float64(iter), true, ""})
	}
	require.NoError(t, rc.Insert("mystream", "", dpz, false))
	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(6007), i)
}

func TestRedisRestamp(t *testing.T) {
	rc, err := NewRedisConnection(&testOptions)
	require.NoError(t, err)
	defer rc.Close()

	require.NoError(t, rc.Clear())

	require.NoError(t, rc.Insert("mystream", "", dpa6, false))
	require.NoError(t, rc.Insert("mystream", "", dpa1, true))

	restampedDpa1 := make(DatapointArray, 2)
	copy(restampedDpa1, dpa1)

	restampedDpa1[0].Timestamp = 5.0
	restampedDpa1[1].Timestamp = 5.0

	dpatest, err := rc.Get("mystream", "")
	require.NoError(t, err)
	require.Equal(t, restampedDpa1.String(), dpatest[5:].String())
}

func TestSubstream(t *testing.T) {
	rc, err := NewRedisConnection(&testOptions)
	require.NoError(t, err)
	defer rc.Close()

	require.NoError(t, rc.Clear())

	require.NoError(t, rc.Insert("mystream", "s1", dpa6, false))

	require.EqualError(t, rc.Insert("mystream", "s1", dpa1, false), ErrTimestamp.Error())

	i, err := rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	i, err = rc.StreamLength("mystream", "s1")
	require.NoError(t, err)
	require.Equal(t, int64(5), i)

	require.NoError(t, rc.Insert("mystream", "", dpa1, false))

	require.NoError(t, rc.DeleteSubstream("mystream", "s1"))
	i, err = rc.StreamLength("mystream", "")
	require.NoError(t, err)
	require.Equal(t, int64(2), i)
	i, err = rc.StreamLength("mystream", "s1")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

}

func BenchmarkRedis1Insert(b *testing.B) {
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Insert("mystream", "", DatapointArray{Datapoint{float64(n), true, ""}}, false)
	}
}

func BenchmarkRedis1InsertRestamp(b *testing.B) {
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

	rc.Insert("mystream", "", DatapointArray{Datapoint{2.0, true, ""}}, false)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rc.Insert("mystream", "", DatapointArray{Datapoint{1.0, true, ""}}, true)
	}
}

func BenchmarkRedis1InsertParallel(b *testing.B) {
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rc.Insert("mystream", "", DatapointArray{Datapoint{1.0, true, ""}}, false)
		}
	})
}

func BenchmarkRedis1000Insert(b *testing.B) {
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

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
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

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
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

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
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

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
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

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
	rc, err := NewRedisConnection(&testOptions)
	if err != nil {
		b.Error(err)
		return
	}
	rc.Clear()
	defer rc.Close()

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
