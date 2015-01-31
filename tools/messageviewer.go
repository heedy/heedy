package main

import (
    "streamdb/timebatchdb"
    "fmt"
    "log"
    "flag"
    "os"
    )

var (
    server = flag.String("msg", "localhost:4222", "The address of the messenger server")
    keys = flag.String("keys", ">", "The keys to watch")
    helpflag = flag.Bool("help", false, "Prints this help message")
)

func MessageView(url string,key string) error {

    log.Printf("MessageViewer (%s,%s)",url,key)

    msg,err := timebatchdb.ConnectMessenger(url)
    if (err != nil) {
        return err
    }
    defer msg.Close()

    recvchan := make(chan timebatchdb.KeyedDatapoint)
    _,err = msg.SubChannel(key,recvchan)
    if err != nil {
        return err
    }

    for m := range recvchan {
        fmt.Printf("%s\n",m)
    }

    return nil
}


func main() {
    flag.Parse()


    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

    log.Fatal(MessageView(*server,*keys))


}
