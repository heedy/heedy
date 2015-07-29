package interpolators

import (
	"connectordb/config"
	"database/sql"
	"os"
	"testing"

	"connectordb/streamdb/datastream"
	"connectordb/streamdb/datastream/rediscache"

	_ "github.com/lib/pq"

	log "github.com/Sirupsen/logrus"
)

var (
	ds  *datastream.DataStream
	err error

	dpa = datastream.DatapointArray{
		datastream.Datapoint{1., "test0", ""},
		datastream.Datapoint{2., "test1", ""},
		datastream.Datapoint{3., "test2", ""},
		datastream.Datapoint{4., "test3", ""},
		datastream.Datapoint{5., "test4", ""},
		datastream.Datapoint{6., "test5", ""},
		datastream.Datapoint{6., "test6", ""},
		datastream.Datapoint{7., "test7", ""},
		datastream.Datapoint{8., "test8", ""},
	}
	dpa2 = datastream.DatapointArray{
		datastream.Datapoint{4., "test3", ""},
		datastream.Datapoint{6., "test5", ""},
		datastream.Datapoint{6.5, "test5", ""},
		datastream.Datapoint{8., "test8", ""},
	}
)

func TestMain(m *testing.M) {

	rc, err := rediscache.NewRedisConnection(&config.DefaultOptions.RedisOptions)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	sqldb, err := sql.Open(config.DefaultOptions.SqlConnectionType, config.DefaultOptions.SqlConnectionString)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}
	ds, err = datastream.OpenDataStream(rediscache.RedisCache{rc}, sqldb, 5)
	if err != nil {
		log.Error(err)
		os.Exit(3)
	}
	ds.Clear()

	go ds.RunWriter()

	_, err = ds.Insert(0, 0, "", dpa, false)
	if err != nil {
		log.Error(err)
		os.Exit(4)
	}

	res := m.Run()

	ds.Close()
	os.Exit(res)
}
