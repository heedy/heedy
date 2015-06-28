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

func TestDatabaseOperatorBasics(t *testing.T) {
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

func TestCacheCuriosities(t *testing.T) {
	//Adding the cache has also added several curious things that can go wrong
	//We make sure that they don't happen here

	//One more thing: Adding the cache adds a lot of question marks in multi-node setups, because
	//there WILL be invalid nodes added at some point. Therefore, it will be important to have a periodic
	//userdb cleanup process that deletes things that are not linked to valid ids

	require.NoError(t, ResetTimeBatch())

	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(t, err)
	defer db.Close()
	//go db.RunWriter()

	usrs, err := db.ReadAllUsers()
	require.NoError(t, err)
	require.Equal(t, 0, len(usrs))

	//Create the user
	require.NoError(t, db.CreateUser("streamdb_test", "root@localhost", "mypass"))
	//Change the user's name
	u, err := db.ReadUser("streamdb_test")
	require.NoError(t, err)
	u.Name = "tstr"
	db.UpdateUser(u)

	//Now let's see if the cache removed the old user name
	u, err = db.ReadUser("streamdb_test")
	require.Error(t, err, "Changing username leaves old name in cache")

	//Now, let's make sure that after a user is deleted, its devices/streams are deleted also
	//This leads to a REALLY weird bug in python tests, which create and delete the same username
	//many times, and the user device gets mismatched between them if the cache is not purged
	//of deleted user's devices/streams
	_, err = db.ReadUser("tstr")
	require.NoError(t, err)

	require.NoError(t, db.CreateDevice("tstr/mydevice"))

	_, err = db.ReadDevice("tstr/mydevice")
	require.NoError(t, err)

	require.NoError(t, db.DeleteUser("tstr"))

	_, err = db.ReadDevice("tstr/mydevice")
	require.Error(t, err, "Deleting a user does not propagate to cached devices")
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
