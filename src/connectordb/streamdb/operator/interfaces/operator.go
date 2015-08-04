package interfaces

import (
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/users"

	"connectordb/streamdb/datastream"

	"github.com/nats-io/nats"
)

//BaseOperatorInterface are the functions which must be implemented in order to use Operator.
//If these functions are implemented, then the operator is complete, and all functionality
//of the database is available
type BaseOperator interface {

	//Returns an identifier for the device this operator is acting as.
	//AuthOperator has this as the path to the device the operator is acting as
	Name() string

	// Gets the user and device associated with the current operator
	// Returns an error if the operator is not an AuthOperator
	User() (*users.User, error)

	// Device gets the device associated with the current operator
	// Returns an error if the operator is not an AuthOperator
	Device() (*users.Device, error)

	// The user read operations work pretty much as advertised. Use them wisely.
	ReadAllUsers() ([]users.User, error)
	CreateUser(username, email, password string) error
	ReadUser(username string) (*users.User, error)
	ReadUserByID(userID int64) (*users.User, error)
	UpdateUser(modifieduser *users.User) error
	DeleteUserByID(userID int64) error

	// Login checks the credentials of the user and returns the user and their
	// device if the credentials match otherwise an error. This function should
	// be used instead of any manual checks because it upgrades password
	// security as it processes.
	Login(username, password string) (*users.User, *users.Device, error)

	//The device operations are exactly the same as user operations. You pass in device paths
	//in the form "username/devicename"
	ReadAllDevicesByUserID(userID int64) ([]users.Device, error)
	CreateDeviceByUserID(userID int64, devicename string) error
	ReadDeviceByID(deviceID int64) (*users.Device, error)
	ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error)
	ReadDeviceByAPIKey(apikey string) (*users.Device, error)
	UpdateDevice(modifieddevice *users.Device) error
	DeleteDeviceByID(deviceID int64) error

	//The stream operations are exactly the same as device operations. You pass in paths
	//in the form "username/devicename/streamname"
	ReadAllStreamsByDeviceID(deviceID int64) ([]users.Stream, error)
	CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error
	ReadStreamByID(streamID int64) (*users.Stream, error)
	ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error)
	UpdateStream(modifiedstream *users.Stream) error
	DeleteStreamByID(streamID int64, substream string) error

	//These operations concern themselves with the IO of a stream
	LengthStreamByID(streamID int64, substream string) (int64, error)
	TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error)
	InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error

	/**GetStreamTimeRangeByID Reads all datapoints in the given time range (t1, t2]

	t1,t2 - Unix time in seconds with up to ns resolution
	limit - The maximum number of datapoints to return, 0 returns everything
	substream - What substream of the stream to use, use empty string.
	transform - the transformation pipeline to apply to the stream before returning it. Use "" if no transform.

	TODO push the substream to an enumerator (Downlink|Null) - | (dkumor) - not sure about this atm, since will
	want substreams to generate "index streams" in the future - so i'd leave it as is
	**/
	GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64, transform string) (datastream.DataRange, error)

	/**GetShiftedStreamTimeRangeByID functions exactly as GetStreamTimeRange. The only
		difference is the extra "shift" argument, which shifts the returned data range by the given
		number of datapoints, either forwards or backwards in time
	**/
	GetShiftedStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, shift, limit int64, transform string) (datastream.DataRange, error)

	/**GetStreamIndexRangeByID Reads all datapoints in the given index range (i1, i2]

	i1,i2 - Index range, supports "fancy" indexing. i2 = 0 means end of stream,
	        negative indices are from the end.
	substream - What substream of the stream to use, use empty string.
	transform - the transformation pipeline to apply to the stream before returning it. Use "" if no transform.
	**/
	GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64, transform string) (datastream.DataRange, error)

	SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error)
	SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error)
	// TODO also change this substream to the enum
	SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error)

	// CountUsers returns the number of existing users in the database at the
	// time of calling or an error if the database could not be reached.
	CountUsers() (uint64, error)

	// CountStreams returns the number of existing streams in the database at the
	// time of calling or an error if the database could not be reached.
	CountStreams() (uint64, error)

	// CountDevices returns the number of existing devices in the database at the
	// time of calling or an error if the database could not be reached.
	CountDevices() (uint64, error)
}

type PathOperator interface {
	// Changes the given device's api key to a new random UUID4. Returns the new
	// key
	ChangeDeviceAPIKey(devicepath string) (apikey string, err error)

	// Updates a user's password with the given one.
	ChangeUserPassword(username, newpass string) error

	// Creates a new device at the given path automatically inferring the
	// device name and user
	CreateDevice(devicepath string) error

	// Creates a new device at the given path automatically inferring the
	// device, stream and user names
	CreateStream(streampath, jsonschema string) error

	// Removes the device at the given path
	DeleteDevice(devicepath string) error

	// Removes the stream at the given path
	DeleteStream(streampath string) error

	// Deletes the user with the given name
	DeleteUser(username string) error

	// Reads all devices for the user with the given name
	ReadAllDevices(username string) ([]users.Device, error)
	// Reads all streams for the device at the given path
	ReadAllStreams(devicepath string) ([]users.Stream, error)

	ReadDevice(devicepath string) (*users.Device, error)
	ReadStream(streampath string) (*users.Stream, error)

	// Sets/removes a user or device from being admin
	SetAdmin(path string, isadmin bool) error

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

type Operator interface {
	BaseOperator
	PathOperator
}
