package streamdb

import (
	"connectordb/streamdb/operator"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

func BenchmarkDeviceLogin(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	if err != nil {
		b.Errorf("Couldn't open database: %v", err)
		return
	}
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	dev, _ := db.ReadDevice("streamdb_test/user")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = db.LoginOperator("streamdb_test/user", dev.ApiKey)
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

	EnableCaching = true
}

func BenchmarkCreateUser(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	db.SetAdmin("streamdb_test", true)

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.CreateUser(name, name+"@localhost", "mypass"))
	}

}

func BenchmarkDeleteUser(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	db.SetAdmin("streamdb_test", true)
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, db.CreateUser(name, name+"@localhost", "mypass"))
	}

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.DeleteUser(name))
	}

}

func BenchmarkReadUserNC(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	db.SetAdmin("streamdb_test", true)
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, db.CreateUser(name, name+"@localhost", "mypass"))
	}

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		_, err := o.ReadUser(name)
		require.NoError(b, err)
	}

}

func BenchmarkReadUser(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := o.ReadUser("streamdb_test")
		require.NoError(b, err)
	}

}

func BenchmarkUpdateUser(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		u, err := o.ReadUser("streamdb_test")
		require.NoError(b, err)
		u.Email = strconv.FormatInt(int64(n), 32) + "@localhost"
		require.NoError(b, o.UpdateUser(u))
	}

}

func BenchmarkCreateStream(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sname := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.CreateStream("streamdb_test/user/"+sname, `{"type": "boolean"}`))
	}
}

func BenchmarkReadStreamNC(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	for n := 0; n < b.N; n++ {
		sname := strconv.FormatInt(int64(n), 32)
		require.NoError(b, db.CreateStream("streamdb_test/user/"+sname, `{"type": "boolean"}`))
	}

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sname := strconv.FormatInt(int64(n), 32)
		_, err := o.ReadStream("streamdb_test/user/" + sname)
		require.NoError(b, err)
	}
}

func BenchmarkReadStream(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := o.ReadStream("streamdb_test/user/mystream")
		require.NoError(b, err)
	}
}

func BenchmarkInsert1(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := []operator.Datapoint{operator.Datapoint{
			Timestamp: float64(n + 1),
			Data:      true,
		}}
		err = o.InsertStream("streamdb_test/user/mystream", data)
		require.NoError(b, err)
	}
}

func BenchmarkStreamLength(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	//go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	data := make([]operator.Datapoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = operator.Datapoint{
			Timestamp: float64(i),
			Data:      true,
		}
	}
	err = o.InsertStream("streamdb_test/user/mystream", data)
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = o.LengthStream("streamdb_test/user/mystream")
		require.NoError(b, err)
	}
}

func BenchmarkInsert1000(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := make([]operator.Datapoint, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = operator.Datapoint{
				Timestamp: float64(1000*n + i),
				Data:      true,
			}
		}
		err = o.InsertStream("streamdb_test/user/mystream", data)
		require.NoError(b, err)
	}
}

func BenchmarkRead1000(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	data := make([]operator.Datapoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = operator.Datapoint{
			Timestamp: float64(i),
			Data:      true,
		}
	}
	err = o.InsertStream("streamdb_test/user/mystream", data)
	require.NoError(b, err)
	time.Sleep(1 * time.Second) //Wait a moment for batch to have some time to write the data

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dr, err := db.GetStreamIndexRange("streamdb_test/user/mystream", 0, 0)
		require.NoError(b, err)
		v, err := dr.Next()
		require.NoError(b, err)
		ctr := 1
		for v != nil {
			v, err = dr.Next()
			require.NoError(b, err)
			ctr++
		}
		require.Equal(b, 1001, ctr)
	}
}

func BenchmarkReadLast10(b *testing.B) {
	ResetTimeBatch()
	db, err := Open("postgres://127.0.0.1:52592/connectordb?sslmode=disable", "localhost:6379", "localhost:4222")
	require.NoError(b, err)
	defer db.Close()
	go db.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	o, err := db.LoginOperator("streamdb_test", "mypass")
	require.NoError(b, err)

	data := make([]operator.Datapoint, 950)
	for i := 0; i < 950; i++ {
		data[i] = operator.Datapoint{
			Timestamp: float64(i),
			Data:      true,
		}
	}
	err = o.InsertStream("streamdb_test/user/mystream", data)
	require.NoError(b, err)
	time.Sleep(1 * time.Second) //Wait a moment for batch to have some time to write the data

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dr, err := db.GetStreamIndexRange("streamdb_test/user/mystream", -10, 0)
		require.NoError(b, err)
		v, err := dr.Next()
		require.NoError(b, err)
		ctr := 1
		for v != nil {
			v, err = dr.Next()
			require.NoError(b, err)
			ctr++
		}
		require.Equal(b, 11, ctr)
	}
}
