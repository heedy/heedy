package datastream

import (
	"database/sql"

	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	log "github.com/Sirupsen/logrus"
)

var (
	sdb *SqlStore
	ds  *DataStream
	mc  *MockCache
	err error
)

//Creates a mocked out cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) StreamLength(deviceID int64, streamID int64, substream string) (int64, error) {
	args := m.Called(deviceID, streamID, substream)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockCache) Insert(deviceID, streamID int64, substream string, dpa DatapointArray, restamp bool) (int64, error) {
	args := m.Called(deviceID, streamID, substream, dpa, restamp)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockCache) DeleteDevice(deviceID int64) error {
	args := m.Called(deviceID)
	return args.Error(0)
}
func (m *MockCache) DeleteStream(deviceID, streamID int64) error {
	args := m.Called(deviceID, streamID)
	return args.Error(0)
}
func (m *MockCache) DeleteSubstream(deviceID, streamID int64, substream string) error {
	args := m.Called(deviceID, streamID, substream)
	return args.Error(0)
}
func (m *MockCache) ReadProcessingQueue() ([]Batch, error) {
	args := m.Called()
	return args.Get(0).([]Batch), args.Error(1)
}
func (m *MockCache) ReadBatches(batchnumber int) ([]Batch, error) {
	args := m.Called(batchnumber)
	return args.Get(0).([]Batch), args.Error(1)
}
func (m *MockCache) ReadRange(deviceID, streamID int64, substream string, i1, i2 int64) (DatapointArray, int64, int64, error) {
	args := m.Called(deviceID, streamID, substream, i1, i2)
	return args.Get(0).(DatapointArray), args.Get(1).(int64), args.Get(2).(int64), args.Error(3)
}
func (m *MockCache) ClearBatches(b []Batch) error {
	args := m.Called(b)
	return args.Error(0)
}
func (m *MockCache) Close() error {
	return nil
}
func (m *MockCache) Clear() error {
	return nil
}

func TestMain(m *testing.M) {
	mc = &MockCache{}
	sqldb, err := sql.Open("postgres", "sslmode=disable dbname=connectordb port=52592")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	_, err = sqldb.Exec(`CREATE TABLE IF NOT EXISTS datastream (
	    StreamId BIGINT NOT NULL,
		Substream VARCHAR,
	    EndTime DOUBLE PRECISION,
	    EndIndex BIGINT,
		Version INTEGER,
	    Data BYTEA,
	    UNIQUE (StreamId, Substream, EndIndex),
	    PRIMARY KEY (StreamId, Substream, EndIndex)
	    );`)

	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	ds, err = OpenDataStream(mc, sqldb, 2)
	if err != nil {
		log.Error(err)
		os.Exit(3)
	}
	ds.Close()

	ds, err = OpenDataStream(mc, sqldb, 2)
	if err != nil {
		log.Error(err)
		os.Exit(4)
	}
	sdb = ds.sqls

	res := m.Run()

	ds.Close()
	os.Exit(res)
}

func TestBasics(t *testing.T) {
	ds.Clear()

	mc.On("DeleteStream", int64(1), int64(2)).Return(nil)
	require.NoError(t, ds.DeleteStream(1, 2))
	mc.AssertExpectations(t)

	mc.On("StreamLength", int64(1), int64(2), "").Return(int64(0), nil)
	i, err := ds.StreamLength(1, 2, "")
	require.NoError(t, err)
	require.Equal(t, int64(0), i)
	mc.AssertExpectations(t)

	mc.On("Insert", int64(1), int64(2), "", dpa6, false).Return(int64(5), nil)
	_, err = ds.Insert(1, 2, "", dpa6, false)
	require.NoError(t, err)
	mc.AssertExpectations(t)

	mc.On("DeleteSubstream", int64(1), int64(2), "").Return(nil)
	require.NoError(t, ds.DeleteSubstream(1, 2, ""))

	mc.On("DeleteDevice", int64(1)).Return(nil)
	require.NoError(t, ds.DeleteDevice(1))
}

/*
func TestInsert(t *testing.T) {
	ds.Clear()
	ds.batchsize = 4

	require.NoError(t, ds.Insert(1, "", dpa7, false))

	require.Error(t, ds.Insert(1, "", dpa4, false))
	require.NoError(t, ds.Insert(1, "", dpa4, true))

}
*/
