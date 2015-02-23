package streamdb

/**
This file provides the unified database public interface for the timebatchdb and
the users database.

If you want to connect to either of these, it is probably best to use this
package as it provides many conveniences.
**/

import (
    "streamdb/users"
    "errors"
    //"streamdb/timebatchdb"
    "streamdb/dtypes"
    )


type Permission int

const (
    USER Permission = iota // the device is a user
    ACTIVE  // The device is enabled
    ADMIN // The device is a superdevice (global superuser)
    WRITE // The device can write to user feeds
    WRITE_ANYWHERE // The device can write to any of a user's feeds
    MODIFY_USER // The device can modify it's owner
)

var (
    PrivligeError = errors.New("Insufficient privileges")
    InvalidParameterError = errors.New("Invalid Parameter Recieved")
    super_privlige = []Permission{ACTIVE, ADMIN}
    modify_user_privlige = []Permission{ACTIVE, MODIFY_USER}
)

// Checks to see if the device has the listed permissions
func HasPermissions(d *users.Device, permissions []Permission) bool {
    for _, p := range permissions {
        switch p {
            case USER:
                if ! d.IsUser() {
                    return false
                }
            case ACTIVE:
                if ! d.IsActive() {
                    return false
                }
            case ADMIN:
                if ! d.IsAdmin() {
                    return false
                }
            case WRITE:
                if ! d.WriteAllowed() {
                    return false
                }
            case WRITE_ANYWHERE:
                if ! d.WriteAnywhereAllowed() {
                    return false
                }
            case MODIFY_USER:
                if ! d.CanModifyUser() {
                    return false
                }
        }
    }

    return true
}


type UnifiedDB struct {
    users.UserDatabase
    dtypes.TypedDatabase
}


// Initializes the database with a local sqlite user store
func CreateLocalUnifiedDB(msgUrl, mongoUrl, mongoName, userdbPath string) (*UnifiedDB, error) {
    var udb UnifiedDB

    err := udb.InitTypedDB(msgUrl, mongoUrl, mongoName)
    if err != nil {
        return nil, err
    }

    err = udb.InitSqliteUserDatabase(userdbPath)
    if err != nil {
        return nil, err
    }

    return &udb, nil
}


// Create user as a proxy.
func (udb *UnifiedDB) CreateUserAs(device *users.Device, Name, Email, Password string) (id int64, err error) {
    if device == nil {
        return -1, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return -1, PrivligeError
    }

    return udb.CreateUser(Name, Email, Password)
}


// Returns a User instance if a user exists with the given email address
func (udb *UnifiedDB) ReadUserByEmailAs(device *users.Device, email string) (*users.User, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return nil, PrivligeError
    }

    return udb.ReadUserByEmail(email)
}

// Attempts to read the user by name as the given device.
func(udb *UnifiedDB) ReadUserByNameAs(device *users.Device, name string) (*users.User, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return nil, PrivligeError
    }

    return udb.ReadUserByName(name)
}

// Attempts to read the user by id as the given device
func(udb *UnifiedDB) ReadUserByIdAs(device *users.Device, id int64) (*users.User, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return nil, PrivligeError
    }

    return udb.ReadUserById(id)

}

// Reads all users, or the device's owner if not allowed all
func (udb *UnifiedDB) ReadAllUsersAs(device *users.Device) ([]*users.User, error) {
    if device == nil {
        return []*users.User{}, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return udb.ReadAllUsers()
    }
    user, err := udb.ReadUserById(device.OwnerId)

    if err != nil {
        return []*users.User{}, err
    }

    return []*users.User{user}, nil
}

// Attempts to update a user as the given device.
func (udb *UnifiedDB) UpdateUserAs(device *users.Device, user *users.User) error {
    if device == nil || user == nil {
        return InvalidParameterError
    }

    if ! HasPermissions(device, modify_user_privlige) || device.OwnerId != user.Id {
        return PrivligeError
    }

    return udb.UpdateUser(user)
}

// Attempts to delete a user as the given device.
func (udb *UnifiedDB) DeleteUserAs(device *users.Device, id int64) error {
    if device == nil {
        return InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return PrivligeError
    }

    return udb.DeleteUser(id)
}


// Attempts to create a phone carrier as the given device
func (udb *UnifiedDB) CreatePhoneCarrierAs(device *users.Device, name, emailDomain string) (int64, error) {
    if device == nil {
        return -1, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return -1, PrivligeError
    }

    return udb.CreatePhoneCarrier(name, emailDomain)
}


// ReadPhoneCarrierByIdAs attempts to select a phone carrier from the database given its ID
func (udb *UnifiedDB) ReadPhoneCarrierByIdAs(device *users.Device, Id int64) (*users.PhoneCarrier, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    // currently no permissions needed for this
    return udb.ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (udb *UnifiedDB) ReadAllPhoneCarriersAs(device *users.Device) ([]*users.PhoneCarrier, error) {
    if device == nil {
        return []*users.PhoneCarrier{}, InvalidParameterError
    }

    // currently no permissions needed for this
    return udb.ReadAllPhoneCarriers()
}

// Attempts to update the phone carrier as the given device
func (udb *UnifiedDB) UpdatePhoneCarrierAs(device *users.Device, carrier *users.PhoneCarrier) error {
    if carrier == nil || device == nil {
        return InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return PrivligeError
    }

    return udb.UpdatePhoneCarrier(carrier)
}

// Attempts to delete the phone carrier as the given device
func (udb *UnifiedDB) DeletePhoneCarrierAs(device *users.Device, carrierId int64) error {
    if device == nil {
        return InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return PrivligeError
    }

    if carrierId < 0 {
        return InvalidParameterError
    }

    return udb.DeletePhoneCarrier(carrierId)
}
