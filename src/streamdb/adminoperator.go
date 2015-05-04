package streamdb

import (
	"streamdb/users"
	"errors"
)

var (
	AdminUnsupported = errors.New("This operation is unsupported for AdminOperators")
	)

//Returns the super operator
func (db *Database) GetAdminOperator() (Operator) {
	return AdminOperator{db}
}

// The admin operator does everything as a root device
type AdminOperator struct {
	db *Database // the database this operator works on
}

func (o AdminOperator) GetDatabase() (*Database) {
	return o.db
}

// Creates a user with a username, password, and email string
func (o AdminOperator) CreateUser(username, email, password string) error {
	return o.GetDatabase().CreateUser(username, email, password)
}

func (o AdminOperator) ReadUser(username string) (*users.User, error) {
	return o.GetDatabase().ReadUserByName(username)
}

func (o AdminOperator) ReadUserById(id int64) (*users.User, error) {
	return o.GetDatabase().ReadUserById(id)
}


// Returns a User instance if a user exists with the given email address
func (o AdminOperator) ReadUserByEmail(email string) (*users.User, error) {
	return o.GetDatabase().ReadUserByEmail(email)
}

// Fetches all users from the database
func (o AdminOperator) ReadAllUsers() ([]users.User, error){
	return o.GetDatabase().ReadAllUsers()
}

// Attempts to update a user as the given device.
func (o AdminOperator) UpdateUser(user, originalUser *users.User) error {
	user.RevertUneditableFields(*originalUser, users.ROOT)
	return o.GetDatabase().UpdateUser(user)
}

// Attempts to delete a user as the given device.
func (o AdminOperator) DeleteUser(id int64) error {
	return o.GetDatabase().DeleteUser(id)
}

// Attempts to create a phone carrier as the given device
func (o AdminOperator) CreatePhoneCarrier(name, emailDomain string) error {
	return o.GetDatabase().CreatePhoneCarrier(name, emailDomain)
}

// ReadPhoneCarrierByIdAs attempts to select a phone carrier from the database given its ID
func (o AdminOperator) ReadPhoneCarrierById(Id int64) (*users.PhoneCarrier, error) {
	return o.GetDatabase().ReadPhoneCarrierById(Id)
}

// Attempts to read phone carriers as the given device
func (o AdminOperator) ReadAllPhoneCarriers() ([]users.PhoneCarrier, error) {
	return o.GetDatabase().ReadAllPhoneCarriers()
}


// Attempts to update the phone carrier as the given device
func (o AdminOperator) UpdatePhoneCarrier(carrier *users.PhoneCarrier) error {
	if carrier == nil {
		return InvalidParameterError
	}
	return o.GetDatabase().UpdatePhoneCarrier(carrier)
}

// Attempts to delete the phone carrier as the given device
func (o AdminOperator) DeletePhoneCarrier(carrierId int64) error {
	return o.GetDatabase().DeletePhoneCarrier(carrierId)
}

func (o AdminOperator) CreateDevice(Name string, Owner *users.User) error {
	if Owner == nil || Name == "" {
		return InvalidParameterError
	}
	return o.GetDatabase().CreateDevice(Name, Owner.UserId)
}

func (o AdminOperator) ReadDevicesForUser(u *users.User) ([]users.Device, error) {
	return o.GetDatabase().ReadDevicesForUserId(u.UserId)
}

func (o AdminOperator) ReadDeviceByApiKey(Key string) (*users.Device, error) {
	return o.db.ReadDeviceByApiKey(Key)
}


func (o AdminOperator) ReadDeviceById(id int64) (*users.Device, error) {
	return o.db.ReadDeviceById(id)
}

func (o AdminOperator) UpdateDevice(update *users.Device, original *users.Device) error {
	// revert the fields we're not allowed to update
	update.RevertUneditableFields(*original, users.ROOT)

	return o.db.UpdateDevice(update)
}

func (o AdminOperator) DeleteDevice(device *users.Device) error {
	if device == nil {
		return InvalidParameterError
	}
	return o.db.DeleteDevice(device.DeviceId)
}

func (o AdminOperator) CreateStream(Name, Type string, owner *users.Device) (error) {
	if owner == nil {
		return InvalidParameterError
	}

	return o.db.CreateStream(Name, Type, owner.DeviceId)
}

func (o AdminOperator) ReadStreamsByDevice(operand *users.Device) ([]users.Stream, error) {
	return o.db.ReadStreamsByDevice(operand.DeviceId)
}


// Reads a stream by id; returns it, it's parent and an error if set
func (o AdminOperator) ReadStreamById(id int64) (*users.Stream, *users.Device, error) {

	stream, err := o.db.ReadStreamById(id)
	if err != nil {
		return nil, nil, err
	}

	device, err := o.db.ReadDeviceById(stream.DeviceId)
	if err != nil {
		return nil, nil, err
	}

	return stream, device, nil
}

func (o AdminOperator) UpdateStream(d *users.Device, stream, originalStream *users.Stream) error {
	if d == nil || stream == nil || originalStream == nil {
		return InvalidParameterError
	}

	stream.RevertUneditableFields(*originalStream, users.ROOT)

	return o.db.UpdateStream(stream)
}

func (o AdminOperator) DeleteStream(toDeleteOwner *users.Device, toDeleteStream *users.Stream) error {
	return o.db.DeleteStream(toDeleteStream.StreamId)
}


func (o AdminOperator) ResolvePath(path string) (*Path, error) {
	return nil, AdminUnsupported
}
