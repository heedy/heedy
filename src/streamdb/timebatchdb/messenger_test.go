package timebatchdb

import (
    "testing"
    "time"
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

    recvchan := make(chan KeyedDatapoint)
    _,err = msg2.SubChannel("user1/>",recvchan)
    if err != nil {
        t.Errorf("Couldn't bind channel: %s",err)
        return
    }
    //The connection needs to be flushed so that we are definitely subscribed to the channel
    //before we publish on it
    msg2.Flush()

    //Now, publish a message
    err = msg.Publish(NewKeyedDatapoint("user1/item1/stream1",1000,[]byte("Hello World!")),"")
    if (err != nil) {
        t.Errorf("Couldn't publish: %s",err)
        return
    }

    go func() {
        time.Sleep(1*time.Second)
        recvchan <- NewKeyedDatapoint("TIMEOUT",0,nil)
    }()

    m := <- recvchan
    if (m.Key()=="TIMEOUT") {
        t.Errorf("Message read timed out!")
        return
    }

    if (m.Timestamp()!=1000 || string(m.Data())!="Hello World!" || m.Key()!="user1/item1/stream1") {
        t.Errorf("Incorrect read %s",m)
        return
    }
}
