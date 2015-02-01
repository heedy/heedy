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
func (u User) ToClean() CleanUser {
    return CleanUser{Name:u.Name}
}

// Sets a new password for an account
func (u User) SetNewPassword(newPass string) {
    u.Password = calcHash(newPass, u.PasswordSalt, u.PasswordHashScheme)
}



// A sanitized version of user
type CleanUser struct {
    Name string  // The username
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
    OwnerId int  // the user that owns this device
}

// Check if the device is enabled
func (d Device) isActive() bool {
    return d.Enabled
}

// Checks if the device is enabled and a superdevice
func (d Device) isAdmin() bool {
    return d.isActive() && d.Superdevice
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
    OwnerId int
}
