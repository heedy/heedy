package streamdb

import (
	"errors"
	"streamdb/users"
	"strings"
)

var (
	PermissionError       = errors.New("Insufficient Privileges")
	InvalidPathError      = errors.New("The given path is invalid")
	InvalidParameterError = errors.New("Invalid Parameter")
)

//Given an API key, returns the  Device object
func (db *Database) GetOperator(apikey string) (Operator, error) {
	dev, err := db.ReadDeviceByApiKey(apikey)
	if err != nil {
		return nil, err
	}
	ao := db.GetAdminOperator()
	return DeviceOperator{db, dev, ao}, nil
}

// Gets an operator for a particular device
func (db *Database) GetOperatorForDevice(device *users.Device) (Operator, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	ao := db.GetAdminOperator()
	return DeviceOperator{db, device, ao}, nil
}

//AuthenticateUser returns the user's operator given a username/password combo
func (db *Database) AuthenticateUser(username, password string) (Operator, error) {
	_, d, err := db.Login(username, password)
	if err != nil {
		return nil, err
	}
	return db.GetOperatorForDevice(d)
}

// The operatror is a database proxy for a particular device, note that these
// should not be operated indefinitely as the users.Device may change over
// time.
type DeviceOperator struct {
	db      *Database     // the database this operator works on
	dev     *users.Device // the device behind this operator
	adminOp Operator
}

func (o DeviceOperator) GetDatabase() *Database {
	return o.db
}

// Creates a user with a username, password, and email string
func (o DeviceOperator) CreateUser(username, email, password string) error {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PermissionError
	}

	return o.GetDatabase().CreateUser(username, email, password)
}

func (o DeviceOperator) ReadUser(username string) (*users.User, error) {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PermissionError
	}

	return o.GetDatabase().ReadUserByName(username)
}

func (o DeviceOperator) ReadUserById(id int64) (*users.User, error) {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PermissionError
	}

	return o.GetDatabase().ReadUserById(id)
}

// Returns a User instance if a user exists with the given email address
func (o DeviceOperator) ReadUserByEmail(email string) (*users.User, error) {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PermissionError
	}

	return o.GetDatabase().ReadUserByEmail(email)
}

// Fetches all users from the database
func (o DeviceOperator) ReadAllUsers() ([]users.User, error) {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PermissionError
	}

	return o.GetDatabase().ReadAllUsers()
}

// Attempts to update a user as the given device.
func (o DeviceOperator) UpdateUser(user, originalUser *users.User) error {
	if user == nil || originalUser == nil {
		return InvalidParameterError
	}

	permission := o.dev.RelationToUser(user)
	if !permission.Gte(users.ROOT) {
		return PermissionError
	}

	user.RevertUneditableFields(*originalUser, permission)

	return o.GetDatabase().UpdateUser(user)
}

// Attempts to delete a user as the given device.
func (o DeviceOperator) DeleteUser(id int64) error {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PermissionError
	}

	return o.GetDatabase().DeleteUser(id)
}

// Attempts to create a phone carrier as the given device
func (o DeviceOperator) CreatePhoneCarrier(name, emailDomain string) error {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PermissionError
	}

	return o.GetDatabase().CreatePhoneCarrier(name, emailDomain)
}

// ReadPhoneCarrierByIdAs attempts to select a phone carrier from the database given its ID
func (o DeviceOperator) ReadPhoneCarrierById(Id int64) (*users.PhoneCarrier, error) {
	if !o.dev.GeneralPermissions().Gte(users.ENABLED) {
		return nil, PermissionError
	}

	// currently no permissions needed for this
	return o.GetDatabase().ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (o DeviceOperator) ReadAllPhoneCarriers() ([]users.PhoneCarrier, error) {
	if !o.dev.GeneralPermissions().Gte(users.ENABLED) {
		return nil, PermissionError
	}

	return o.GetDatabase().ReadAllPhoneCarriers()
}

// Attempts to update the phone carrier as the given device
func (o DeviceOperator) UpdatePhoneCarrier(carrier *users.PhoneCarrier) error {
	if carrier == nil {
		return InvalidParameterError
	}

	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PermissionError
	}

	return o.GetDatabase().UpdatePhoneCarrier(carrier)
}

// Attempts to delete the phone carrier as the given device
func (o DeviceOperator) DeletePhoneCarrier(carrierId int64) error {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return PermissionError
	}

	return o.GetDatabase().DeletePhoneCarrier(carrierId)
}

func (o DeviceOperator) CreateDevice(Name string, Owner *users.User) error {
	if Owner == nil || Name == "" {
		return InvalidParameterError
	}

	if !o.dev.RelationToUser(Owner).Gte(users.USER) {
		return PermissionError
	}

	return o.GetDatabase().CreateDevice(Name, Owner.UserId)
}

func (o DeviceOperator) ReadDevicesForUser(u *users.User) ([]users.Device, error) {
	if !o.dev.RelationToUser(u).Gte(users.FAMILY) {
		return nil, PermissionError
	}

	return o.GetDatabase().ReadDevicesForUserId(u.UserId)
}

func (o DeviceOperator) ReadDeviceByApiKey(Key string) (*users.Device, error) {
	if !o.dev.GeneralPermissions().Gte(users.ROOT) {
		return nil, PermissionError
	}

	return o.db.ReadDeviceByApiKey(Key)
}

func (o DeviceOperator) ReadDeviceById(id int64) (*users.Device, error) {
	newdev, err := o.db.ReadDeviceById(id)
	if err != nil {
		return nil, err
	}

	if !o.dev.RelationToDevice(newdev).Gte(users.FAMILY) {
		return nil, PermissionError
	}

	return newdev, nil
}

func (o DeviceOperator) UpdateDevice(update *users.Device, original *users.Device) error {
	if update == nil || original == nil {
		return InvalidParameterError
	}

	permission := o.dev.RelationToDevice(update)
	if !permission.Gte(users.DEVICE) {
		return PermissionError
	}

	// revert the fields we're not allowed to update
	update.RevertUneditableFields(*original, permission)

	return o.db.UpdateDevice(update)
}

func (o DeviceOperator) DeleteDevice(device *users.Device) error {
	if device == nil {
		return InvalidParameterError
	}

	if !o.dev.RelationToDevice(device).Gte(users.USER) {
		return PermissionError
	}

	return o.db.DeleteDevice(device.DeviceId)
}

func (o DeviceOperator) CreateStream(Name, Type string, owner *users.Device) error {
	if owner == nil {
		return InvalidParameterError
	}

	if !o.dev.RelationToDevice(owner).Gte(users.USER) {
		return PermissionError
	}

	return o.db.CreateStream(Name, Type, owner.DeviceId)
}

func (o DeviceOperator) ReadStreamsByDevice(operand *users.Device) ([]users.Stream, error) {
	if !o.dev.RelationToDevice(operand).Gte(users.FAMILY) {
		return nil, PermissionError
	}

	return o.db.ReadStreamsByDevice(operand.DeviceId)
}

// Reads a stream by id; returns it, it's parent and an error if set
func (o DeviceOperator) ReadStreamById(id int64) (*users.Stream, *users.Device, error) {

	stream, err := o.db.ReadStreamById(id)
	if err != nil {
		return nil, nil, err
	}

	device, err := o.db.ReadDeviceById(stream.DeviceId)
	if err != nil {
		return nil, nil, err
	}

	if !o.dev.RelationToStream(stream, device).Gte(users.FAMILY) {
		return nil, nil, PermissionError
	}

	return stream, device, nil
}

func (o DeviceOperator) UpdateStream(d *users.Device, stream, originalStream *users.Stream) error {
	if d == nil || stream == nil || originalStream == nil {
		return InvalidParameterError
	}

	permission := o.dev.RelationToStream(stream, d)
	if !permission.Gte(users.USER) {
		return PermissionError
	}

	stream.RevertUneditableFields(*originalStream, permission)

	return o.db.UpdateStream(stream)
}

func (o DeviceOperator) DeleteStream(toDeleteOwner *users.Device, toDeleteStream *users.Stream) error {
	if !o.dev.RelationToStream(toDeleteStream, toDeleteOwner).Gte(users.USER) {
		return PermissionError
	}

	return o.db.DeleteStream(toDeleteStream.StreamId)
}

/**
// Returns a request environment for performing a specific query.
func (o DeviceOperator) GetRequestEnvironment(path string) (ore *OperatorRequestEnv, error) {
	u, d, s, err := ResolvePath(path)

	return &OperatorRequestEnv{o.db, o.dev, u, d, s}, err
}
**/

/**
Converts a path like user/device/stream into the literal user, device and stream

The path may only fill from the left, e.g. "user//" meaning it will only return
the user and nil for the others. Otherwise, the path may fill from the right,
e.g. "/devicename/stream" in which case the user is implicitly the user belonging
to the operator's device.

**/
func (o DeviceOperator) ResolvePath(path string) (*Path, error) {
	var err error
	var user *users.User
	var device *users.Device
	var stream *users.Stream

	pathsplit := strings.Split(path, "/")
	if len(pathsplit) != 3 {
		err = InvalidPathError
		return &Path{o, user, device, stream}, err
	}

	uname := pathsplit[0]
	dname := pathsplit[1]
	sname := pathsplit[2]

	// Parse the user
	if uname == "" {
		user, err = o.ReadUserById(o.dev.UserId)

		if err != nil {
			goto returnpath
		}
	} else {
		user, err = o.ReadUser(uname)

		if err != nil {
			goto returnpath
		}
	}

	// Parse the device
	if dname == "" {
		device = o.dev
	} else {
		device, err = o.db.ReadDeviceForUserByName(user.UserId, dname)
		if err != nil {
			goto returnpath
		}
	}

	if sname != "" {
		stream, err = o.db.ReadStreamByDeviceIdAndName(device.DeviceId, sname)
	}

returnpath:
	return &Path{o, user, device, stream}, err
}
