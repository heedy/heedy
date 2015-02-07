package timebatchdb

import (
    "log"
    )

func DatabaseWriter(msgurl string,mgourl string,mdb string, router string) error {

    log.Printf("DBWriter (MSG:%s,MGO:%s,K:%s,DB:%s)\n",msgurl,mgourl,router,mdb)

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
    log.Printf("DBWriter Ready.\n")
    for msg := range recvchan {
        log.Printf("%s\n",msg)
        err = m.Append(msg.Key(),NewDatapointArray([]Datapoint{msg.Datapoint()}))
        if err != nil {
            log.Printf("DBWriter ERROR: %s\n",err)
        }
    }

    return nil
}
