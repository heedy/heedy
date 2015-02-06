package timebatchdb

import (
    "log"
    )

func DatabaseWriter(msgurl string,mgourl string, router string,mdb string) error {

    log.Printf("Database Writer (MSG:%s,MGO:%s,K:%s,DB:%s)\n",msgurl,mgourl,router,mdb)

    m,err := OpenMongoStore(mgourl,mdb)
    if err!=nil {
        return err
    }
    defer m.Close()

    msgr,err := ConnectMessenger(msgurl)
    if (err != nil) {
        return err
    }
    defer msgr.Close()

    recvchan := make(chan KeyedDatapoint)
    _,err = msgr.SubChannel(router,recvchan)
    if err != nil {
        return err
    }
    log.Printf("Ready.\n")
    for msg := range recvchan {
        m.Append(msg.Key(),NewDatapointArray([]Datapoint{msg.Datapoint()}))
        log.Printf("%s\n",msg)

    }

    return nil
}
