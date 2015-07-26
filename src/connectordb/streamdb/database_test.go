package streamdb

import (
	"testing"

	"connectordb/config"

	"github.com/stretchr/testify/require"
)

func TestDataBaseOperatorInterfaceBasics(t *testing.T) {

	db, err := Open(config.DefaultOptions)
	require.NoError(t, err)
	db.Clear()

	defer db.Close()
	go db.RunWriter()

	_, err = db.User()
	require.Equal(t, err, ErrAdmin)

	_, err = db.Device()
	require.Equal(t, err, ErrAdmin)

	require.Equal(t, AdminName, db.Name())

	i, err := db.CountAllUsers()
	require.NoError(t, err)
	require.EqualValues(t, 0, i)

	i, err = db.CountAllDevices()
	require.NoError(t, err)
	require.EqualValues(t, 0, i)

	i, err = db.CountAllStreams()
	require.NoError(t, err)
	require.EqualValues(t, 0, i)

}
