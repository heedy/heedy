package connectordb

import (
	"connectordb/authoperator"
	"connectordb/users"
)

// DeviceAuthOperator logs in the given device object
func (db *Database) DeviceAuthOperator(dev *users.Device) (*authoperator.AuthOperator, error) {
	o, err := AddMetaLog(dev.UserID, db)
	if err != nil {
		return nil, err
	}
	return authoperator.NewAuthOperator(o, dev.DeviceID)
}

// AsUser returns the AuthOperator for the given user
func (db *Database) AsUser(username string) (*authoperator.AuthOperator, error) {
	return db.AsDevice(username + "/user")
}

// AsDevice returns the device operator
func (db *Database) AsDevice(devicepath string) (*authoperator.AuthOperator, error) {
	dev, err := db.ReadDevice(devicepath)
	if err != nil {
		return nil, err
	}

	return db.DeviceAuthOperator(dev)
}

// UserLogin attempts to log in using a username and password
func (db *Database) UserLogin(username, password string) (*authoperator.AuthOperator, error) {
	_, dev, err := db.Userdb.Login(username, password)
	if err != nil {
		return nil, err
	}

	return db.DeviceAuthOperator(dev)
}

// DeviceLogin logs in as a device with the giben api key
func (db *Database) DeviceLogin(apikey string) (*authoperator.AuthOperator, error) {
	dev, err := db.Userdb.ReadDeviceByAPIKey(apikey)
	if err != nil {
		return nil, err
	}

	return db.DeviceAuthOperator(dev)
}

// Nobody returns the operator of a "nobody" - it will behave as someone who has "nobody" permissions
func (db *Database) Nobody() *authoperator.AuthOperator {
	return authoperator.NewNobody(db)
}
