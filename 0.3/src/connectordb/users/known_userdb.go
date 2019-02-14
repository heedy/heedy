/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.

This is the known userdb, it probably won't be used in production, but it
can be handy while doing testing to ensure everything returns a success.
**/
package users

var (
	KnownDevice = Device{Name: "KnownDev"}
	KnownStream = Stream{Name: "KnownStream"}
	KnownUser   = User{Name: "KnownUser"}
)

type KnownUserdb struct {
}

func (userdb *KnownUserdb) Clear() {
}

func (userdb *KnownUserdb) CreateDevice(dm *DeviceMaker) error {
	return nil
}

func (userdb *KnownUserdb) CreateStream(sm *StreamMaker) error {
	return nil
}

func (userdb *KnownUserdb) CreateUser(um *UserMaker) error {
	return nil
}

func (userdb *KnownUserdb) DeleteDevice(Id int64) error {
	return nil
}

func (userdb *KnownUserdb) DeleteStream(Id int64) error {
	return nil
}

func (userdb *KnownUserdb) DeleteUser(UserID int64) error {
	return nil
}

func (userdb *KnownUserdb) Login(Username, Password string) (*User, *Device, error) {
	return &KnownUser, &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadAllUsers() ([]*User, error) {
	return []*User{&KnownUser}, nil
}

func (userdb *KnownUserdb) ReadDeviceByAPIKey(Key string) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadDeviceByID(DeviceID int64) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadDeviceForUserByName(userid int64, devicename string) (*Device, error) {
	return &KnownDevice, nil
}

func (userdb *KnownUserdb) ReadDevicesForUserID(UserID int64) ([]*Device, error) {
	return []*Device{&KnownDevice}, nil
}

func (userdb *KnownUserdb) ReadStreamByDeviceIDAndName(DeviceID int64, streamName string) (*Stream, error) {
	return &KnownStream, nil
}

func (userdb *KnownUserdb) ReadStreamsByUser(UserID int64, downlink, public, hidden bool) ([]*DevStream, error) {
	return []*DevStream{&DevStream{Stream: KnownStream}}, nil
}

func (userdb *KnownUserdb) ReadStreamByID(StreamID int64) (*Stream, error) {
	return &KnownStream, nil
}

func (userdb *KnownUserdb) ReadStreamsByDevice(DeviceID int64) ([]*Stream, error) {
	return []*Stream{&KnownStream}, nil
}

func (userdb *KnownUserdb) ReadUserById(UserID int64) (*User, error) {
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

func (userdb *KnownUserdb) CountUsers() (int64, error) {
	return 1, nil
}

func (userdb *KnownUserdb) CountStreams() (int64, error) {
	return 1, nil
}

func (userdb *KnownUserdb) CountDevices() (int64, error) {
	return 1, nil
}
