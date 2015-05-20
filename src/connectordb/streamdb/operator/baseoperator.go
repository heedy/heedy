package operator

import "connectordb/streamdb/users"

//BaseOperator are the functions which must be implemented in order to use Operator.
//If these functions are implemented, then the operator is complete, and all functionality
//of the database is available
type BaseOperator interface {

	//Returns an identifier for the device this operator is acting as.
	//AuthOperator has this as the path to the device the operator is acting as
	Name() string

	//Gets the user and device associated with the current operator
	User() (*users.User, error)
	Device() (*users.Device, error)

	// The user read operations work pretty much as advertised. Use them wisely.
	ReadAllUsers() ([]users.User, error)
	CreateUser(username, email, password string) error
	ReadUser(username string) (*users.User, error)
	ReadUserByID(userID int64) (*users.User, error)
	ReadUserByEmail(email string) (*users.User, error)
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
	GetStreamTimeRangeByID(streamID int64, t1 float64, t2 float64, limit int64, substream string) (DatapointReader, error)
	GetStreamIndexRangeByID(streamID int64, i1 int64, i2 int64, substream string) (DatapointReader, error)
}
