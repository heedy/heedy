package streamdb

import (
	"streamdb/users"
	"errors"
)

var (
	PERMISSION_ERROR = errors.New("Insufficient Privileges")
)

type Device struct {
	Db     *Database
	Device *users.Device
}

//Returns the Administrator device (which has all possible permissions)
//Having a nil users.Device means that it is administrator
func (db *Database) GetAdminDevice() *Device {
	return &Device{db, nil}
}

//Given an API key, returns the  Device object
func (db *Database) GetDevice(apikey string) (*Device, error) {
	dev, err := db.ReadDeviceByApiKey(apikey)
	if err != nil {
		return nil, err
	}
	return &Device{db, dev}, nil
}


// The operatror is a database proxy for a particular device, note that these
// should not be operated indefinitely as the users.Device may change over
// time.
type Operator struct {
	db *Database // the database this operator works on
	dev *users.Device // the device behind this operator
}

//BUG: josephlewis42 TODO add function to update operator periodically

// Gets the operator's user, warning this is not permission checked!
func (o *Operator) ReadOperatorUser() (*users.User, error) {
	var u users.User
	err := o.db.Get(u, "SELECT * FROM Users WHERE UserId = ?", o.dev.UserId)
	return &u, err
}

func (o *Operator) GetDevice() (*users.Device) {
	return o.dev
}

func (o *Operator) GetDatabase() (*Database) {
	return o.db
}

// Creates a user with a username, password, and email string
func (o *Operator) CreateUser(username, password, email string) error {
	if ! o.dev.IsAdmin {
		return PERMISSION_ERROR
	}

	pass, salt, hashfun := users.UpgradePassword(password)

	_, err := o.db.Exec(`INSERT INTO Users (
	    Name,
	    Email,
	    Password,
	    PasswordSalt,
	    PasswordHashScheme) VALUES (?,?,?,?,?,?);`, username, email, pass, salt, hashfun)

	return err
}

func (o *Operator) ReadUser(username string) (*users.User, error) {
	if ! o.dev.IsAdmin {
		return nil, PERMISSION_ERROR
	}

	var u users.User
	err := o.db.Get(u, "SELECT * FROM Users WHERE Name = ?", username)
	return &u, err
}

func (o *Operator) ReadUserById(id int64) (*users.User, error) {
	if ! o.dev.IsAdmin {
		return nil, PERMISSION_ERROR
	}

	var u users.User
	err := o.db.Get(u, "SELECT * FROM Users WHERE UserId = ?", id)
	return &u, err
}


// Returns a User instance if a user exists with the given email address
func (o *Operator) ReadUserByEmail(email string) (*users.User, error) {

	if ! o.dev.IsAdmin {
		return nil, PrivilegeError
	}

	return o.db.ReadUserByEmail(email)
}

// Fetches all users from the database
func (o *Operator) ReadAllUsers() ([]users.User, error){
	if ! o.dev.IsAdmin {
		return nil, PrivilegeError
	}

	return o.db.ReadAllUsers()
}



// Attempts to update a user as the given device.
func (o *Operator) UpdateUser(user *users.User) error {
	if user == nil {
		return InvalidParameterError
	}

	if ! HasPermissions(o.dev, modify_user_privilege) || o.dev.UserId != user.UserId {
		return PrivilegeError
	}

	return o.db.UpdateUser(user)
}

// Attempts to delete a user as the given device.
func (o *Operator) DeleteUser(id int64) error {
	if !o.dev.IsAdmin {
		return PrivilegeError
	}

	return o.db.DeleteUser(id)
}

// Attempts to create a phone carrier as the given device
func (o *Operator) CreatePhoneCarrier(name, emailDomain string) error {
	if !o.dev.IsAdmin {
		return PrivilegeError
	}

	return o.db.CreatePhoneCarrier(name, emailDomain)
}

// ReadPhoneCarrierByIdAs attempts to select a phone carrier from the database given its ID
func (o *Operator) ReadPhoneCarrierById(Id int64) (*users.PhoneCarrier, error) {

	if !HasPermissions(o.dev, active_privilege) {
		return nil, PrivilegeError
	}

	// currently no permissions needed for this
	return o.db.ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (o *Operator) ReadAllPhoneCarriers() ([]users.PhoneCarrier, error) {
	if !HasPermissions(o.dev, active_privilege) {
		return nil, PrivilegeError
	}

	return o.db.ReadAllPhoneCarriers()
}












// Attempts to update the phone carrier as the given device
func (o *Operator) UpdatePhoneCarrier(carrier *users.PhoneCarrier) error {
	if carrier == nil {
		return InvalidParameterError
	}

	if !o.dev.IsAdmin {
		return PrivilegeError
	}

	return o.db.UpdatePhoneCarrier(carrier)
}

// Attempts to delete the phone carrier as the given device
func (o *Operator) DeletePhoneCarrier(carrierId int64) error {
	if ! o.dev.IsAdmin {
		return PrivilegeError
	}

	return o.db.DeletePhoneCarrier(carrierId)
}

func (o *Operator) CreateDevice(Name string, Owner *users.User) error {
	if Owner == nil {
		return InvalidParameterError
	}

	if !HasPermissions(o.dev, super_privilege) || o.dev.UserId != Owner.UserId {
		return PrivilegeError
	}

	return o.db.CreateDevice(Name, Owner.UserId)
}

func (o *Operator) ReadDevicesForUserId(Id int64) ([]users.Device, error) {
	if !HasPermissions(o.dev, active_privilege) || o.dev.UserId != Id {
		return nil, PrivilegeError
	}

	return o.db.ReadDevicesForUserId(Id)
}

func (o *Operator) ReadDeviceById(Id int64) (*users.Device, error) {
	if !HasPermissions(o.dev, active_privilege) {
		return nil, PrivilegeError
	}

	dev, err := o.db.ReadDeviceById(Id)

	if err != nil {
		return nil, err
	}

	if o.dev.UserId != dev.UserId {
		return nil, PrivilegeError
	}

	return dev, nil

}

func (o *Operator) ReadDeviceByApiKey(Key string) (*users.Device, error) {
	if ! o.dev.IsAdmin {
		return nil, PrivilegeError
	}

	return o.db.ReadDeviceByApiKey(Key)
}

func (o *Operator) UpdateDevice(update *users.Device) error {
	if update == nil {
		return InvalidParameterError
	}

	// same device or appropriate permissions
	if HasPermissions(o.dev, active_privilege) && (o.dev.DeviceId == update.DeviceId || HasAnyPermission(o.dev, user_authorized_privilege)) {
		if !HasPermissions(o.dev, super_privilege) {
			update.IsAdmin = false
			update.UserId = o.dev.UserId
		}

		return o.db.UpdateDevice(update)
	}

	return PrivilegeError
}

func (o *Operator) DeleteDevice(Id int64) error {

	if HasPermissions(o.dev, active_privilege) && HasAnyPermission(o.dev, user_authorized_privilege) {
		// remove important bits
		return o.db.DeleteDevice(Id)
	}

	return PrivilegeError
}

func (o *Operator) CreateStream(Name, Type string, owner *users.Device) (error) {
	if owner == nil {
		return InvalidParameterError
	}

	if !HasPermissions(o.dev, active_privilege) {
		return PrivilegeError
	}

	if HasAnyPermission(o.dev, user_authorized_privilege) {
		return o.db.CreateStream(Name, Type, owner.DeviceId)
	}

	if HasPermissions(o.dev, []Permission{WRITE}) && o.dev.DeviceId == owner.DeviceId {
		return o.db.CreateStream(Name, Type, owner.DeviceId)
	}

	return PrivilegeError
}

func (o *Operator) ReadStreamById(id int64) (*users.Stream, error) {

	// ignore inactive devices
	if !HasPermissions(o.dev, active_privilege) {
		return nil, PrivilegeError
	}

	// grant all superusers
	if HasPermissions(o.dev, super_privilege) {
		return o.db.ReadStreamById(id)
	}

	// Check the owners for the last bit
	owner, err := o.db.ReadStreamOwner(id)

	// If the device is owned by the user
	if err == nil && owner.UserId == o.dev.UserId {
		return o.db.ReadStreamById(id)
	}

	return nil, PrivilegeError
}

func (o *Operator) ReadStreamByDeviceAndName(dev *users.Device, name string) (*users.Stream, error) {
	// ignore inactive devices
	if !HasPermissions(o.dev, active_privilege) {
		return nil, PrivilegeError
	}

	stream, err := o.db.ReadStreamByDeviceIdAndName(dev.DeviceId, name)
	if err != nil {
		return nil, err
	}

	// grant all superusers
	if HasPermissions(o.dev, super_privilege) {
		return stream, nil
	}

	// Check the owners for the last bit
	owner, err := o.db.ReadStreamOwner(stream.StreamId)

	// If the device is owned by the user
	if err == nil && owner.UserId == o.dev.UserId {
		return stream, nil
	}

	return nil, PrivilegeError
}

func (o *Operator) ReadStreamsByDevice(operand *users.Device) ([]users.Stream, error) {
	// ignore inactive devices
	if !HasPermissions(o.dev, active_privilege) {
		return nil, PrivilegeError
	}

	// grant all superusers
	if HasPermissions(o.dev, super_privilege) {
		return o.db.ReadStreamsByDevice(operand.DeviceId)
	}

	// If the device is owned by the user
	if o.dev.UserId == operand.UserId {
		return o.db.ReadStreamsByDevice(operand.DeviceId)
	}

	return nil, PrivilegeError
}

func (o *Operator) UpdateStream(stream *users.Stream) error {
	if stream == nil {
		return InvalidParameterError
	}

	// ignore inactive devices
	if !HasPermissions(o.dev, active_privilege) {
		return PrivilegeError
	}

	// grant all superusers
	if HasPermissions(o.dev, super_privilege) {
		return o.db.UpdateStream(stream)
	}

	// Must be able to modify user information
	if !HasAnyPermission(o.dev, user_authorized_privilege) {
		return PrivilegeError
	}

	// Check the owners for the last bit
	owner, err := o.db.ReadStreamOwner(stream.StreamId)

	// If the device is owned by the user
	if err == nil && owner.UserId == o.dev.UserId {
		return o.db.UpdateStream(stream)
	}

	return PrivilegeError
}

func (o *Operator) DeleteStream(Id int64) error {
	// ignore inactive devices
	if !HasPermissions(o.dev, active_privilege) {
		return PrivilegeError
	}

	// grant all superusers
	if HasPermissions(o.dev, super_privilege) {
		return o.db.DeleteStream(Id)
	}

	// Must be able to modify user information
	if !HasAnyPermission(o.dev, user_authorized_privilege) {
		return PrivilegeError
	}

	// Check the owners for the last bit
	owner, err := o.db.ReadStreamOwner(Id)

	// If the device is owned by the user
	if err == nil && owner.UserId == o.dev.UserId {
		return o.db.DeleteStream(Id)
	}

	return PrivilegeError
}

// Reads a stream by URI, returing all components up to an error, blank
// items will be returned as nil
func (o *Operator) ReadStreamByUri(user, device, stream string) (*users.User, *users.Device, *users.Stream, error) {

	if !HasPermissions(o.dev, active_privilege) {
		return nil, nil, nil, PrivilegeError
	}

	userobj, err := o.db.ReadUserByName(user)
	if err != nil {
		return userobj, nil, nil, err
	}

	// TODO convert this to an as
	deviceobj, err := o.db.ReadDeviceForUserByName(userobj.UserId, device)
	if err != nil {
		return userobj, deviceobj, nil, err
	}

	streamobj, err := o.ReadStreamByDeviceAndName(deviceobj, stream)
	return userobj, deviceobj, streamobj, err
}

/**
Converts a path like user/device/stream into the literal user, device and stream

The path may only fill from the left, e.g. "/user//" meaning it will only return
the user and nil for the others. Otherwise, the path may fill from the right,
e.g. "/devname/stream" in which case the user is implicitly the user belonging
to the operator's device.

**/
func (o *Operator) ResolvePath(path string) (*users.User, *users.Device, *users.Stream, error) {
	// TODO fill me out
	return nil, nil, nil, errors.New("not implemented")
}
