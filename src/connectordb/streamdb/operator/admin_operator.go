package operator

/**

The administrator's database access operator.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>

All Rights Reserved

**/

import (
	"connectordb/streamdb/users"
	"errors"
)

var (
	ErrNotChangeable = errors.New("The given fields are not modifiable.")
)

type AdminOperator struct {
	PlainOperator
}

// UpdateDevice updates the device at devicepath to the modifed device passed in
func (o *AdminOperator) UpdateDevice(modifieddevice *users.Device) error {
	dev, err := o.ReadDeviceByID(modifieddevice.DeviceId)
	if err != nil {
		return err
	}
	if modifieddevice.RevertUneditableFields(*dev, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	return o.PlainOperator.UpdateDevice(modifieddevice)
}

//UpdateUser performs the given modifications
func (o *AdminOperator) UpdateUser(modifieduser *users.User) error {
	user, err := o.ReadUserByID(modifieduser.UserId)
	if err != nil {
		return err //Workaround for issue #81
	}
	if modifieduser.RevertUneditableFields(*user, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	return o.PlainOperator.UpdateUser(modifieduser)
}

//UpdateStream updates the stream. BUG(daniel) the function currently does not give an error
//if someone attempts to update the schema (which is an illegal operation anyways)
func (o *AdminOperator) UpdateStream(modifiedstream *users.Stream) error {
	strm, err := o.ReadStreamByID(modifiedstream.StreamId)
	if err != nil {
		return err
	}

	if modifiedstream.RevertUneditableFields(strm.Stream, users.ROOT) > 0 {
		return ErrNotChangeable
	}

	return o.PlainOperator.UpdateStream(modifiedstream)
}
