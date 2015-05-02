package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"plugins/rest"
	"streamdb"
)

var (
	serverport = flag.Int("port", 8000, "The port number to listen on")
	helpflag   = flag.Bool("help", false, "Prints this message")

	sqlserver   = flag.String("sql", "postgres://127.0.0.1:52592/connectordb?sslmode=disable", "")
	redisserver = flag.String("redis", "localhost:6379", "The address to the redis instance")
	msgserver   = flag.String("msg", "localhost:4222", "The address of the messenger server")
)

func main() {
	flag.Parse()

	if *helpflag {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	db, err := streamdb.Open(*sqlserver, *redisserver, *msgserver)

	if err != nil {
		log.Print("Cannot open StreamDB")
		panic(err.Error())
	}
	defer db.Close()

	log.Printf("Running REST API on port %d", *serverport)

	r := rest.Router(db, nil)
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverport), nil))

}
