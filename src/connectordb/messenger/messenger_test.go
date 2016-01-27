/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package messenger

import (
	"config"
	"connectordb/datastream"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMessenger(t *testing.T) {
	newerr := errors.New("FAIL")
	_, err := ConnectMessenger(nil, newerr)
	if err != newerr {
		t.Errorf("Error chain failed: %s", err)
		return
	}

	msg, err := ConnectMessenger(&config.TestOptions.NatsOptions, nil)
	require.NoError(t, err)
	defer msg.Close()

	msg2, err := ConnectMessenger(&config.TestOptions.NatsOptions, nil)
	require.NoError(t, err)
	defer msg2.Close()

	recvchan := make(chan Message)

	//We bind a timeout to the channel, since we want the test to fail if no messages come through
	go func() {
		time.Sleep(2 * time.Second)
		recvchan <- Message{"TIMEOUT", "", []datastream.Datapoint{}}
	}()

	_, err = msg2.Subscribe("user1/device1/stream1", recvchan)
	require.NoError(t, err)

	//The connection needs to be flushed so that we are definitely subscribed to the channel
	//before we publish on it
	msg2.Flush()

	//Now, publish a message
	err = msg.Publish("user1/device1/stream1/", Message{"user1/device1/stream1", "", []datastream.Datapoint{datastream.Datapoint{Data: "Hello"}}})
	require.NoError(t, err)

	m := <-recvchan
	require.Equal(t, m.Stream, "user1/device1/stream1")
	require.Equal(t, "Hello", m.Data[0].Data)

	_, err = msg2.Subscribe("user1/device2/>", recvchan)
	require.NoError(t, err)

	msg2.Flush()
	require.NoError(t, msg.Publish("user1/device2/stream2", Message{"user1/device2/stream2", "", []datastream.Datapoint{datastream.Datapoint{Data: "Hi"}}}))

	m = <-recvchan
	require.Equal(t, m.Stream, "user1/device2/stream2")
	require.Equal(t, "Hi", m.Data[0].Data)

}
