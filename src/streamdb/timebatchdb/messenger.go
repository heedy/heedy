package timebatchdb

import (
    "github.com/apcera/nats"
    "strings"
    "time"
    )

type Messenger struct {
    conn *nats.Conn             //The NATS connection
    econn *nats.EncodedConn     //The Encoded conn, ie, a data message
}

type Message struct {
    Timestamp uint64            //The timestamp associated with data
    Key string                  //The key name associated with message
    Data []byte                 //The associated data
}

func (m *Message) String() string {
    return "[KEY="+m.Key+" TIME="+time.Unix(0,int64(m.Timestamp)).String()+" DATA="+string(m.Data)+"]"
}

type SubscriptionFunction func(*Message)


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

func (m *Messenger) Publish(key string,timestamp uint64,data []byte) error {
    return m.econn.Publish(strings.Replace(key,"/",".",-1),&Message{timestamp,key,data})
}

func (m *Messenger) Subscribe(key string, fn SubscriptionFunction) (*nats.Subscription,error){
    return m.econn.Subscribe(strings.Replace(key,"/",".",-1),fn)
}

func (m *Messenger) SubChannel(key string, chn chan *Message) (*nats.Subscription,error) {
    return m.econn.BindRecvChan(strings.Replace(key,"/",".",-1),chn)
}
