package connectordb

import (
	"connectordb/authoperator"
	"errors"
)

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
	return authoperator.NewAuthOperator(db, dev.DeviceID)
}

// UserLogin attempts to log in using a username and password
func (db *Database) UserLogin(username, password string) (*authoperator.AuthOperator, error) {
	usr, err := db.ReadUser(username)
	if err != nil {
		return nil, err
	}
	if !usr.ValidatePassword(password) {
		return nil, errors.New("Incorrect password")
	}
	return db.AsUser(username)
}

// DeviceLogin logs in as a device with the giben api key
func (db *Database) DeviceLogin(apikey string) (*authoperator.AuthOperator, error) {
	dev, err := db.Userdb.ReadDeviceByAPIKey(apikey)
	if err != nil {
		return nil, err
	}
	return authoperator.NewAuthOperator(db, dev.DeviceID)
}
