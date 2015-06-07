package main

import (
	"connectordb/plugins/rest"
	"connectordb/streamdb"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	serverport = flag.Int("port", 8000, "The port number to listen on")
	helpflag   = flag.Bool("help", false, "Prints this message")

	sqlserver   = flag.String("sql", "postgres://127.0.0.1:52592/connectordb?sslmode=disable", "")
	redisserver = flag.String("redis", "localhost:6379", "The address to the redis instance")
	msgserver   = flag.String("msg", "localhost:4222", "The address of the messenger server")

	runwriter = flag.Bool("dbwriter", true, "Run the Database Writer (needed if dbwriter service off)")
	loglevel  = flag.String("log", "INFO", "The log level to run at")
	logfile   = flag.String("logfile", "", "The log file to write to")
)

func main() {
	flag.Parse()

	//Set up the log file
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Could not open file %s: %s", *logfile, err.Error())
		}
		defer f.Close()
		log.SetFormatter(new(log.JSONFormatter))
		log.SetOutput(f)
	}

	switch *loglevel {
	default:
		log.Fatalln("Unrecognized log level ", *loglevel)
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
		log.Fatalln("Cannot open StreamDB: ", err)
	}
	defer db.Close()

	if *runwriter {
		go db.RunWriter()
	}
	r := mux.NewRouter()
	rest.Router(db, r.PathPrefix("/api/v1").Subrouter())
	http.Handle("/", r)

	fmt.Println("Running REST API on port", *serverport)

	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", *serverport), nil))

}
