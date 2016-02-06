/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package connectordb

import (
	"config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDataBaseOperatorInterfaceBasics(t *testing.T) {

	db, err := Open(config.TestConfiguration.Options())
	require.NoError(t, err)

	defer db.Close()
	go db.RunWriter()

	require.NotEqual(t, db.GetUserDatabase(), nil)
	require.Equal(t, db.GetDatastream(), db.ds)
	require.Equal(t, db.GetMessenger(), db.msg)
}
