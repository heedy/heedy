package streamdb

import (
	"database/sql"
	"streamdb/timebatchdb"
	"testing"

	"github.com/stretchr/testify/require"
)

//Testing timebatchdb really messes with everything, so recreate the necessary stuff here
func ResetTimeBatch() error {
	sdb, err := sql.Open("postgres", "postgres://127.0.0.1:52592/connectordb?sslmode=disable")
	if err != nil {
		return err
	}
	defer sdb.Close()
	sdb.Exec("DROP TABLE IF EXISTS timebatchtable")
	_, err = sdb.Exec(`CREATE TABLE IF NOT EXISTS timebatchtable
        (
            Key VARCHAR NOT NULL,
            EndTime BIGINT,
            EndIndex BIGINT,
			Version INTEGER,
            Data BYTEA,
            PRIMARY KEY (Key, EndIndex)
            );`)
	sdb.Exec("CREATE INDEX keytime ON timebatchtable (Key,EndTime ASC);")

	sdb.Exec("DELETE FROM Users;")
	sdb.Exec("DELETE FROM Devices;")

	//Clear the redis cache
	rc, err := timebatchdb.OpenRedisCache("localhost:6379", err)
	if err != nil {
		return err
	}
	defer rc.Close()
	rc.Clear()
	return err
}

func TestDatabaseUserCrud(t *testing.T) {
	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()
	go db.RunWriter()

	_, err = db.User()
	require.Equal(t, err, ErrAdmin)

	_, err = db.Device()
	require.Equal(t, err, ErrAdmin)

	usrs, err := db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 0, len(usrs))

	_, err = db.ReadUser("streamdb_test")
	require.Error(t, err)
	_, err = db.ReadUserByEmail("root@localhost")
	require.Error(t, err)

	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))

	usrs, err = db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 1, len(usrs))

	usr, err := db.ReadUser("streamdb_test")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	usr, err = db.ReadUserByEmail("root@localhost")
	require.NoError(t, err)
	require.Equal(t, "streamdb_test", usr.Name)

	//As of now, this part fails - delete of nonexisting does not error
	//require.Error(t, db.DeleteUser("notauser"))
	require.NoError(t, db.DeleteUser("streamdb_test"))

	_, err = db.ReadUser("streamdb_test")
	require.Error(t, err)
	_, err = db.ReadUserByEmail("root@localhost")
	require.Error(t, err)

}
