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

//Let's see if the cache actually helps much with login speed
func BenchmarkUserLogin(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = db.LoginOperator("streamdb_test", "mypass")
		if err != nil {
			b.Errorf("Login Failed: %v", err)
			return
		}
	}
}

func BenchmarkUserLoginNoCache(b *testing.B) {
	ResetTimeBatch()
	EnableCaching = false
	//CacheExpireTime = 0 //Cache expires IMMEDIATELY
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = db.LoginOperator("streamdb_test", "mypass")
			if err != nil {
				b.Errorf("Login Failed: %v", err)
				return
			}
		}
	})
}
