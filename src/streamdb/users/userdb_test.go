/** Package users provides an API for managing user information.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

package users


import (
    //"os/exec"
    //"github.com/nu7hatch/gouuid"
    //"os"
    //"net"
    "log"
    "testing"
    //"time"
    )

var (
    postgres_folder string
    portnum = "52592"
)

/*BUG(daniel): hard-coding the postgres version and location causes explosion on my computer (my pgres is 9.4).
I also switched the port to the one used by the postgres run script

func start_psql() {
    log.Printf("Starting postgres, port: %v, dir: %v", portnum, postgres_folder)
    cmd := exec.Command("/usr/lib/postgresql/9.3/bin/postgres", "-D", postgres_folder, "-p", portnum)
    err := cmd.Run()

    if err != nil {
        log.Printf("Postgres Crashed %v", err)
    }
    log.Printf("Postgres Quit")

}


func init() {
    var err error

    //ApiKey, _ := uuid.NewV4()
    postgres_folder = "TESTING_postgresdb" //+ ApiKey.String()

    _ = os.RemoveAll(postgres_folder)

    ModePerm := os.FileMode(0700)
    os.Mkdir(postgres_folder, ModePerm)

    err = exec.Command("/usr/lib/postgresql/9.3/bin/initdb", "-D", postgres_folder).Run()

    if err != nil {
    	log.Fatal(err)
    }

    file, err := os.OpenFile(postgres_folder + "/postgresql.conf", os.O_RDWR|os.O_APPEND, os.ModeAppend) // For read access.
    if err != nil {
    	log.Fatal(err)
    }
    defer file.Close()

    file.WriteString("port = " + portnum + "\n")
    file.WriteString("listen_addresses = 'localhost'\n")
    file.WriteString("unix_socket_directories = '/tmp'\n\n")


    go start_psql()


    var conn net.Conn
    for i := 0; i < 30; i += 1 {
        log.Printf("Trying connection...")
        time.Sleep(time.Duration(1) * time.Second)
        conn, err = net.DialTimeout("tcp", "localhost:" + portnum, time.Duration(5) * time.Second)

        if conn != nil {
            break
        }
    }
    if err != nil {
        log.Fatal(err)
    }

    if conn != nil {
        conn.Close()
    }

    err = exec.Command("psql", "-h", "localhost", "-p", portnum, "-d", "postgres", "-c", "CREATE DATABASE connectordb;").Run()
    if err != nil {
        log.Printf("Could not create database %v", err)
    }
}
*/


func TestPostgresInit(t *testing.T) {

    log.Printf("Testing postgres init")
    db, err := NewPostgresUserDatabase("sslmode=disable dbname=connectordb port=" + portnum)


    if db == nil {
        t.Errorf("DB was returned nil")
    }


    if err != nil  && db != nil {
        t.Errorf("Err was not nil %v", err)
    }
}
