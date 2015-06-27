package operator

import (
	"connectordb/streamdb/users"

	"github.com/nats-io/nats"
)

//BaseOperatorInterface are the functions which must be implemented in order to use Operator.
//If these functions are implemented, then the operator is complete, and all functionality
//of the database is available
type BaseOperatorInterface interface {

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
	ReadUserByEmail(email string) (*users.User, error)
	// It is assumed that modifieduser has the same UserId as the original.
	UpdateUser(modifieduser *users.User) error
	DeleteUserByID(userID int64) error

	//The device operations are exactly the same as user operations. You pass in device paths
	//in the form "username/devicename"
	ReadAllDevicesByUserID(userID int64) ([]users.Device, error)
	CreateDeviceByUserID(userID int64, devicename string) error
	ReadDevice(devicepath string) (*users.Device, error)
	ReadDeviceByID(deviceID int64) (*users.Device, error)
	ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error)
	UpdateDevice(modifieddevice *users.Device) error
	DeleteDeviceByID(deviceID int64) error

	//The stream operations are exactly the same as device operations. You pass in paths
	//in the form "username/devicename/streamname"
	ReadAllStreamsByDeviceID(deviceID int64) ([]Stream, error)
	CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error
	ReadStream(streampath string) (*Stream, error)
	ReadStreamByID(streamID int64) (*Stream, error)
	ReadStreamByDeviceID(deviceID int64, streamname string) (*Stream, error)
	UpdateStream(modifiedstream *Stream) error
	DeleteStreamByID(streamID int64, substream string) error

	//These operations concern themselves with the IO of a stream
	LengthStreamByID(streamID int64) (int64, error)
	TimeToIndexStreamByID(streamID int64, time float64) (int64, error)
	InsertStreamByID(streamID int64, data []Datapoint, substream string) error

	/**GetStreamTimeRangeByID Reads all datapoints in the given time range (t1, t2]

	t1,t2 - Unix time in seconds with up to ns resolution
	limit - The maximum number of datapoints to return, 0 returns everything
	substream - What substream of the stream to use, use empty string.

	TODO push the substream to an enumerator (Downlink|Null)
	**/
	GetStreamTimeRangeByID(streamID int64, t1 float64, t2 float64, limit int64, substream string) (DatapointReader, error)

	/**GetStreamIndexRangeByID Reads all datapoints in the given index range (i1, i2]

	i1,i2 - Index range, supports "fancy" indexing. i2 = 0 means end of stream,
	        negative indices are from the end.
	substream - What substream of the stream to use, use empty string.
	**/
	GetStreamIndexRangeByID(streamID int64, i1 int64, i2 int64, substream string) (DatapointReader, error)

	SubscribeUserByID(userID int64, chn chan Message) (*nats.Subscription, error)
	SubscribeDeviceByID(deviceID int64, chn chan Message) (*nats.Subscription, error)
	// TODO also change this substream to the enum
	SubscribeStreamByID(streamID int64, substream string, chn chan Message) (*nats.Subscription, error)
}
