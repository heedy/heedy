package main

/**

This file provides us a way to initialize and update databases to the most
current standards.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"streamdb/dbutil"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage %s connectionstring [drop_old_tables]\n", os.Args[0])

        fmt.Fprintf(os.Stderr, `

Use this program to upgrade or initialize a new datbase with the given
connection string.

If you wish to drop the old tables upon upgrade, specify that by appending
"drop_old_tables" after the connection string. Otherwise, the original tables
from before the upgrade will be stored in backup locations within the database
for safety reasons.

If you decide to drop_old_tables, please do a full database backup beforehand
to ensure integrity in case something fails.

Example Usage:

    initdb /etc/nonexistant.sqlite3 // creates a new sqlite3 db
    initdb /etc/existing.sqlite3    // updates an existing sqlite3

    // Upgrades the database and deletes the backup tables from before the
    // upgrade
    initdb /etc/existing.sqlite3 drop_old_tables

NOTE for SQLITE3 databases:

    You must have sqlite3 on your $PATH somewhere if you want to upgrade a
    sqlite3 database due to driver issues.

    You must have write permissions in the current directory so this command
    can generate a temporary upgrade file to upgrade the database with.
`)
        return
	}

    cxnstring := os.Args[1]
    dropold := false

    if len(os.Args) >= 3 && os.Args[2] == "drop_old_tables" {
        dropold = true
    }


    err := dbutil.UpgradeDatabase(cxnstring, dropold)
    if err != nil {
        log.Fatal("Fatal Error: ", err.Error())
    }

    log.Printf("Success")
}
