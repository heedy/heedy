package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "flag"
    "fmt"
    "streamdb/users"
    "os"
    "plugins/web_client"
    "streamdb/timebatchdb"
    "streamdb/dtypes"
    )

var (
    serverport = flag.Int("port", 8080, "The port number for the server to listen on.")
    helpflag = flag.Bool("help", false, "Prints this help message")

    msgserver        = flag.String("msg", "localhost:4222", "The address of the messenger server")
    mgoserver        = flag.String("mgo", "localhost", "The address of the MongoDB server")
    mgodb            = flag.String("mgodb", "production_timebatchdb", "The name of the MongoDB database")
    routes           = flag.String("route", ">", "The routes to write to database")

)


func main() {
    flag.Parse()


    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

    var err error
    userdb, err := users.NewSqliteUserDatabase("production.sqlite")

    if err != nil {
        userdb = nil
        log.Print("Cannot open user database")
    }

    go timebatchdb.DatabaseWriter(*msgserver, *mgoserver, *mgodb, *routes)
    timedb, err := dtypes.Open(*msgserver,*mgoserver,*mgodb)

    if err != nil {
        timedb = nil
        log.Print("Cannot open timeseries database")

    }

    log.Printf("Starting Server on port %d", *serverport)

    r := mux.NewRouter()
    web_client.GetSubrouter(userdb, timedb, r)
    web_client.Setup(r, userdb)
    //r.HandleFunc("/", HomeHandler)
    http.Handle("/", r)

    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverport), nil))
}
