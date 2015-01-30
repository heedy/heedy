package timebatchdb

import (
    "log"
    )

func MessageWriter(url string,key string) error {

    log.Printf("Database Writer (%s,%s)",url,key)

    msg,err := ConnectMessenger(url)
    if (err != nil) {
        return err
    }
    defer msg.Close()

    recvchan := make(chan *KeyedDatapoint)
    _,err = msg.SubChannel(key,recvchan)
    if err != nil {
        return err
    }

    for m := range recvchan {
        log.Printf("%s\n",m)
    }


    return nil
}
