package users

/** This is the known userdb, it probably won't be used in production, but it
can be handy while doing testing to ensure everything returns a success.
**/

var (
	KnownDevice = Device{Name: "KnownDev"}
	KnownStream = Stream{Name: "KnownStream"}
	KnownUser   = User{Name: "KnownUser"}
)

type KnownUserdb struct {
}

func (userdb *KnownUserdb) CreateDevice(Name string, UserId int64) error {
	return nil
}

func (userdb *KnownUserdb) CreateStream(Name, Type string, DeviceId int64) error {
	return nil
}

func (userdb *KnownUserdb) CreateUser(Name, Email, Password string) error {
	return nil
}

func (userdb *KnownUserdb) DeleteDevice(Id int64) error {
	return nil
}

func (userdb *KnownUserdb) DeleteStream(Id int64) error {
	return nil
}

func (userdb *KnownUserdb) DeleteUser(UserId int64) error {
	return nil
}

func (userdb *KnownUserdb) Login(Username, Password string) (*User, *Device, error) {
	return &KnownUser, &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadAllUsers() ([]User, error) {
	return []User{KnownUser}, nil
}

func (userdb *KnownUserdb) ReadDeviceByApiKey(Key string) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadDeviceById(DeviceId int64) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadDevicesForUserId(UserId int64) ([]Device, error) {
	return []Device{KnownDevice}, nil
}

func (userdb *KnownUserdb) ReadStreamByDeviceIdAndName(DeviceId int64, streamName string) (*Stream, error) {
	return &KnownStream, nil
}

func (userdb *KnownUserdb) ReadStreamById(StreamId int64) (*Stream, error) {
	return &KnownStream, nil
}

func (userdb *KnownUserdb) ReadStreamsByDevice(DeviceId int64) ([]Stream, error) {
	return []Stream{KnownStream}, nil
}

func (userdb *KnownUserdb) ReadUserById(UserId int64) (*User, error) {
	return &KnownUser, nil
}

func (userdb *KnownUserdb) ReadUserByName(Name string) (*User, error) {
	return &KnownUser, nil
}

func (userdb *KnownUserdb) ReadUserOperatingDevice(user *User) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) UpdateDevice(device *Device) error {
	return nil
}

func (userdb *KnownUserdb) UpdateStream(stream *Stream) error {
	return nil
}

func (userdb *KnownUserdb) UpdateUser(user *User) error {
	return nil
}
