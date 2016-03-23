/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package connectordb

import (
	"connectordb/datastream"
	"connectordb/users"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamTransform(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	//Let's create a stream
	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "tst", Email: "root@localhost", Password: "mypass", Role: "user", Public: true},
		Devices: map[string]*users.DeviceMaker{
			"tst": &users.DeviceMaker{Streams: map[string]*users.StreamMaker{
				"tst": &users.StreamMaker{Stream: users.Stream{
					Schema: `{"type": "number"}`,
				}},
			}},
		},
	}))

	data := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1.0, Data: 1336},
		datastream.Datapoint{Timestamp: 2.0, Data: 3.0},
		datastream.Datapoint{Timestamp: 3.0, Data: 12},
		datastream.Datapoint{Timestamp: 4.0, Data: 1000.0}}
	tdata := datastream.DatapointArray{
		datastream.Datapoint{Timestamp: 1.0, Data: true},
		datastream.Datapoint{Timestamp: 2.0, Data: false},
		datastream.Datapoint{Timestamp: 4.0, Data: true}}
	badtransform := "lt('"
	//transform := "if ($ > 20 and $ < 10) | $ > 300"
	transform := "if $ > 20 or $ < 10 | $ > 300"

	require.NoError(t, db.InsertStream("tst/tst/tst", data, false))

	_, err := db.GetStreamTimeRange("tst/tst/tst", 0.0, 0, 0, badtransform)
	require.Error(t, err)
	_, err = db.GetStreamIndexRange("tst/tst/tst", 0, 0, badtransform)
	require.Error(t, err)

	tr, err := db.GetStreamTimeRange("tst/tst/tst", 0.0, 0, 0, transform)
	require.NoError(t, err)

	for i := 0; i < len(tdata); i++ {
		fmt.Println(i)
		dp, err := tr.Next()
		require.NotNil(t, dp, dp.String())
		require.NoError(t, err)
		require.Equal(t, tdata[i].String(), dp.String())
	}
	dp, err := tr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	tr.Close()

	tr, err = db.GetStreamIndexRange("tst/tst/tst", 0, 0, transform)
	require.NoError(t, err)

	for i := 0; i < len(tdata); i++ {
		dp, err = tr.Next()
		require.NoError(t, err)
		require.Equal(t, tdata[i].String(), dp.String())
	}
	dp, err = tr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)
	tr.Close()

}

func TestStreamIO(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	//Let's create a user/device/stream combo
	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "tst", Email: "root@localhost", Password: "mypass", Role: "user", Public: true},
		Devices: map[string]*users.DeviceMaker{
			"tst": &users.DeviceMaker{Streams: map[string]*users.StreamMaker{
				"tst": &users.StreamMaker{Stream: users.Stream{
					Schema: `{"type": "string"}`,
				}},
			}},
		},
	}))

	//Now make sure that length is 0
	l, err := db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	strm, err := db.ReadStream("tst/tst/tst")
	require.NoError(t, err)
	l, err = db.LengthStreamByID(strm.StreamID, "")

	require.NoError(t, err)
	require.Equal(t, int64(0), l)

	data := datastream.DatapointArray{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      1336,
	}}
	require.Error(t, db.InsertStream("tst/tst/tst", data, false), "insert succeeds on data which does not fit schema")

	data = []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      "Hello World!",
	}}
	require.NoError(t, db.InsertStream("tst/tst/tst", data, false))

	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(1), l)

	data = []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 2.0,
		Data:      "2",
	}, datastream.Datapoint{
		Timestamp: 3.0,
		Data:      "3",
	}}
	require.NoError(t, db.InsertStream("tst/tst/tst", data, false))

	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(3), l)

	dr, err := db.GetStreamTimeRange("tst/tst/tst", 0.0, 2.5, 1, "")
	require.NoError(t, err)

	dp, err := dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, "Hello World!", dp.Data)
	require.Equal(t, 1.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr, err = db.GetShiftedStreamTimeRange("tst/tst/tst", 0.0, 2.5, 1, 1, "")
	require.NoError(t, err)

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

	dr, err = db.GetStreamIndexRange("tst/tst/tst", 0, 2, "")
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

	dr, err = db.GetStreamIndexRange("tst/tst/tst", -1, 0, "")
	require.NoError(t, err)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.NotNil(t, dp)
	require.Equal(t, "3", dp.Data)
	require.Equal(t, 3.0, dp.Timestamp)
	require.Equal(t, "", dp.Sender)

	dp, err = dr.Next()
	require.NoError(t, err)
	require.Nil(t, dp)

	dr.Close()

	dr, err = db.GetStreamIndexRange("tst/tst/tst", -2, -1, "")
	require.NoError(t, err)

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

	i, err := db.TimeToIndexStream("tst/tst/tst", 1.3)
	require.NoError(t, err)
	require.Equal(t, int64(1), i)
	i, err = db.TimeToIndexStream("tst/tst/tst", 0.3)
	require.NoError(t, err)
	require.Equal(t, int64(0), i)

	//Now let's make sure that stuff is deleted correctly
	require.NoError(t, db.DeleteStream("tst/tst/tst"))
	require.NoError(t, db.CreateStream("tst/tst/tst", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "string"}`}}))
	l, err = db.LengthStream("tst/tst/tst")
	require.NoError(t, err)
	require.Equal(t, int64(0), l, "Timebatch has residual data from deleted stream")
}
