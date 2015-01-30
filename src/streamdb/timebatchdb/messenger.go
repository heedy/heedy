package timebatchdb

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
    m.econn.Flush()
    m.conn.Flush()
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

func (m *Messenger) Publish(d *KeyedDatapoint) error {
    return m.econn.Publish(strings.Replace(d.Key(),"/",".",-1),d)
}

func (m *Messenger) Subscribe(key string, fn SubscriptionFunction) (*nats.Subscription,error){
    return m.econn.Subscribe(strings.Replace(key,"/",".",-1),fn)
}

func (m *Messenger) SubChannel(key string, chn chan *KeyedDatapoint) (*nats.Subscription,error) {
    return m.econn.BindRecvChan(strings.Replace(key,"/",".",-1),chn)
}
