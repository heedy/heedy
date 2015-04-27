package streamdb

import (
	"streamdb/users"
)

// An operator is an object that wraps the active streamdb databases and allows
// operations to be done on them collectively. It differs from the straight
// timebatchdb/userdb as it allows some checking to be done with regards to
// permissions and such beforehand. If at all possible you should use this
// interface to perform operations because it will remain stable, secure and
// independent of future backends we implement.
type Operator interface {
	// Returns the database underlying this operator
    GetDatabase() *Database

	// Creates a user with the given name, email and password
    CreateUser(username, email, password string) error

	// The user read operations work pretty much as advertised

    ReadAllUsers() ([]users.User, error)
    ReadUser(username string) (*users.User, error)
    ReadUserByEmail(email string) (*users.User, error)
    ReadUserById(id int64) (*users.User, error)

	// Removes the user from the database
    DeleteUser(id int64) error

	// Updates the given user, the original user is the one that will be
	// validated against.
    UpdateUser(user, originalUser *users.User) error

	// Creates a device with the given name and owner
	CreateDevice(Name string, Owner *users.User) error
    ReadDeviceByApiKey(Key string) (*users.Device, error)
    ReadDeviceById(id int64) (*users.Device, error)
    ReadDevicesForUser(u *users.User) ([]users.Device, error)
    UpdateDevice(update *users.Device, original *users.Device) error
    DeleteDevice(device *users.Device) error

	// Creates a new phone carrier in the system
    CreatePhoneCarrier(name, emailDomain string) error
    ReadPhoneCarrierById(Id int64) (*users.PhoneCarrier, error)
    ReadAllPhoneCarriers() ([]users.PhoneCarrier, error)
    UpdatePhoneCarrier(carrier *users.PhoneCarrier) error
    DeletePhoneCarrier(carrierId int64) error

	// Creates a new stream with the given name, type for the given device
    CreateStream(Name, Type string, owner *users.Device) error
    ReadStreamById(id int64) (*users.Stream, *users.Device, error)
    ReadStreamsByDevice(operand *users.Device) ([]users.Stream, error)
    UpdateStream(d *users.Device, stream, originalStream *users.Stream) error
    DeleteStream(toDeleteOwner *users.Device, toDeleteStream *users.Stream) error

	/**
	Converts a path like user/device/stream into the literal user, device and stream

	The path may only fill from the left, e.g. "user//" meaning it will only return
	the user and nil for the others. Otherwise, the path may fill from the right,
	e.g. "/devicename/stream" in which case the user is implicitly the user belonging
	to the operator's device.

	**/
    ResolvePath(path string) (*Path, error)
}
