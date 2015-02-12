package datastore

import (
    "github.com/apcera/nats"
    "strings"
    )

type Messenger struct {
    conn *nats.Conn             //The NATS connection
    econn *nats.EncodedConn     //The Encoded conn, ie, a data message
}

type SubscriptionFunction func(*KeyedDatapoint)


func (m *Messenger) Close() {
    m.econn.Close()
    m.conn.Close()
}

func ConnectMessenger(url string) (*Messenger,error){
    conn, err := nats.Connect("nats://"+url)
    if err!=nil {
        return nil,err
    }
    econn, err := nats.NewEncodedConn(conn,"gob")
    if err!=nil {
        conn.Close()
        return nil,err
    }

    return &Messenger{conn,econn},nil
}

func (m *Messenger) Publish(d KeyedDatapoint,routing string) error {
    if (len(routing)==0) {
        routing = d.Key()
    }
    routing=strings.Replace(routing,"/",".",-1)
    return m.econn.Publish(routing,d)
}

func (m *Messenger) Subscribe(router string, fn SubscriptionFunction) (*nats.Subscription,error){
    return m.econn.Subscribe(strings.Replace(router,"/",".",-1),fn)
}

func (m *Messenger) SubChannel(router string, chn chan KeyedDatapoint) (*nats.Subscription,error) {
    return m.econn.BindRecvChan(strings.Replace(router,"/",".",-1),chn)
}

//Makes sure all commands are acknowledged by the server
func (m *Messenger) Flush() {
    m.econn.Flush()
}
