package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"plugins/rest"
	"streamdb"

	log "github.com/Sirupsen/logrus"
)

var (
	serverport = flag.Int("port", 8000, "The port number to listen on")
	helpflag   = flag.Bool("help", false, "Prints this message")

	sqlserver   = flag.String("sql", "postgres://127.0.0.1:52592/connectordb?sslmode=disable", "")
	redisserver = flag.String("redis", "localhost:6379", "The address to the redis instance")
	msgserver   = flag.String("msg", "localhost:4222", "The address of the messenger server")

	runwriter = flag.Bool("dbwriter", true, "Run the Database Writer (needed if dbwriter service off)")
	loglevel  = flag.String("log", "INFO", "The log level to run at")
)

func main() {
	flag.Parse()

	switch *loglevel {
	default:
		log.Panic("Unrecognized log level ", *loglevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	}

	if *helpflag {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	db, err := streamdb.Open(*sqlserver, *redisserver, *msgserver)

	if err != nil {
		log.Panic("Cannot open StreamDB", err)
	}
	defer db.Close()

	if *runwriter {
		go db.RunWriter()
	}

	r := rest.Router(db, nil)
	http.Handle("/", r)

	fmt.Println("Running REST API on port", *serverport)

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", *serverport), nil))

}
