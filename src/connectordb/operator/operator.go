package operator

import (
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/users"

	"github.com/nats-io/nats"
)

//Operator represents the functions which must be implemented in order to use ConnectorDB.
//If these functions are implemented, then Operator can use pathoperator wrapper to generate a PathOperator
// and all functionality of the database is available
type Operator interface {

	// AdminOperator returns the administrative operator (usually the *Database object).
	// This allows things like the path wrapper to be able to read underlying users
	// even when the user does not have access
	AdminOperator() PathOperator

	//Returns an identifier for the device this operator is acting as.
	//AuthOperator has this as the path to the device the operator is acting as
	Name() string

	// Gets the user associated with the current operator
	// Returns an error if the operator is not an AuthOperator
	User() (*users.User, error)

	// Device gets the device associated with the current operator
	// Returns an error if the operator is not an AuthOperator
	Device() (*users.Device, error)

	// The user/device/stream operations should be fairly self-explanatory.
	ReadAllUsers() ([]*users.User, error)
	CreateUser(name, email, password, role string, public bool) error
	ReadUser(username string) (*users.User, error)
	ReadUserByID(userID int64) (*users.User, error)
	UpdateUserByID(userID int64, updates map[string]interface{}) error
	DeleteUserByID(userID int64) error

	ReadAllDevicesByUserID(userID int64) ([]*users.Device, error)
	CreateDeviceByUserID(userID int64, devicename string) error
	ReadDeviceByID(deviceID int64) (*users.Device, error)
	ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error)
	UpdateDeviceByID(deviceID int64, updates map[string]interface{}) error
	DeleteDeviceByID(deviceID int64) error

	ReadAllStreamsByDeviceID(deviceID int64) ([]*users.Stream, error)
	CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error
	ReadStreamByID(streamID int64) (*users.Stream, error)
	ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error)
	UpdateStreamByID(streamID int64, updates map[string]interface{}) error
	DeleteStreamByID(streamID int64, substream string) error // The substream represents things like the downlink

	//These operations concern themselves with the IO of a stream
	LengthStreamByID(streamID int64, substream string) (int64, error)
	TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error)
	InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error

	/**GetStreamTimeRangeByID Reads all datapoints in the given time range (t1, t2]

	t1,t2 - Unix time in seconds with up to ns resolution
	limit - The maximum number of datapoints to return, 0 returns everything
	substream - What substream of the stream to use, use empty string.
	transform - the transformation pipeline to apply to the stream before returning it. Use "" if no transform.

	**/
	GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error)

	/**GetStreamIndexRangeByID Reads all datapoints in the given index range (i1, i2]

	i1,i2 - Index range, supports "fancy" indexing. i2 = 0 means end of stream,
			negative indices are from the end.
	substream - What substream of the stream to use, use empty string.
	transform - the transformation pipeline to apply to the stream before returning it. Use "" if no transform.
	**/
	GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64, transform string) (datastream.DataRange, error)

	/**GetShiftedStreamTimeRangeByID functions exactly as GetStreamTimeRange. The only
		difference is the extra "shift" argument, which shifts the returned data range by the given
		number of datapoints, either forwards or backwards in time
	**/
	GetShiftedStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, shift, limit int64, transform string) (datastream.DataRange, error)

	SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error)
	SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error)
	SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error)

	// CountUsers returns the number of existing users in the database at the
	// time of calling or an error if the database could not be reached.
	CountUsers() (int64, error)

	// CountStreams returns the number of existing streams in the database at the
	// time of calling or an error if the database could not be reached.
	CountStreams() (int64, error)

	// CountDevices returns the number of existing devices in the database at the
	// time of calling or an error if the database could not be reached.
	CountDevices() (int64, error)
}
