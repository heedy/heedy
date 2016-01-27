package pathoperator

import (
	"connectordb"
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/users"

	"github.com/apcera/nats"
)

// Wrapper takes the underlying database, and an Operator instance, and enables path-based operations on the
// operator. The reason it can't use just the operator is due to permissions values - while for things liek AuthOperator,
// returned objects are only if permissions are given, the PathOperatorMixin needs to query the database for the information
// needed to actually create the desired query in the first place!
//
// Implementation warning: Only the ID portion of the returned user/device/stream is guaranteed to be valid, as other fields,
// including name might have been censored by previous operators (such as authoperator). Therefore functions in Wrapper
// can't rely on anything other than ID in returned users/devices/streams. Such implementations also can't call anything OTHER
// than read functions (other than their mirror function in the underlying operator) due to possible permissions errors.
type Wrapper struct {
	connectordb.Operator
}

// Wrap wraps an Operator such that it conforms the the PathOperator interface
func Wrap(o connectordb.Operator) Wrapper {
	return Wrapper{o}
}

// PathOperator is a wrapper for an Operator which simplifies querying of the ConnectorDB database.
// With a PathOperator, you don't need to deal with IDs of users/devices/streams, but rather deal with
// their named paths. For example, a stream called "mystream" is "test" user's "blah" device will be
// "test/blah/mystream" as its path.
type PathOperator interface {
	connectordb.Operator

	// ReadUser/CreateUser is already implemented in Operator, so no need for path versions of these
	// functions
	UpdateUser(username string, updates map[string]interface{}) error
	DeleteUser(username string) error

	ReadAllDevices(username string) ([]*users.Device, error)
	CreateDevice(devicepath string) error
	ReadDevice(devicepath string) (*users.Device, error)
	UpdateDevice(devicepath string, updates map[string]interface{}) error
	DeleteDevice(devicepath string) error

	ReadDeviceStreams(devicepath string) ([]*users.Stream, error)
	CreateStream(streampath, jsonschema string) error
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
