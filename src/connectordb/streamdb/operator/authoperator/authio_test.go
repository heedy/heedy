package authoperator

import (
	"testing"

	"connectordb/config"
	"connectordb/streamdb/datastream"

	"github.com/stretchr/testify/require"
)

func TestAuthStreamIO(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	//Let's create a stream
	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/tst"))

	o, err := db.GetOperator("tst/tst")
	require.NoError(t, err)

	require.NoError(t, o.CreateStream("tst/tst/tst", `{"type": "integer"}`))

	//Now make sure that length is 0
	l, err := o.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	strm, err := o.ReadStream("tst/tst/tst")
	require.NoError(t, err)
	l, err = o.LengthStreamByID(strm.StreamId, "")

	data := []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      1336,
	}}
	require.NoError(t, o.InsertStream("tst/tst/tst", data, false))

	l, err = o.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(1), l)

	dr, err := o.GetStreamTimeRange("tst/tst/tst", 0.0, 2.5, 0)
	require.NoError(t, err)

	dp, err := dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1336), dp.Data.(int64))
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	dr, err = o.GetStreamIndexRange("tst/tst/tst", 0, 1)
	require.NoError(t, err)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1336), dp.Data.(int64))
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	i, err := db.TimeToIndexStream("tst/tst/tst", 0.3)
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	//Now let's make sure that stuff is deleted correctly
	require.NoError(t, o.DeleteStream("tst/tst/tst"))
	require.NoError(t, db.CreateStream("tst/tst/tst", `{"type": "string"}`))
	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l, "Timebatch has residual data from deleted stream")
}

func TestAuthSubstream(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	defer db.Close()
	db.Clear()

	//Let's create a stream
	require.NoError(t, db.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, db.CreateDevice("tst/tst"))
	require.NoError(t, db.CreateDevice("tst/tst2"))
	require.NoError(t, db.CreateStream("tst/tst2/tst", `{"type": "integer"}`))
	s, err := db.ReadStream("tst/tst2/tst")
	require.NoError(t, err)
	s.Downlink = true
	require.NoError(t, db.UpdateStream(s))

	require.NoError(t, db.SetAdmin("tst/tst", true))
	o, err := db.GetOperator("tst/tst")
	require.NoError(t, err)

	data := []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      1336,
	}}
	require.NoError(t, o.InsertStream("tst/tst2/tst", data, false))

	l, err := o.LengthStream("tst/tst2/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	dr, err := o.GetStreamTimeRange("tst/tst2/tst/downlink", 0.0, 2.5, 0)
	require.NoError(t, err)

	dp, err := dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, int64(1336), dp.Data.(int64))
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "tst/tst", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

}
