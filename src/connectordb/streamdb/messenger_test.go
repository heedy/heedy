package streamdb

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMessenger(t *testing.T) {
	newerr := errors.New("FAIL")
	_, err := ConnectMessenger("localhost:4222", newerr)
	if err != newerr {
		t.Errorf("Error chain failed: %s", err)
		return
	}

	_, err = ConnectMessenger("localhost:13378", nil)
	require.Error(t, err)

	msg, err := ConnectMessenger("localhost:4222", nil)
	require.NoError(t, err)
	defer msg.Close()

	msg2, err := ConnectMessenger("localhost:4222", nil)
	require.NoError(t, err)
	defer msg2.Close()

	recvchan := make(chan Message)

	//We bind a timeout to the channel, since we want the test to fail if no messages come through
	go func() {
		time.Sleep(2 * time.Second)
		recvchan <- Message{"TIMEOUT", []Datapoint{}}
	}()

	_, err = msg2.Subscribe("user1/device1/stream1", recvchan)
	require.NoError(t, err)

	//The connection needs to be flushed so that we are definitely subscribed to the channel
	//before we publish on it
	msg2.Flush()

	//Now, publish a message
	err = msg.Publish("user1/device1/stream1/", Message{"user1/device1/stream1", []Datapoint{Datapoint{Data: "Hello"}}})
	require.NoError(t, err)

	m := <-recvchan
	require.Equal(t, m.Stream, "user1/device1/stream1")
	require.Equal(t, "Hello", m.Data[0].Data)

	require.Equal(t, "[S=user1/device1/stream1]", m.String())

	_, err = msg2.Subscribe("user1/device2/>", recvchan)
	require.NoError(t, err)

	msg2.Flush()
	require.NoError(t, msg.Publish("user1/device2/stream2", Message{"user1/device2/stream2", []Datapoint{Datapoint{Data: "Hi"}}}))

	m = <-recvchan
	require.Equal(t, m.Stream, "user1/device2/stream2")
	require.Equal(t, "Hi", m.Data[0].Data)

}
