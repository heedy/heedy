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

func TestDataBaseBasics(t *testing.T) {
	var o Operator
	db, err := Open(config.TestConfiguration.Options())
	require.NoError(t, err)

	// This esnures that Database conforms to Operator
	o = db

	defer db.Close()
	go db.RunWriter()

	require.Equal(t, o.Name(), Name)
	_, err = o.User()
	require.Error(t, err)
	_, err = o.Device()
	require.Error(t, err)

}
