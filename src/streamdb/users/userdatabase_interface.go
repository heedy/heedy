
package users

// The generic user database interactor, this exists so we can replace it with
// a foreign function interface later if need be that will be seamlessly called
// over http
/**
type UserDatabaseInteractor interface {
    CreateDevice(Name string, OwnerId *User) (int64, error)
    CreatePhoneCarrier(Name, EmailDomain string) (int64, error)
    CreateStream(Name, Schema_Json, Defaults_Json string, owner *Device) (int64, error)
    CreateUser(Name, Email, Password string) (id int64, err error)
    DeleteDevice(Id int64) error
    DeletePhoneCarrier(carrierId int64) error
    DeleteStream(Id int64) error
    DeleteUser(id int64) error
    ReadAllPhoneCarriers() ([]*PhoneCarrier, error)
    ReadAllUsers() ([]*User, error)
    ReadDeviceByApiKey(Key string) (*Device, error)
    ReadDeviceById(Id int64) (*Device, error)
    ReadPhoneCarrierById(Id int64) (*PhoneCarrier, error)
    ReadStreamById(id int64) (*Stream, error)
    ReadUserByEmail(Email string) (*User, error)
    ReadUserById(Id int64) (*User, error)
    ReadUserByName(Name string) (*User, error)
    UpdateDevice(device *Device) error
    UpdatePhoneCarrier(carrier *PhoneCarrier) error
    UpdateStream(stream *Stream) error
    UpdateUser(user *User) error
    ValidateUser(UsernameOrEmail, Password string) bool
}
**/
