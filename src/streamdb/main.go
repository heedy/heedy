package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "flag"
    "fmt"
    "streamdb/users"
    "os"
    )

var (
    serverport = flag.Int("port", 8080, "The port number for the server to listen on.")
    helpflag = flag.Bool("help", false, "Prints this help message")
)


func main() {
    flag.Parse()


    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

    log.Printf("Starting Server")
    
    r := mux.NewRouter()
    users.GetSubrouter(r)
    //r.HandleFunc("/", HomeHandler)
    http.Handle("/", r)

    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverport), nil))
}
