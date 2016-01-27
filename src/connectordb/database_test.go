/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package connectordb

import (
	"config"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

var Tdb *Database

func init() {
	db, err := Open(config.TestConfiguration.Options())
	if err != nil {
		log.Fatal(err)
	}
	Tdb = db
	go db.RunWriter()
}

func TestDataBaseBasics(t *testing.T) {
	var o PathOperator
	o = Tdb
	require.Equal(t, o.Name(), Name)
	_, err := o.User()
	require.Error(t, err)
	_, err = o.Device()
	require.Error(t, err)

}
