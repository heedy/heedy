package main

import (
    "streamdb/timebatchdb"
    "database/sql"
	_ "github.com/lib/pq"
    "fmt"
    "log"
    "flag"
    "os"
    )

var (
    sql_server = flag.String("postgres","localhost:52592","location of postgres server")
    redis_server = flag.String("redis","localhost:6379","location of redis server")
    batch_size = flag.Int("batchsize",100,"The number of datapoints per batch of data")
    helpflag = flag.Bool("help", false, "Prints this help message")
)

func main() {
    flag.Parse()


    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

    //First create the postgres connection string
    postgres_conn := fmt.Sprintf("postgres://%s/connectordb?sslmode=disable",*sql_server)

    pdb,err := sql.Open("postgres",postgres_conn)
    if err!=nil {
        log.Fatal("Couldn't connect to postgres: ",err)
        return
    }
    defer pdb.Close()

    db,err := timebatchdb.Open(pdb,"postgres",*redis_server,*batch_size,nil)
    if err!=nil {
        log.Fatal("Couldn't open TimebatchDB: ",err)
        return
    }
    defer db.Close()


    log.Fatal(db.WriteDatabase())

}
