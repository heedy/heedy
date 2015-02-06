package main

import (
    "streamdb/timebatchdb"
    "fmt"
    "log"
    "flag"
    "os"
    )

var (
    msgserver = flag.String("msg", "localhost:4222", "The address of the messenger server")
    mgoserver = flag.String("mgo", "localhost", "The address of the MongoDB server")
    mgodb = flag.String("mgodb", "production_timebatchdb", "The name of the MongoDB database")
    routes = flag.String("route", ">", "The routes to write to database")
    helpflag = flag.Bool("help", false, "Prints this help message")
)

func main() {
    flag.Parse()


    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

    log.Fatal(timebatchdb.DatabaseWriter(*msgserver,*mgodb,*mgodb, *routes))


}
