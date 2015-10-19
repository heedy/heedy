package authoperator

import (
	"connectordb/datastream"
	"connectordb/operator/interfaces"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func BenchmarkDeviceLogin(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	dev, _ := db.ReadDevice("streamdb_test/user")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = NewAPILoginOperator(db, dev.ApiKey)
		if err != nil {
			b.Errorf("Login Failed: %v", err)
			return
		}
	}
}

func BenchmarkCreateUser(b *testing.B) {

	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()
	//go db.RunWriter()
	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	db.SetAdmin("streamdb_test", true)

	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, o.CreateUser(name, name+"@localhost", "mypass"))
	}

}

func BenchmarkDeleteUser(b *testing.B) {

	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	db.SetAdmin("streamdb_test", true)
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, db.CreateUser(name, name+"@localhost", "mypass"))
	}

	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		name := strconv.FormatInt(int64(n), 32)
		require.NoError(b, pathOperator.DeleteUser(name))
	}

}

func BenchmarkReadUser(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")

	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := o.ReadUser("streamdb_test")
		require.NoError(b, err)
	}

}

func BenchmarkUpdateUser(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
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
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sname := strconv.FormatInt(int64(n), 32)
		require.NoError(b, pathOperator.CreateStream("streamdb_test/user/"+sname, `{"type": "boolean"}`))
	}
}

func BenchmarkReadStream(b *testing.B) {

	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := pathOperator.ReadStream("streamdb_test/user/mystream")
		require.NoError(b, err)
	}
}

func BenchmarkInsert1(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := []datastream.Datapoint{datastream.Datapoint{
			Timestamp: float64(n + 1),
			Data:      true,
		}}
		err = pathOperator.InsertStream("streamdb_test/user/mystream", data, false)
		require.NoError(b, err)
	}
}

func BenchmarkStreamLength(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	data := make([]datastream.Datapoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = datastream.Datapoint{
			Timestamp: float64(i + 1),
			Data:      true,
		}
	}
	err = pathOperator.InsertStream("streamdb_test/user/mystream", data, false)
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err = pathOperator.LengthStream("streamdb_test/user/mystream")
		require.NoError(b, err)
	}
}

func BenchmarkInsert1000(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		data := make([]datastream.Datapoint, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = datastream.Datapoint{
				Timestamp: float64(1000*n + i + 1),
				Data:      true,
			}
		}
		err = pathOperator.InsertStream("streamdb_test/user/mystream", data, false)
		require.NoError(b, err)
	}
	b.StopTimer()
}

func BenchmarkRead1000(b *testing.B) {
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	go database.RunWriter()

	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	data := make([]datastream.Datapoint, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = datastream.Datapoint{
			Timestamp: float64(i + 1),
			Data:      true,
		}
	}
	err = pathOperator.InsertStream("streamdb_test/user/mystream", data, false)
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
	database, db, err := OpenDb(b)
	require.NoError(b, err)
	defer database.Close()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	o, err := NewUserLoginOperator(db, "streamdb_test", "mypass")
	require.NoError(b, err)
	pathOperator := interfaces.PathOperatorMixin{o}

	go database.RunWriter()

	db.CreateUser("streamdb_test", "root@localhost", "mypass")
	require.NoError(b, db.CreateStream("streamdb_test/user/mystream", `{"type": "boolean"}`))

	data := make([]datastream.Datapoint, 950)
	for i := 0; i < 950; i++ {
		data[i] = datastream.Datapoint{
			Timestamp: float64(i + 1),
			Data:      true,
		}
	}
	err = pathOperator.InsertStream("streamdb_test/user/mystream", data, false)
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
