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

	//The following are path operations that simplify working with paths
	//GetDeviceUser()
	//GetStreamUser()
	//GetStreamDevice()

	// The user read operations work pretty much as advertised
	ReadAllUsers() ([]users.User, error)

	CreateUser(username, email, password string) error
	ReadUser(username string) (*users.User, error)
	ReadUserByEmail(email string) (*users.User, error)
	UpdateUser(username string, modifieduser *users.User) error
	ChangeUserPassword(username, newpass string) error
	DeleteUser(username string) error

	//SetAdmin can set a user or a device to have administrator permissions
	SetAdmin(path string, isadmin bool) error

	//The device operations are exactly the same as user operations. You pass in device paths
	//in the form "username/devicename"
	ReadAllDevices(username string) ([]users.Device, error)

	CreateDevice(devicepath string) error
	ReadDevice(devicepath string) (*users.Device, error)
	UpdateDevice(devicepath string, modifieddevice *users.Device) error
	ChangeDeviceAPIKey(devicepath string) (apikey string, err error)
	DeleteDevice(devicepath string) error
}
