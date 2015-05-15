package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"streamdb"
)

var (
	helpflag = flag.Bool("help", false, "Prints this message")

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

	//First create the postgres connection string
	postgresConn := fmt.Sprintf("postgres://%s/connectordb?sslmode=disable", *sqlserver)

	db, err := streamdb.Open(postgresConn, *redisserver, *msgserver)

	if err != nil {
		log.Print("Cannot open StreamDB")
		panic(err.Error())
	}
	defer db.Close()

	db.RunWriter()

}
