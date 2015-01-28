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
    keys = flag.String("keys", ">", "The keys to write to database")
    helpflag = flag.Bool("help", false, "Prints this help message")
)

func main() {
    flag.Parse()


    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

    log.Fatal(timebatchdb.MessageWriter(*server,*keys))


}
