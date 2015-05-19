package streamdb

import "streamdb/users"

// An Operator is an object that wraps the active streamdb databases and allows
// operations to be done on them collectively. It differs from the straight
// timebatchdb/userdb as it allows some checking to be done with regards to
// permissions and such beforehand. If at all possible you should use this
// interface to perform operations because it will remain stable, secure and
// independent of future backends we implement.
type Operator interface {

	//Returns an identifier for the device this operator is acting as.
	//AuthOperator has this as the path to the device the operator is acting as
	Name() string

	//Returns the underlying database
	Database() *Database

	//Reload makes sure that the operator is syncd with most recent changes to database
	Reload() error

	//Gets the user and device associated with the current operator
	User() (*users.User, error)
	Device() (*users.Device, error)

	//SetAdmin can set a user or a device to have administrator permissions
	SetAdmin(path string, isadmin bool) error

	// The user read operations work pretty much as advertised. Use them wisely.
	ReadAllUsers() ([]users.User, error)

	CreateUser(username, email, password string) error

	ReadUser(username string) (*users.User, error)
	ReadUserByID(userID int64) (*users.User, error)
	ReadUserByEmail(email string) (*users.User, error)

	UpdateUser(modifieduser *users.User) error
	ChangeUserPassword(username, newpass string) error

	DeleteUser(username string) error
	DeleteUserByID(userID int64) error

	//The device operations are exactly the same as user operations. You pass in device paths
	//in the form "username/devicename"
	ReadAllDevices(username string) ([]users.Device, error)
	ReadAllDevicesByUserID(userID int64) ([]users.Device, error)

	CreateDevice(devicepath string) error
	CreateDeviceByUserID(userID int64, devicename string) error

	ReadDevice(devicepath string) (*users.Device, error)
	ReadDeviceByID(deviceID int64) (*users.Device, error)
	ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error)

	UpdateDevice(modifieddevice *users.Device) error
	ChangeDeviceAPIKey(devicepath string) (apikey string, err error)

	DeleteDevice(devicepath string) error
	DeleteDeviceByID(deviceID int64) error

	//The stream operations are exactly the same as device operations. You pass in paths
	//in the form "username/devicename/streamname"
	ReadAllStreams(devicepath string) ([]Stream, error)
	ReadAllStreamsByDeviceID(deviceID int64) ([]Stream, error)

	CreateStream(streampath, jsonschema string) error
	CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error

	ReadStream(streampath string) (*Stream, error)
	ReadStreamByID(streamID int64) (*Stream, error)
	ReadStreamByDeviceID(deviceID int64, streamname string) (*Stream, error)

	UpdateStream(modifiedstream *Stream) error

	DeleteStream(streampath string) error
	DeleteStreamByID(streamID int64, substream string) error

	LengthStream(streampath string) (int64, error)
	LengthStreamByID(streamID int64) (int64, error)

	TimeToIndexStream(streampath string, time float64) (int64, error)
	TimeToIndexStreamByID(streamID int64, time float64) (int64, error)

	InsertStream(streampath string, data []Datapoint) error
	InsertStreamByID(streamID int64, data []Datapoint, substream string) error

	GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64) (DatapointReader, error)
	GetStreamTimeRangeByID(streamID int64, t1 float64, t2 float64, limit int64, substream string) (DatapointReader, error)

	GetStreamIndexRange(streampath string, i1 int64, i2 int64) (DatapointReader, error)
	GetStreamIndexRangeByID(streamID int64, i1 int64, i2 int64, substream string) (DatapointReader, error)
}
