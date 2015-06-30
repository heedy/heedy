package streamdb

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

//Testing timebatchdb really messes with everything, so recreate the necessary stuff here
func ResetTimeBatch() error {
	sdb, err := sql.Open("postgres", "postgres://127.0.0.1:52592/connectordb?sslmode=disable")
	if err != nil {
		return err
	}
	sdb.Exec("DELETE FROM Users;")
	sdb.Exec("DELETE FROM Devices;")
	sdb.Close()

	//CLear timebatch
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	if err != nil {
		return err
	}
	return db.tdb.Clear()
}

func TestDataBaseOperatorInterfaceBasics(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()
	go db.RunWriter()

	_, err = db.User()
	require.Equal(t, err, ErrAdmin)

	_, err = db.Device()
	require.Equal(t, err, ErrAdmin)

	require.Equal(t, AdminName, db.Name())

}
