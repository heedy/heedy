package timebatchdb

import (
    "testing"
    )

func TestMessenger(t *testing.T) {
    msg,err := ConnectMessenger("localhost:4222")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer msg.Close()

    msg2,err := ConnectMessenger("localhost:4222")
    if err!=nil {
        t.Errorf("Couldn't connect: %s",err)
        return
    }
    defer msg2.Close()

    recvchan := make(chan *Message)
    _,err = msg2.SubChannel("user1/>",recvchan)
    if err != nil {
        t.Errorf("Couldn't bind channel: %s",err)
        return
    }

    //Now, publish a message
    err = msg.Publish("user1/item1/stream1",1000,[]byte("Hello World!"))
    if (err != nil) {
        t.Errorf("Couldn't publish: %s",err)
        return
    }

    m := <- recvchan

    if (m.Timestamp!=1000 || string(m.Data)!="Hello World!" || m.Key!="user1/item1/stream1") {
        t.Errorf("Incorrect read %s",m)
        return
    }


}
