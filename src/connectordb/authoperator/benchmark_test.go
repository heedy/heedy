/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package authoperator_test

import (
	"connectordb/datastream"
	"connectordb/users"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func BenchmarkDeviceLogin(b *testing.B) {
	db.Clear()

	//go db.RunWriter()
	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	dev, _ := db.ReadDevice("streamdb_test/user")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := db.DeviceLogin(dev.APIKey)
		if err != nil {
			b.Errorf("Login Failed: %v", err)
			return
		}
	}
}

func BenchmarkCreateUser(b *testing.B) {
	db.Clear()
	//go db.RunWriter()
	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})

	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.CreateUser(&users.UserMaker{User: users.User{Name: name, Email: name + "@localhost", Password: "mypass", Role: "user", Public: true}}))
	}

}

func BenchmarkDeleteUser(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, db.CreateUser(&users.UserMaker{User: users.User{Name: name, Email: name + "@localhost", Password: "mypass", Role: "user", Public: true}}))
	}

	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.DeleteUser(name))
	}

}

func BenchmarkReadUser(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})

	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := o.ReadUser("streamdb_test")
		require.NoError(b, err)
	}

}

func BenchmarkUpdateUser(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {

		require.NoError(b, o.UpdateUser("streamdb_test", map[string]interface{}{"email": strconv.FormatInt(int64(n), 32) + "@localhost"}))
	}

}

func BenchmarkCreateStream(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sname := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.CreateStream("streamdb_test/user/"+sname, &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))
	}
}

func BenchmarkReadStream(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := o.ReadStream("streamdb_test/user/mystream")
		require.NoError(b, err)
	}
}

func BenchmarkInsert1(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := []datastream.Datapoint{datastream.Datapoint{
			Timestamp: float64(n + 1),
			Data:      true,
		}}
		err = o.InsertStream("streamdb_test/user/mystream", data, false)
		require.NoError(b, err)
	}
}

func BenchmarkStreamLength(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))

	data := make([]datastream.Datapoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = datastream.Datapoint{
			Timestamp: float64(i + 1),
			Data:      true,
		}
	}
	err = o.InsertStream("streamdb_test/user/mystream", data, false)
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = o.LengthStream("streamdb_test/user/mystream")
		require.NoError(b, err)
	}
}

func BenchmarkInsert1000(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := make([]datastream.Datapoint, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = datastream.Datapoint{
				Timestamp: float64(1000*n + i + 1),
				Data:      true,
			}
		}
		err = o.InsertStream("streamdb_test/user/mystream", data, false)
		require.NoError(b, err)
	}
	b.StopTimer()
}

func BenchmarkRead1000(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))

	data := make([]datastream.Datapoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = datastream.Datapoint{
			Timestamp: float64(i + 1),
			Data:      true,
		}
	}
	err = o.InsertStream("streamdb_test/user/mystream", data, false)
	require.NoError(b, err)
	time.Sleep(1 * time.Second) //Wait a moment for batch to have some time to write the data

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		dr, err := db.GetStreamIndexRange("streamdb_test/user/mystream", 0, 0, "")
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
		dr.Close()
	}
	b.StopTimer()
}

func BenchmarkReadLast10(b *testing.B) {
	db.Clear()

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	o, err := db.UserLogin("streamdb_test", "mypass")
	require.NoError(b, err)

	db.CreateUser(&users.UserMaker{User: users.User{Name: "streamdb_test", Email: "root@localhost", Password: "mypass", Role: "admin", Public: true}})
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", &users.StreamMaker{Stream: users.Stream{Schema: `{"type": "boolean"}`}}))

	data := make([]datastream.Datapoint, 950)
	for i := 0; i < 950; i++ {
		data[i] = datastream.Datapoint{
			Timestamp: float64(i + 1),
			Data:      true,
		}
	}
	err = o.InsertStream("streamdb_test/user/mystream", data, false)
	require.NoError(b, err)
	time.Sleep(500 * time.Millisecond) //Wait a moment for batch to have some time to write the data
	b.ResetTimer()
	//t:=time.Now()
	for n := 0; n < b.N; n++ {
		//fmt.Println("Starting")

		//fmt.Println("T=", time.Since(t))
		dr, err := db.GetStreamIndexRange("streamdb_test/user/mystream", -10, 0, "")
		require.NoError(b, err)
		//fmt.Println("T=", time.Since(t))
		v, err := dr.Next()
		//fmt.Println("T=", time.Since(t))
		require.NoError(b, err)
		ctr := 1
		for v != nil {
			v, err = dr.Next()
			require.NoError(b, err)
			ctr++
		}

		require.Equal(b, 11, ctr)
		dr.Close()
		//fmt.Println("T=", time.Since(t))
	}
	b.StopTimer()

}
