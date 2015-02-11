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
}

// Converts a user to a sanitized version
func (u User) ToClean() User {
    return User{0, u.Name, "","","","",false,"",0,0,0,0}
}


// Sets a new password for an account
func (u User) SetNewPassword(newPass string) {
    u.Password = calcHash(newPass, u.PasswordSalt, u.PasswordHashScheme)
}


// Checks if the device is enabled and a superdevice
func (u User) IsAdmin() bool {
    return u.Admin
}


func (u User) OwnsDevice(device *Device) bool {
    return u.Id == device.OwnerId
}


// converts a user to a device for handling requests with a username/password
func (usr User) ToDevice() *Device {
    requester := new(Device)
    requester.Superdevice = usr.IsAdmin()
    requester.Enabled = true
    requester.Shortname = usr.Name
    requester.Name = usr.Name
    requester.OwnerId = usr.Id
    requester.Id = -1

    requester.user = &usr

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

    user *User // If this device is a user in disguise
}

// Checks to see if this is a pseudo-device created with User.ToDevice()
func (d Device) IsUser() bool {
    return d.user != nil
}

// If this device was created from a user, get it otherwise return nil
// Scooby dooby doo!
func (d Device) Unmask() *User {
    return d.user
}

// Check if the device is enabled
func (d Device) isActive() bool {
    return d.Enabled
}

// Checks if the device is enabled and a superdevice
func (d Device) isAdmin() bool {
    return d.isActive() && d.Superdevice
}

func (d Device) IsOwnedBy(user *User) bool {
    return d.OwnerId == user.Id
}

func (d Device) ToClean() CleanDevice {
    return CleanDevice{Id: d.Id,
        Name:d.Name,
        Enabled:d.Enabled,
        Icon_PngB64:d.Icon_PngB64,
        Shortname:d.Shortname}
}

// A cleaned up version of the device that is publically accessable.
type CleanDevice struct {
    Id int64  // The primary key of the device
    Name string  // The name of the device
    Enabled bool  // Whether or not this device is enabled
    Icon_PngB64 string  // The icon of this device
    Shortname string  // The human readable shortname of this device.
}


type Stream struct {
    Id int64
    Name string
    Active bool
    Public bool
    Schema_Json string
    Defaults_Json string
    OwnerId int64
}

type CleanStream struct {
    Id int64
    Name string
    Schema_Json string
    Defaults_Json string
}

func (d Stream) ToClean() CleanStream {
    return CleanStream{Id: d.Id,
        Name:d.Name,
        Schema_Json:d.Schema_Json,
        Defaults_Json:d.Defaults_Json}
}
