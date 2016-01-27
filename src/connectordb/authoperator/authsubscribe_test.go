/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator

import (
	"connectordb/datastream"
	"connectordb/operator/interfaces"
	"connectordb/operator/messenger"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthSubscribe(t *testing.T) {
	fmt.Println("test auth subscribe")

	database, baseOperator, err := OpenDb(t)
	require.NoError(t, err)
	defer database.Close()

	//Let's create a stream
	require.NoError(t, baseOperator.CreateUser("tst", "root@localhost", "mypass"))
	require.NoError(t, baseOperator.CreateDevice("tst/tst"))
	require.NoError(t, baseOperator.CreateDevice("tst/tst2"))
	require.NoError(t, baseOperator.CreateStream("tst/tst/tst", `{"type": "string"}`))

	// Make sure we can't subscribe to streams we have no access to
	{
		ao, err := NewDeviceAuthOperator(baseOperator, "tst/tst2")
		require.NoError(t, err)
		o := interfaces.PathOperatorMixin{ao}
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
		ao, err := NewDeviceAuthOperator(baseOperator, "tst/tst")
		require.NoError(t, err)
		o := interfaces.PathOperatorMixin{ao}
		recvchan := make(chan messenger.Message, 2)
		recvchan2 := make(chan messenger.Message, 2)
		recvchan3 := make(chan messenger.Message, 2)

		_, err = o.Subscribe("tst", recvchan)
		require.Error(t, err)

		_, err = o.Subscribe("tst/tst", recvchan2)
		require.NoError(t, err)

		_, err = o.Subscribe("tst/tst/tst", recvchan3)
		require.NoError(t, err)
	}

	//
	{
		baseOperator.SetAdmin("tst/tst", true) //TODO: Subscriptions should be dumped on a permissions change, and that does not happen

		ao, err := NewDeviceAuthOperator(baseOperator, "tst/tst")
		require.NoError(t, err)
		o := interfaces.PathOperatorMixin{ao}

		database.GetMessenger().Flush()

		recvuser := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst", recvuser)
		require.NoError(t, err)

		recvdevice := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst/tst", recvdevice)
		require.NoError(t, err)

		recvstream := make(chan messenger.Message, 2)
		_, err = o.Subscribe("tst/tst/tst", recvstream)
		require.NoError(t, err)

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

			recvuser <- messenger.Message{"TIMEOUT", "", data}
			recvdevice <- messenger.Message{"TIMEOUT", "", data}
			recvstream <- messenger.Message{"TIMEOUT", "", data}
		}()

		m := <-recvuser
		assert.Equal(t, "tst/tst/tst", m.Stream)
		assert.Equal(t, "Hello World!", m.Data[0].Data)

		m = <-recvdevice
		assert.Equal(t, "tst/tst/tst", m.Stream)
		assert.Equal(t, "Hello World!", m.Data[0].Data)

		m = <-recvstream
		assert.Equal(t, "tst/tst/tst", m.Stream)
		assert.Equal(t, "Hello World!", m.Data[0].Data)

	}
}
