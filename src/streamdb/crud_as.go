package streamdb

/**
This file provides the unified database public interface for the timebatchdb and
the users database.

If you want to connect to either of these, it is probably best to use this
package as it provides many conveniences.
**/

import (
	"errors"
	"streamdb/users"
	//"streamdb/timebatchdb"
	//"streamdb/dtypes"
)

type Permission int

const (
	USER           Permission = iota // the device is a user
	ACTIVE                           // The device is enabled
	ADMIN                            // The device is a superdevice (global superuser)
	WRITE                            // The device can write to user feeds
	WRITE_ANYWHERE                   // The device can write to any of a user's feeds
	MODIFY_USER                      // The device can modify it's owner
)

var (
	PrivilegeError            = errors.New("Insufficient privileges")
	InvalidParameterError     = errors.New("Invalid Parameter Recieved")
	super_privilege           = []Permission{ACTIVE, ADMIN}
	modify_user_privilege     = []Permission{ACTIVE, MODIFY_USER}
	active_privilege          = []Permission{ACTIVE}
	user_authorized_privilege = []Permission{ADMIN, USER, MODIFY_USER}
	write_privilege           = []Permission{WRITE, ACTIVE}
	write_anywhere_privilege  = []Permission{WRITE_ANYWHERE, ACTIVE}
	read_privilege            = []Permission{ACTIVE}
)

// Checks to see if the device has the listed permissions
func HasPermissions(d *users.Device, permissions []Permission) bool {
	for _, p := range permissions {
		switch p {
		case USER:
			if !d.IsUser() {
				return false
			}
		case ACTIVE:
			if !d.IsActive() {
				return false
			}
		case ADMIN:
			if !d.IsAdmin() {
				return false
			}
		case WRITE:
			if !d.WriteAllowed() {
				return false
			}
		case WRITE_ANYWHERE:
			if !d.WriteAnywhereAllowed() {
				return false
			}
		case MODIFY_USER:
			if !d.CanModifyUser() {
				return false
			}
		}
	}

	return true
}

func HasAnyPermission(d *users.Device, permissions []Permission) bool {
	for _, p := range permissions {
		switch p {
		case USER:
			if d.IsUser() {
				return true
			}
		case ACTIVE:
			if d.IsActive() {
				return true
			}
		case ADMIN:
			if d.IsAdmin() {
				return true
			}
		case WRITE:
			if d.WriteAllowed() {
				return true
			}
		case WRITE_ANYWHERE:
			if d.WriteAnywhereAllowed() {
				return true
			}
		case MODIFY_USER:
			if d.CanModifyUser() {
				return true
			}
		}
	}

	return false
}

// Create user as a proxy.
func (db *Database) CreateUserAs(device *users.Device, Name, Email, Password string) (id int64, err error) {
	if device == nil {
		return -1, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return -1, PrivilegeError
	}

	return db.CreateUser(Name, Email, Password)
}

// Returns a User instance if a user exists with the given email address
func (db *Database) ReadUserByEmailAs(device *users.Device, email string) (*users.User, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return nil, PrivilegeError
	}

	return db.ReadUserByEmail(email)
}

// Attempts to read the user by name as the given device.
func (db *Database) ReadUserByNameAs(device *users.Device, name string) (*users.User, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return nil, PrivilegeError
	}

	return db.ReadUserByName(name)
}

// Attempts to read the user by id as the given device
func (db *Database) ReadUserByIdAs(device *users.Device, id int64) (*users.User, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return nil, PrivilegeError
	}

	return db.ReadUserById(id)

}

// Reads all users, or the device's owner if not allowed all
func (db *Database) ReadAllUsersAs(device *users.Device) ([]*users.User, error) {
	if device == nil {
		return []*users.User{}, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return db.ReadAllUsers()
	}
	user, err := db.ReadUserById(device.OwnerId)

	if err != nil {
		return []*users.User{}, err
	}

	return []*users.User{user}, nil
}

// Attempts to update a user as the given device.
func (db *Database) UpdateUserAs(device *users.Device, user *users.User) error {
	if device == nil || user == nil {
		return InvalidParameterError
	}

	if !HasPermissions(device, modify_user_privilege) || device.OwnerId != user.Id {
		return PrivilegeError
	}

	return db.UpdateUser(user)
}

// Attempts to delete a user as the given device.
func (db *Database) DeleteUserAs(device *users.Device, id int64) error {
	if device == nil {
		return InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return PrivilegeError
	}

	return db.DeleteUser(id)
}

// Attempts to create a phone carrier as the given device
func (db *Database) CreatePhoneCarrierAs(device *users.Device, name, emailDomain string) (int64, error) {
	if device == nil {
		return -1, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return -1, PrivilegeError
	}

	return db.CreatePhoneCarrier(name, emailDomain)
}

// ReadPhoneCarrierByIdAs attempts to select a phone carrier from the database given its ID
func (db *Database) ReadPhoneCarrierByIdAs(device *users.Device, Id int64) (*users.PhoneCarrier, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	if !HasPermissions(device, active_privilege) {
		return nil, PrivilegeError
	}

	// currently no permissions needed for this
	return db.ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (db *Database) ReadAllPhoneCarriersAs(device *users.Device) ([]*users.PhoneCarrier, error) {
	if device == nil {
		return []*users.PhoneCarrier{}, InvalidParameterError
	}

	if !HasPermissions(device, active_privilege) {
		return []*users.PhoneCarrier{}, PrivilegeError
	}

	return db.ReadAllPhoneCarriers()
}

// Attempts to update the phone carrier as the given device
func (db *Database) UpdatePhoneCarrierAs(device *users.Device, carrier *users.PhoneCarrier) error {
	if carrier == nil || device == nil {
		return InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return PrivilegeError
	}

	return db.UpdatePhoneCarrier(carrier)
}

// Attempts to delete the phone carrier as the given device
func (db *Database) DeletePhoneCarrierAs(device *users.Device, carrierId int64) error {
	if device == nil {
		return InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return PrivilegeError
	}

	if carrierId < 0 {
		return InvalidParameterError
	}

	return db.DeletePhoneCarrier(carrierId)
}

func (db *Database) CreateDeviceAs(device *users.Device, Name string, Owner *users.User) (int64, error) {
	if device == nil || Owner == nil {
		return 0, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) || device.OwnerId != Owner.Id {
		return 0, PrivilegeError
	}

	return db.CreateDevice(Name, Owner)
}

func (db *Database) ReadDevicesForUserIdAs(device *users.Device, Id int64) ([]*users.Device, error) {
	if device == nil {
		return []*users.Device{}, InvalidParameterError
	}

	if !HasPermissions(device, active_privilege) || device.OwnerId != Id {
		return []*users.Device{}, PrivilegeError
	}

	return db.ReadDevicesForUserId(Id)
}

func (db *Database) ReadDeviceByIdAs(device *users.Device, Id int64) (*users.Device, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	if !HasPermissions(device, active_privilege) {
		return nil, PrivilegeError
	}

	dev, err := db.ReadDeviceById(Id)

	if err != nil {
		return nil, err
	}

	if device.OwnerId != dev.OwnerId {
		return nil, PrivilegeError
	}

	return dev, nil

}

func (db *Database) ReadDeviceByApiKeyAs(device *users.Device, Key string) (*users.Device, error) {
	if device == nil || Key == "" {
		return nil, InvalidParameterError
	}

	if !HasPermissions(device, super_privilege) {
		return nil, PrivilegeError
	}

	return db.ReadDeviceByApiKey(Key)
}

func (db *Database) UpdateDeviceAs(device *users.Device, update *users.Device) error {
	if device == nil || update == nil {
		return InvalidParameterError
	}

	// same device or appropriate permissions
	if HasPermissions(device, active_privilege) && (device.Id == update.Id || HasAnyPermission(device, user_authorized_privilege)) {
		if !HasPermissions(device, super_privilege) {
			update.Superdevice = false
			update.OwnerId = device.OwnerId
		}

		return db.UpdateDevice(update)
	}

	return PrivilegeError
}

func (db *Database) DeleteDeviceAs(device *users.Device, Id int64) error {
	if device == nil {
		return InvalidParameterError
	}

	if HasPermissions(device, active_privilege) && HasAnyPermission(device, user_authorized_privilege) {
		// remove important bits
		return db.DeleteDevice(Id)
	}

	return PrivilegeError
}

func (db *Database) CreateStreamAs(device *users.Device, Name, Type string, owner *users.Device) (int64, error) {
	if device == nil || owner == nil {
		return 0, InvalidParameterError
	}

	if !HasPermissions(device, active_privilege) {
		return 0, PrivilegeError
	}

	if HasAnyPermission(device, user_authorized_privilege) {
		return db.CreateStream(Name, Type, owner)
	}

	if HasPermissions(device, []Permission{WRITE}) && device.Id == owner.Id {
		return db.CreateStream(Name, Type, owner)
	}

	return 0, PrivilegeError
}

func (db *Database) ReadStreamByIdAs(device *users.Device, id int64) (*users.Stream, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	// ignore inactive devices
	if !HasPermissions(device, active_privilege) {
		return nil, PrivilegeError
	}

	// grant all superusers
	if HasPermissions(device, super_privilege) {
		return db.ReadStreamById(id)
	}

	// Check the owners for the last bit
	owner, err := db.ReadStreamOwner(id)

	// If the device is owned by the user
	if err == nil && owner.Id == device.OwnerId {
		return db.ReadStreamById(id)
	}

	return nil, PrivilegeError
}

func (db *Database) ReadStreamByDeviceAndNameAs(device *users.Device, dev *users.Device, name string) (*users.Stream, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	// ignore inactive devices
	if !HasPermissions(device, active_privilege) {
		return nil, PrivilegeError
	}

	stream, err := db.ReadStreamByDeviceIdAndName(dev.Id, name)
	if err != nil {
		return nil, err
	}

	// grant all superusers
	if HasPermissions(device, super_privilege) {
		return stream, nil
	}

	// Check the owners for the last bit
	owner, err := db.ReadStreamOwner(stream.Id)

	// If the device is owned by the user
	if err == nil && owner.Id == device.OwnerId {
		return stream, nil
	}

	return nil, PrivilegeError
}

func (db *Database) ReadStreamsByDeviceAs(device *users.Device, operand *users.Device) ([]*users.Stream, error) {
	if device == nil {
		return nil, InvalidParameterError
	}

	// ignore inactive devices
	if !HasPermissions(device, active_privilege) {
		return nil, PrivilegeError
	}

	// grant all superusers
	if HasPermissions(device, super_privilege) {
		return db.ReadStreamsByDevice(operand)
	}

	// If the device is owned by the user
	if device.OwnerId == operand.OwnerId {
		return db.ReadStreamsByDevice(operand)
	}

	return nil, PrivilegeError
}

func (db *Database) UpdateStreamAs(device *users.Device, stream *users.Stream) error {
	if device == nil || stream == nil {
		return InvalidParameterError
	}

	// ignore inactive devices
	if !HasPermissions(device, active_privilege) {
		return PrivilegeError
	}

	// grant all superusers
	if HasPermissions(device, super_privilege) {
		return db.UpdateStream(stream)
	}

	// Must be able to modify user information
	if !HasAnyPermission(device, user_authorized_privilege) {
		return PrivilegeError
	}

	// Check the owners for the last bit
	owner, err := db.ReadStreamOwner(stream.Id)

	// If the device is owned by the user
	if err == nil && owner.Id == device.OwnerId {
		return db.UpdateStream(stream)
	}

	return PrivilegeError
}

func (db *Database) DeleteStreamAs(device *users.Device, Id int64) error {
	if device == nil {
		return InvalidParameterError
	}

	// ignore inactive devices
	if !HasPermissions(device, active_privilege) {
		return PrivilegeError
	}

	// grant all superusers
	if HasPermissions(device, super_privilege) {
		return db.DeleteStream(Id)
	}

	// Must be able to modify user information
	if !HasAnyPermission(device, user_authorized_privilege) {
		return PrivilegeError
	}

	// Check the owners for the last bit
	owner, err := db.ReadStreamOwner(Id)

	// If the device is owned by the user
	if err == nil && owner.Id == device.OwnerId {
		return db.DeleteStream(Id)
	}

	return PrivilegeError
}

// Reads a stream by URI, returing all components up to an error, blank
// items will be returned as nil
func (db *Database) ReadStreamByUriAs(proxy *users.Device, user, device, stream string) (*users.User, *users.Device, *users.Stream, error) {
	if proxy == nil {
		return nil, nil, nil, InvalidParameterError
	}

	if !HasPermissions(proxy, active_privilege) {
		return nil, nil, nil, PrivilegeError
	}

	userobj, err := db.ReadUserByNameAs(proxy, user)
	if err != nil {
		return userobj, nil, nil, err
	}

	// TODO convert this to an as
	deviceobj, err := db.ReadDeviceForUserByName(userobj.Id, device)
	if err != nil {
		return userobj, deviceobj, nil, err
	}

	streamobj, err := db.ReadStreamByDeviceAndNameAs(proxy, deviceobj, stream)
	return userobj, deviceobj, streamobj, err
}
