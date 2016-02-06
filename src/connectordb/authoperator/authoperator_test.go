/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator_test

import (
	"config"
	"connectordb"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestAuthOperatorBasics(t *testing.T) {
	db.Clear()
	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass", "user", true))

	_, err := db.UserLogin("streamdb_test", "wrongpass")
	require.Error(t, err)

	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(t, err)

	require.Equal(t, "streamdb_test/user", o.Name())

	u, err := o.User()
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", u.Name)

	d, err := o.Device()
	require.NoError(t, err)
	require.Equal(t, "user", d.Name)

	apiOp, err := db.DeviceLogin(d.APIKey)
	require.NoError(t, err)
	require.Equal(t, "streamdb_test/user", apiOp.Name())

}

func TestCountAll(t *testing.T) {
	db.Clear()

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass", "user", true))

	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(t, err)

	_, err = o.CountUsers()
	require.Error(t, err)
	_, err = o.CountDevices()
	require.Error(t, err)
	_, err = o.CountStreams()
	require.Error(t, err)

	db.UpdateUser("streamdb_test", map[string]interface{}{"role": "admin"})

	i, err := o.CountUsers()
	require.NoError(t, err)
	require.EqualValues(t, i, 1)
	i, err = o.CountDevices()
	require.NoError(t, err)
	require.True(t, i >= 1)
	i, err = o.CountStreams()
	require.NoError(t, err)
	require.True(t, i >= 1)
}
