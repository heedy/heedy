/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator_test

import (
	"config"
	"connectordb"
	"log"
)

var db *connectordb.Database

func init() {
	tdb, err := connectordb.Open(config.TestConfiguration.Options())
	if err != nil {
		log.Fatal(err)
	}
	db = tdb
	go db.RunWriter()
}
