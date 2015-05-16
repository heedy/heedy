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
		recvchan <- Message{"TIMEOUT", "", "", []Datapoint{}}
	}()

	_, err = msg2.SubscribeStream("*", "user1/device1/stream1", recvchan)
	require.NoError(t, err)

	//The connection needs to be flushed so that we are definitely subscribed to the channel
	//before we publish on it
	msg2.Flush()

	//Now, publish a message
	err = msg.Publish(Message{"user1/device1/stream1", "user1/user", "d", []Datapoint{Datapoint{Data: "Hello"}}})
	require.NoError(t, err)

	m := <-recvchan
	require.NotEqual(t, m.To, "TIMEOUT")
	require.Equal(t, m.To, "user1/device1/stream1")
	require.Equal(t, "user1/user", m.From)
	require.Equal(t, "Hello", m.Data[0].Data)
	require.Equal(t, "d", m.Prefix)

	require.Equal(t, "[To=user1/device1/stream1 From=user1/user Pre=d]", m.String())

	_, err = msg2.SubscribeSenderDevice("*", "user1/device2", recvchan)
	require.NoError(t, err)

	msg2.Flush()
	require.NoError(t, msg.Publish(Message{"user1/device1/stream2", "user1/device2", "d", []Datapoint{Datapoint{Data: "Hi"}}}))

	m = <-recvchan
	require.NotEqual(t, m.To, "TIMEOUT")
	require.Equal(t, m.To, "user1/device1/stream2")
	require.Equal(t, "user1/device2", m.From)
	require.Equal(t, "Hi", m.Data[0].Data)
	require.Equal(t, "d", m.Prefix)

	_, err = msg2.SubscribeReceiverDevice("*", "user1/device6", recvchan)
	require.NoError(t, err)

	msg2.Flush()
	require.NoError(t, msg.Publish(Message{"user1/device6/stream2", "user1/device9", "d", []Datapoint{Datapoint{Data: "z"}}}))

	m = <-recvchan
	require.NotEqual(t, m.To, "TIMEOUT")
	require.Equal(t, m.To, "user1/device6/stream2")
	require.Equal(t, "user1/device9", m.From)
	require.Equal(t, "z", m.Data[0].Data)
	require.Equal(t, "d", m.Prefix)

	_, err = msg2.SubscribePrefix("t", recvchan)
	require.NoError(t, err)

	msg2.Flush()
	require.NoError(t, msg.Publish(Message{"uper1/desfce1/stream2", "gher1/device2", "t", []Datapoint{Datapoint{Data: "per"}}}))

	m = <-recvchan
	require.NotEqual(t, m.To, "TIMEOUT")
	require.Equal(t, m.To, "uper1/desfce1/stream2")
	require.Equal(t, "gher1/device2", m.From)
	require.Equal(t, "per", m.Data[0].Data)
	require.Equal(t, "t", m.Prefix)

}
