package datastream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBatch(t *testing.T) {
	sdb.Clear()
	b := []Batch{Batch{"2", "", "1", 0, dpa1}, Batch{"3", "grr", "1", 0, dpa2}}

	require.NoError(t, sdb.WriteBatches(b))

	i, err := sdb.GetEndIndex(2, "")
	require.NoError(t, err)
	require.EqualValues(t, i, len(dpa1))

	i, err = sdb.GetEndIndex(3, "grr")
	require.NoError(t, err)
	require.EqualValues(t, i, len(dpa2))

}

func TestEndIndex(t *testing.T) {
	sdb.Clear()

	i, err := sdb.GetEndIndex(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	require.NoError(t, sdb.Append(1, "", dpa1))

	i, err = sdb.GetEndIndex(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), i)

	require.NoError(t, sdb.Append(1, "", dpa4))

	i, err = sdb.GetEndIndex(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(3), i)

	i, err = sdb.GetEndIndex(1, "sub")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	require.NoError(t, sdb.Append(1, "sub", dpa4))

	i, err = sdb.GetEndIndex(1, "sub")
	require.NoError(t, err)
	require.Equal(t, int64(1), i)
}

func TestSqlRangeNextArray(t *testing.T) {
	sdb.Clear()

	require.NoError(t, sdb.Append(1, "", dpa1))
	require.NoError(t, sdb.Append(1, "", dpa4))

	sr, di, err := sdb.GetByIndex(1, "", 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), di)
	defer sr.Close()

	dpa, err := sr.NextArray()
	require.NoError(t, err)
	require.NotNil(t, dpa)
	require.Equal(t, dpa.String(), dpa1.String())
	dpa, err = sr.NextArray()
	require.NoError(t, err)
	require.NotNil(t, dpa)
	require.Equal(t, dpa.String(), dpa4.String())

	dpa, err = sr.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)

	dpa, err = sr.NextArray()
	require.NoError(t, err)
	require.Nil(t, dpa)
}

func TestSqlRangeNext(t *testing.T) {
	sdb.Clear()

	require.NoError(t, sdb.Append(1, "", dpa1))
	require.NoError(t, sdb.Append(1, "", dpa4))

	sr, di, err := sdb.GetByIndex(1, "", 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), di)
	defer sr.Close()

	dp, err := sr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa1[0].String())

	dp, err = sr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa1[1].String())

	dp, err = sr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, dp.String(), dpa4[0].String())

	dp, err = sr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dp, err = sr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
}

func TestDeleteStream(t *testing.T) {
	sdb.Clear()

	require.NoError(t, sdb.Append(1, "", dpa1))
	require.NoError(t, sdb.Append(1, "", dpa4))
	require.NoError(t, sdb.Append(1, "sub", dpa7))
	require.NoError(t, sdb.Append(2, "", dpa5))

	require.NoError(t, sdb.DeleteStream(1))

	i, err := sdb.GetEndIndex(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	i, err = sdb.GetEndIndex(1, "sub")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	i, err = sdb.GetEndIndex(2, "")
	require.NoError(t, err)
	require.Equal(t, int64(1), i)

}

func TestDeleteSubstream(t *testing.T) {
	sdb.Clear()

	require.NoError(t, sdb.Append(1, "sub", dpa1))
	require.NoError(t, sdb.Append(1, "sub", dpa4))
	require.NoError(t, sdb.Append(1, "", dpa7))
	require.NoError(t, sdb.Append(2, "sub", dpa5))

	require.NoError(t, sdb.DeleteSubstream(1, "sub"))

	i, err := sdb.GetEndIndex(1, "")
	require.NoError(t, err)
	require.Equal(t, int64(9), i)
	i, err = sdb.GetEndIndex(1, "sub")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	i, err = sdb.GetEndIndex(2, "sub")
	require.NoError(t, err)
	require.Equal(t, int64(1), i)

}

func TestGetIndex(t *testing.T) {
	sdb.Clear()

	require.NoError(t, sdb.Append(1, "sub", dpa7))
	sr, di, err := sdb.GetByIndex(1, "", 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), di)
	dp, err := sr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	sr.Close()

	sr, di, err = sdb.GetByIndex(1, "sub", 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), di)
	dpa, err := sr.NextArray()
	sr.Close()
	require.NoError(t, err)
	require.Equal(t, dpa7.String(), dpa.String())

	sr, di, err = sdb.GetByIndex(1, "sub", 9)
	require.NoError(t, err)
	dp, err = sr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	sr.Close()

	sr, di, err = sdb.GetByIndex(1, "sub", 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), di)
	dpa, err = sr.NextArray()
	sr.Close()
	require.NoError(t, err)
	require.Equal(t, dpa7[1:].String(), dpa.String())

}

func TestGetTime(t *testing.T) {
	sdb.Clear()

	require.NoError(t, sdb.Append(1, "sub", dpa7))
	sr, di, err := sdb.GetByTime(1, "", 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), di)
	dp, err := sr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	sr.Close()

	sr, di, err = sdb.GetByTime(1, "sub", 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), di)
	dpa, err := sr.NextArray()
	sr.Close()
	require.NoError(t, err)
	require.Equal(t, dpa7.String(), dpa.String())

	sr, di, err = sdb.GetByTime(1, "sub", 8)
	require.NoError(t, err)
	dp, err = sr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	sr.Close()

	sr, di, err = sdb.GetByTime(1, "sub", 6)
	require.NoError(t, err)
	require.Equal(t, int64(7), di)
	dpa, err = sr.NextArray()
	sr.Close()
	require.NoError(t, err)
	require.Equal(t, dpa7[7:].String(), dpa.String())
}

func BenchmarkSql250Append(b *testing.B) {
	sdb.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		require.NoError(b, sdb.Append(1, "", dpa))
	}
}

func BenchmarkSql250Insert(b *testing.B) {
	sdb.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		require.NoError(b, sdb.Insert(1, "", int64(n*250), dpa))
	}
}

func BenchmarkSql1000Read(b *testing.B) {
	sdb.Clear()

	dpa := make(DatapointArray, 250)
	for i := 0; i < 250; i++ {
		dpa[i] = Datapoint{1.0, true, ""}
	}
	for i := 0; i < 5; i++ {
		require.NoError(b, sdb.Append(1, "", dpa))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _, err := sdb.GetByIndex(1, "", 250)
		require.NoError(b, err)
		for i := 0; i < 4; i++ {
			d, err := r.NextArray()
			require.NoError(b, err)
			require.NotNil(b, d)
		}
		r.Close()
	}
}
