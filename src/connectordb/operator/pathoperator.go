package operator

import (
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/users"

	"github.com/nats-io/nats"
)

// PathOperator is a wrapper for Operator which simplifies querying of the ConnectorDB database.
// With a PathOperator, you don't need to deal with IDs of users/devices/streams as in Operator, but rather deal with
// their named paths. For example, a stream called "mystream" is "test" user's "blah" device will be
// "test/blah/mystream" as its path.
type PathOperator interface {
	Operator

	// ReadUser/CreateUser is already implemented in Operator, so no need for path versions of these
	// functions
	UpdateUser(username string, updates map[string]interface{}) error
	DeleteUser(username string) error

	ReadUserDevices(username string) ([]*users.Device, error)
	CreateDevice(devicepath string, d *users.DeviceMaker) error
	ReadDevice(devicepath string) (*users.Device, error)
	UpdateDevice(devicepath string, updates map[string]interface{}) error
	DeleteDevice(devicepath string) error

	ReadDeviceStreams(devicepath string) ([]*users.Stream, error)
	CreateStream(streampath string, s *users.StreamMaker) error
	ReadStream(streampath string) (*users.Stream, error)
	UpdateStream(streampath string, updates map[string]interface{}) error
	DeleteStream(streampath string) error

	GetStreamIndexRange(streampath string, i1 int64, i2 int64, transform string) (datastream.DataRange, error)
	GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error)
	GetShiftedStreamTimeRange(streampath string, t1 float64, t2 float64, ishift, limit int64, transform string) (datastream.DataRange, error)
	InsertStream(streampath string, data datastream.DatapointArray, restamp bool) error
	LengthStream(streampath string) (int64, error)

	Subscribe(path string, chn chan messenger.Message) (*nats.Subscription, error)
	SubscribeDevice(devpath string, chn chan messenger.Message) (*nats.Subscription, error)
	SubscribeStream(streampath string, chn chan messenger.Message) (*nats.Subscription, error)
	SubscribeUser(username string, chn chan messenger.Message) (*nats.Subscription, error)
	TimeToIndexStream(streampath string, time float64) (int64, error)
}
