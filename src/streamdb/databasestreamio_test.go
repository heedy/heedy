package streamdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamIO(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()

	//Let's create a stream
	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/tst"))
	require.NoError(t, db.CreateStream("tst/tst/tst", `{"type": "string"}`))

	//Now make sure that length is 0
	l, err := db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	strm, err := db.ReadStream("tst/tst/tst")
	require.NoError(t, err)
	l, err = db.LengthStreamByID(strm.StreamId)

	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	data := []Datapoint{Datapoint{
		Timestamp: 1.0,
		Data:      1336,
	}}
	require.Error(t, db.InsertStream("tst/tst/tst", data), "insert succeeds on data which does not fit schema")

	data = []Datapoint{Datapoint{
		Timestamp: 1.0,
		Data:      "Hello World!",
	}}
	require.NoError(t, db.InsertStream("tst/tst/tst", data))

	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(1), l)

	data = []Datapoint{Datapoint{
		Timestamp: 2.0,
		Data:      "2",
	}, Datapoint{
		Timestamp: 3.0,
		Data:      "3",
	}}
	require.NoError(t, db.InsertStream("tst/tst/tst", data))

	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(3), l)

	dr, err := db.GetStreamTimeRange("tst/tst/tst", 0.0, 2.5)
	require.NoError(t, err)

	dp, err := dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, "Hello World!", dp.Data)
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, "2", dp.Data)
	require.Equal(t, 2.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	dr, err = db.GetStreamIndexRange("tst/tst/tst", 0, 2)
	require.NoError(t, err)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, "Hello World!", dp.Data)
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, "2", dp.Data)
	require.Equal(t, 2.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	//Now let's make sure that stuff is deleted correctly
	require.NoError(t, db.DeleteStream("tst/tst/tst"))
	require.NoError(t, db.CreateStream("tst/tst/tst", `{"type": "string"}`))
	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l, "Timebatch has residual data from deleted stream")
}
