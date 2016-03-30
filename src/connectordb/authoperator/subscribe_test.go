package authoperator_test

import (
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/users"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthSubscribe(t *testing.T) {
	db.Clear()
	//Let's create a stream
	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "tst", Email: "root@localhost", Password: "mypass", Role: "user", Public: true}}))
	require.NoError(t, db.CreateDevice("tst/tst", &users.DeviceMaker{}))
	require.NoError(t, db.CreateDevice("tst/tst2", &users.DeviceMaker{}))
	require.NoError(t, db.CreateStream("tst/tst/tst", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "string"}`}}))

	// Make sure we can't subscribe to streams we have no access to
	{
		o, err := db.AsDevice("tst/tst2")
		require.NoError(t, err)
		recvchan := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst", recvchan)
		require.Error(t, err)
		_, err = o.Subscribe("tst/tst", recvchan)
		require.Error(t, err)

		_, err = o.Subscribe("tst/tst/tst", recvchan)
		require.Error(t, err)
	}

	// Make sure we can subscribe to streams we do have access to
	{
		require.NoError(t, db.UpdateDevice("tst/tst2", map[string]interface{}{"role": "writer"}))
		o, err := db.AsDevice("tst/tst2")
		require.NoError(t, err)
		recvchan := make(chan messenger.Message, 2)
		recvchan2 := make(chan messenger.Message, 2)
		recvchan3 := make(chan messenger.Message, 2)

		_, err = o.Subscribe("tst", recvchan)
		require.Error(t, err)

		_, err = o.Subscribe("tst/tst", recvchan2)
		require.Error(t, err)
		_, err = o.Subscribe("tst/tst/tst", recvchan3)
		require.NoError(t, err)
	}
	//
	{
		o, err := db.AsDevice("tst/tst2")
		require.NoError(t, err)

		db.Messenger.Flush()

		recvuser := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst", recvuser)
		require.Error(t, err)

		recvdevice := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst/tst", recvdevice)
		require.Error(t, err)

		recvstream := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst/tst/tst", recvstream)
		require.NoError(t, err)

		db.Messenger.Flush()

		data := []datastream.Datapoint{datastream.Datapoint{
			Timestamp: 1.0,
			Data:      "Hello World!",
		}}
		require.NoError(t, o.InsertStream("tst/tst/tst", data, false))
		//We bind a timeout to the channel, since we want the test to fail if no messages come through

		go func() {
			time.Sleep(5 * time.Second)

			// We send a stream with one blank point so we can do
			// easy assert tests rather than require.
			data := []datastream.Datapoint{datastream.Datapoint{}}

			recvstream <- messenger.Message{"TIMEOUT", "", data}
		}()
		m := <-recvstream
		assert.Equal(t, "tst/tst/tst", m.Stream)
		assert.Equal(t, "Hello World!", m.Data[0].Data)
	}
}
