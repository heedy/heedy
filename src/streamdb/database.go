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
    active_privlige = []Permission{ACTIVE}
    user_authorized_privlige = []Permission{ADMIN, USER, MODIFY_USER}
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


type UnifiedDB struct {
    users.UserDatabase
    dtypes.TypedDatabase
}


// Initializes the database with a local sqlite user store
func CreateLocalUnifiedDB(msgUrl, mongoUrl, mongoName, userdbPath string) (*UnifiedDB, error) {
    var udb UnifiedDB

    err := udb.InitTypedDB(mongoUrl, mongoName)
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

    if ! HasPermissions(device, active_privlige) {
        return nil, PrivligeError
    }

    // currently no permissions needed for this
    return udb.ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (udb *UnifiedDB) ReadAllPhoneCarriersAs(device *users.Device) ([]*users.PhoneCarrier, error) {
    if device == nil {
        return []*users.PhoneCarrier{}, InvalidParameterError
    }

    if ! HasPermissions(device, active_privlige) {
        return []*users.PhoneCarrier{}, PrivligeError
    }

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


func (udb *UnifiedDB) CreateDeviceAs(device *users.Device, Name string, Owner *users.User) (int64, error) {
    if device == nil || Owner == nil {
        return 0, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) || device.OwnerId == Owner.Id {
        return 0, PrivligeError
    }

    return udb.CreateDevice(Name, Owner)
}

func (udb *UnifiedDB) ReadDevicesForUserIdAs(device *users.Device, Id int64) ([]*users.Device, error) {
    if device == nil {
        return []*users.Device{}, InvalidParameterError
    }

    if ! HasPermissions(device, active_privlige) || device.OwnerId != Id {
        return []*users.Device{}, PrivligeError
    }

    return udb.ReadDevicesForUserId(Id)
}

func (udb *UnifiedDB) ReadDeviceByIdAs(device *users.Device, Id int64) (*users.Device, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    if ! HasPermissions(device, active_privlige){
        return nil, PrivligeError
    }

    dev, err := udb.ReadDeviceById(Id)

    if err != nil {
        return nil, err
    }

    if device.OwnerId != dev.OwnerId {
        return nil, PrivligeError
    }

    return dev, nil

}

func (udb *UnifiedDB) ReadDeviceByApiKeyAs(device *users.Device, Key string) (*users. Device, error) {
    if device == nil || Key == "" {
        return nil, InvalidParameterError
    }

    if ! HasPermissions(device, super_privlige) {
        return nil, PrivligeError
    }

    return udb.ReadDeviceByApiKey(Key)
}

func (udb *UnifiedDB) UpdateDeviceAs(device *users.Device, update *users.Device) error {
    if device == nil || update == nil {
        return InvalidParameterError
    }

    // same device or appropriate permissions
    if HasPermissions(device, active_privlige) && (device.Id == update.Id || HasAnyPermission(device, user_authorized_privlige)) {
        return udb.UpdateDevice(update)
    }

    return PrivligeError
}


func (udb *UnifiedDB) DeleteDeviceAs(device *users.Device, Id int64) error {
    if device == nil {
        return InvalidParameterError
    }

    if HasPermissions(device, active_privlige) && HasAnyPermission(device, user_authorized_privlige) {
        return udb.DeleteDevice(Id)
    }

    return PrivligeError
}


func (udb *UnifiedDB) CreateStreamAs(device *users.Device, Name, Type string, owner *users.Device) (int64, error) {
    if device == nil || owner == nil {
        return 0, InvalidParameterError
    }

    if ! HasPermissions(device, active_privlige) {
        return 0, PrivligeError
    }

    if HasAnyPermission(device, user_authorized_privlige) {
        return udb.CreateStream(Name, Type, owner)
    }

    if HasPermissions(device, []Permission{WRITE}) && device.Id == owner.Id {
        return udb.CreateStream(Name, Type, owner)
    }

    return 0, PrivligeError
}

func (udb *UnifiedDB) ReadStreamByIdAs(device *users.Device, id int64) (*users.Stream, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    // ignore inactive devices
    if ! HasPermissions(device, active_privlige) {
        return nil, PrivligeError
    }

    // grant all superusers
    if HasPermissions(device, super_privlige) {
        return udb.ReadStreamById(id)
    }

    // Check the owners for the last bit
    owner, err := udb.ReadStreamOwner(id)

    // If the device is owned by the user
    if err == nil && owner.Id == device.OwnerId {
        return udb.ReadStreamById(id)
    }

    return nil, PrivligeError
}

func (udb *UnifiedDB) ReadStreamByDeviceAs(device *users.Device, operand *users.Device) ([]*users.Stream, error) {
    if device == nil {
        return nil, InvalidParameterError
    }

    // ignore inactive devices
    if ! HasPermissions(device, active_privlige) {
        return nil, PrivligeError
    }

    // grant all superusers
    if HasPermissions(device, super_privlige) {
        return udb.ReadStreamsByDevice(operand)
    }

    // If the device is owned by the user
    if device.OwnerId == operand.OwnerId {
        return udb.ReadStreamsByDevice(operand)
    }

    return nil, PrivligeError
}


func (udb *UnifiedDB) UpdateStreamAs(device *users.Device, stream *users.Stream) error {
    if device == nil || stream == nil{
        return InvalidParameterError
    }

    // ignore inactive devices
    if ! HasPermissions(device, active_privlige) {
        return PrivligeError
    }

    // grant all superusers
    if HasPermissions(device, super_privlige) {
        return udb.UpdateStream(stream)
    }

    // Must be able to modify user information
    if ! HasAnyPermission(device, user_authorized_privlige) {
        return PrivligeError
    }

    // Check the owners for the last bit
    owner, err := udb.ReadStreamOwner(stream.Id)

    // If the device is owned by the user
    if err == nil && owner.Id == device.OwnerId {
        return udb.UpdateStream(stream)
    }

    return PrivligeError
}

func (udb *UnifiedDB) DeleteStreamAs(device *users.Device, Id int64) error {
    if device == nil {
        return InvalidParameterError
    }

    // ignore inactive devices
    if ! HasPermissions(device, active_privlige) {
        return PrivligeError
    }

    // grant all superusers
    if HasPermissions(device, super_privlige) {
        return udb.DeleteStream(Id)
    }

    // Must be able to modify user information
    if ! HasAnyPermission(device, user_authorized_privlige) {
        return PrivligeError
    }

    // Check the owners for the last bit
    owner, err := udb.ReadStreamOwner(Id)

    // If the device is owned by the user
    if err == nil && owner.Id == device.OwnerId {
        return udb.DeleteStream(Id)
    }

    return PrivligeError
}
