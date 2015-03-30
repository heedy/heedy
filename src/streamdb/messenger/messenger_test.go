package messenger

import (
	"errors"
	"testing"
	"time"
)

func TestMessenger(t *testing.T) {
	newerr := errors.New("FAIL")
	_, err := Connect("localhost:4222", newerr)
	if err != newerr {
		t.Errorf("Error chain failed: %s", err)
		return
	}

	msg, err := Connect("localhost:4222", nil)
	if err != nil {
		t.Errorf("Couldn't connect: %s", err)
		return
	}
	defer msg.Close()

	msg2, err := Connect("localhost:4222", nil)
	if err != nil {
		t.Errorf("Couldn't connect: %s", err)
		return
	}
	defer msg2.Close()

	recvchan := make(chan Message)

	//We bind a timeout to the channel, since we want the test to fail if no messages come through
	go func() {
		time.Sleep(2 * time.Second)
		recvchan <- Message{"", "", "", "", []byte("TIMEOUT")}
	}()

	_, err = msg2.SubscribeStream("*", "user1/device1/stream1", recvchan)
	if err != nil {
		t.Errorf("Couldn't bind stream: %s", err)
		return
	}
	//The connection needs to be flushed so that we are definitely subscribed to the channel
	//before we publish on it
	msg2.Flush()

	//Now, publish a message
	err = msg.Publish(Message{"user1/device1/stream1", "user1/user", "d", "", []byte("Hello")})
	if err != nil {
		t.Errorf("Couldn't publish: %s", err)
		return
	}

	m := <-recvchan
	if string(m.Data) == "TIMEOUT" {
		t.Errorf("Message read timed out!")
		return
	}
	if m.To != "user1/device1/stream1" || m.From != "user1/user" || string(m.Data) != "Hello" || m.Prefix != "d" || m.IsFromSelf() {
		t.Errorf("Incorrect read %s", m)
		return
	}
	if m.String() != "[To=user1/device1/stream1 From=user1/user Pre=d]" {
		t.Errorf("Incorrect string %s", m)
		return
	}

	_, err = msg2.SubscribeSenderDevice("*", "user1/device2", recvchan)
	if err != nil {
		t.Errorf("Couldn't bind device: %s", err)
		return
	}
	msg2.Flush()
	err = msg.Publish(Message{"user1/device1/stream2", "user1/device2", "d", "", []byte("Hello")})
	if err != nil {
		t.Errorf("Couldn't publish: %s", err)
		return
	}
	m = <-recvchan
	if string(m.Data) == "TIMEOUT" {
		t.Errorf("Message read timed out!")
		return
	}
	if m.To != "user1/device1/stream2" || m.From != "user1/device2" || string(m.Data) != "Hello" || m.Prefix != "d" {
		t.Errorf("Incorrect read %s", m)
		return
	}

	_, err = msg2.SubscribeReceiverDevice("*", "user1/device6", recvchan)
	if err != nil {
		t.Errorf("Couldn't bind device: %s", err)
		return
	}
	msg2.Flush()
	err = msg.Publish(Message{"user1/device6/stream2", "user1/device9", "d", "", []byte("Hello")})
	if err != nil {
		t.Errorf("Couldn't publish: %s", err)
		return
	}
	m = <-recvchan
	if string(m.Data) == "TIMEOUT" {
		t.Errorf("Message read timed out!")
		return
	}
	if m.To != "user1/device6/stream2" || m.From != "user1/device9" || string(m.Data) != "Hello" || m.Prefix != "d" {
		t.Errorf("Incorrect read %s", m)
		return
	}

	_, err = msg2.SubscribePrefix("t", recvchan)
	if err != nil {
		t.Errorf("Couldn't bind device: %s", err)
		return
	}
	msg2.Flush()
	err = msg.Publish(Message{"uper1/desfce1/stream2", "gher1/device2", "t", "", []byte("Hello")})
	if err != nil {
		t.Errorf("Couldn't publish: %s", err)
		return
	}
	m = <-recvchan
	if string(m.Data) == "TIMEOUT" {
		t.Errorf("Message read timed out!")
		return
	}
	if m.To != "uper1/desfce1/stream2" || m.From != "gher1/device2" || string(m.Data) != "Hello" || m.Prefix != "t" {
		t.Errorf("Incorrect read %s", m)
		return
	}
}
