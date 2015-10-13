package connectordb

import (
	"config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDataBaseOperatorInterfaceBasics(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)

	defer db.Close()
	go db.RunWriter()

	require.NotEqual(t, db.GetUserDatabase(), nil)
	require.Equal(t, db.GetDatastream(), db.ds)
	require.Equal(t, db.GetMessenger(), db.msg)
}
