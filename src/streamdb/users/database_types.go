// Package users provides an API for managing user information.
package users

// User is the storage type for rows of the database.
type User struct {
    Id int64  // The primary key
    Name string  // The public username of the user
    Email string  // The user's email address

    Password string  // A hash of the user's password
    PasswordSalt string  // The password salt to be attached to the end of the password
    PasswordHashScheme string // A string representing the hashing scheme used

    Admin bool  // True/False if this is an administrator
    Phone string  // The user's phone number
    PhoneCarrier int // phone carrier id

    UploadLimit_Items int // upload limit in items/day
    ProcessingLimit_S int // processing limit in seconds/day
    StorageLimit_Gb int // storage limit in GB

    CreateTime int64 // The time the user was created
    ModifyTime int64 // the last time the user was modified
    UserGroup  int  // Unused for now, in the future we can place certain users into gropus for testing new features
}

// Converts a user to a sanitized version
func (u *User) ToClean() User {
    return User{Name:u.Name}
}


// Sets a new password for an account
func (u *User) SetNewPassword(newPass string) {
    u.Password = calcHash(newPass, u.PasswordSalt, u.PasswordHashScheme)
}


// Checks if the device is enabled and a superdevice
func (u *User) IsAdmin() bool {
    return u.Admin
}


func (u *User) OwnsDevice(device *Device) bool {
    return u.Id == device.OwnerId
}

func (u *User) ValidatePassword(password string) bool {
    return calcHash(password, u.PasswordSalt, u.PasswordHashScheme) == u.Password
}

// converts a user to a device for handling requests with a username/password
func (usr *User) ToDevice() *Device {
    requester := new(Device)
    requester.Superdevice = usr.IsAdmin()
    requester.Enabled = true
    requester.Shortname = usr.Name
    requester.Name = usr.Name
    requester.OwnerId = usr.Id
    requester.Id = -1

    requester.CanWrite = true
    requester.CanWriteAnywhere = true
    requester.UserProxy = true

    requester.user = usr

    return requester
}


// A phone carrier is a mobile service provider that has email forwarding of
// its
type PhoneCarrier struct {
    Id int64
    Name string
    EmailDomain string
}

// Devices are general purposed external and internal data users,
//
type Device struct {
    Id int64  // The primary key of this device
    Name string  // The registered name of this device, should be universally unique like "Devicename_serialnum"
    ApiKey string  // A uuid used as an api key to verify against
    Enabled bool  // Whether or not this device can do reading and writing
    Icon_PngB64 string  // a png image in base64
    Shortname string  // The human readable name of this device
    Superdevice bool  // Whether or not this is a "superdevice" which has access to the whole API
    OwnerId int64  // the user that owns this device

    CanWrite bool // Can this device write to streams? (inactive right now)
    CanWriteAnywhere bool // Can this device write to others streams? (inactive right now)
    UserProxy bool // Can this device operate as a user? (inactive right now)

    user *User // If this device is a user in disguise
}

// Checks to see if this is a pseudo-device created with User.ToDevice()
func (d *Device) IsUser() bool {
    return d.user != nil
}

// If this device was created from a user, get it otherwise return nil
// Scooby dooby doo!
func (d *Device) Unmask() *User {
    return d.user
}

// Check if the device is enabled
func (d *Device) IsActive() bool {
    return d.Enabled
}

// Checks if the device is enabled and a superdevice
func (d *Device) IsAdmin() bool {
    return d.IsActive() && d.Superdevice
}

func (d *Device) WriteAllowed() bool {
    return d.CanWrite
}

func (d *Device) WriteAnywhereAllowed() bool {
    return d.CanWriteAnywhere
}

func (d *Device) CanModifyUser() bool {
    return d.UserProxy
}

func (d *Device) IsOwnedBy(user *User) bool {
    return d.OwnerId == user.Id
}

func (d *Device) ToClean() Device {
    var tmp Device

    tmp.Id = d.Id
    tmp.Name = d.Name
    tmp.Enabled = d.Enabled
    tmp.Icon_PngB64 = d.Icon_PngB64
    tmp.Shortname = d.Shortname

    return tmp
}

type Stream struct {
    Id int64
    Name string
    Active bool
    Public bool // TODO kill me off
    Type string
    OwnerId int64
    Ephemeral bool // Currently inactive
    Output bool // Currently inactive
}

func (d *Stream) ToClean() Stream {
    return Stream{Id: d.Id,
        Name:d.Name,
        Type:d.Type,
        Ephemeral:d.Ephemeral,
        Output:d.Output}
}
