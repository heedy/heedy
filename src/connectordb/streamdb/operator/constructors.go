package operator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/adminoperator"
	"connectordb/streamdb/operator/authoperator"
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/operator/plainoperator"
	"connectordb/streamdb/users"
)

type Database interface {
	GetUserDatabase() users.UserDatabase
	GetDatastream() *datastream.DataStream
	GetMessenger() *messenger.Messenger
}

/** Gets a general-purpose administrative operator without database-ruining
permissions.
**/
func NewOperator(db Database) Operator {
	return newAdminOperator(db)
}

func newAdminOperator(db Database) adminoperator.AdminOperator {
	op := newPlainOperator(db)
	return adminoperator.AdminOperator{&op}
}

func newPlainOperator(db Database) plainoperator.PlainOperator {
	return plainoperator.NewPlainOperator(db.GetUserDatabase(), db.GetDatastream(), db.GetMessenger())
}

/*
NewUserOperator creates an operator that can act on behalf of the given user; returns an error
if the username does not exist.
**/
func NewUserOperator(db Database, username string) (Operator, error) {
	bootstrapOperator := NewOperator(db)
	return authoperator.NewUserAuthOperator(bootstrapOperator, username)
}

/*
NewDeviceOperator creates an operator that keeps permissions contained to the
device at the given path.
*/
func NewDeviceOperator(db Database, devicepath string) (Operator, error) {
	bootstrapOperator := NewOperator(db)
	return authoperator.NewDeviceAuthOperator(bootstrapOperator, devicepath)
}

/*
NewDeviceApiOperator creates an operator that keeps permissions contained to the
device at the given path. Additionally, it fails to be created if the given
apikey does not match the one for the specified device.
*/
func NewDeviceApiOperator(db Database, devicepath, apikey string) (Operator, error) {
	bootstrapOperator := NewOperator(db)
	return authoperator.DeviceLoginOperator(bootstrapOperator, devicepath, apikey)

}

/*
NewDeviceIdOperator creates an operator that contains what it can do to the
scope of the device with the given id.
*/
func NewDeviceIdOperator(db Database, deviceID int64) (Operator, error) {
	bootstrapOperator := NewOperator(db)
	return authoperator.NewDeviceIdOperator(bootstrapOperator, deviceID)
}

/*
NewUserLoginOperator creates an operator that contains what it can do to the
scope of the user with the given username and password.
*/
func NewUserLoginOperator(db Database, username, password string) (Operator, error) {
	bootstrapOperator := NewOperator(db)
	return authoperator.NewUserLoginOperator(bootstrapOperator, username, password)
}
