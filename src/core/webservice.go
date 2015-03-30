package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	//"plugins/web_client"
	"streamdb"
)

var (
	serverport = flag.Int("port", 8080, "The port number for the server to listen on.")
	helpflag   = flag.Bool("help", false, "Prints this help message")

	sqlserver   = flag.String("sql", "webservice.sqlite3", "")
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

	var err error
	db, err := streamdb.Open(*sqlserver, *redisserver, *msgserver)

	if err != nil {
		log.Print("Cannot open StreamDB")
		panic(err.Error())
	}
	defer db.Close()
	log.Fatal("This program is disabled until v0.2 StreamDB API is complete.")
	/*
		log.Printf("Starting Server on port %d", *serverport)

		r := mux.NewRouter()
		web_client.GetSubrouter(userdb, r)
		web_client.Setup(r, userdb)
		//r.HandleFunc("/", HomeHandler)
		http.Handle("/", r)

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *serverport), nil))
	*/
}
