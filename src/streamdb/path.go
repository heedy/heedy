package streamdb

/**

The Path object provides operations within a given user context, somewhat
sandboxing what they can do to the world.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved
**/

import (
	"streamdb/users"
)


// The operating environment for a particular operator request,
// the idea is you can construct one using an operator, then perform a plethora
// of operations using it.
type Path struct {
	operator Operator

	RequestUser   *users.User
	RequestDevice *users.Device
	RequestStream *users.Stream
}


// Returns the owning operator of this pagh
func (p *Path) GetOperator() Operator {
    return p.operator
}

// Deletes the user referenced in the given path
func (p *Path) DeleteUser() error {
    err := p.operator.DeleteUser(p.RequestUser.UserId)
    if err != nil {
        return err
    }

    // Remove references to items that will be deleted
    p.RequestUser = nil
    p.RequestDevice = nil
    p.RequestStream = nil
    return nil
}

// Removes the path's device
func (p *Path) DeleteDevice() error {
    err := p.operator.DeleteDevice(p.RequestDevice)
    if err != nil {
        return err
    }

    // Remove references to items that will be deleted
    p.RequestDevice = nil
    p.RequestStream = nil
    return nil
}

// Removes the path's stream
func (p *Path) DeleteStream() error {
    err := p.operator.DeleteStream(p.RequestDevice, p.RequestStream)
    if err != nil {
        return err
    }

    // Remove references to items that will be deleted
    p.RequestStream = nil
    return nil
}

// Creates a device for a path's user.
func (p *Path) CreateDevice(name string) error {
    return p.operator.CreateDevice(name, p.RequestUser)
}

// Creates a stream for a path's device
func (p *Path) CreateStream(name, stype string) error {
    return p.operator.CreateStream(name, stype, p.RequestDevice)
}

// Updates the user referenced in the path with the given user
func (p *Path) UpdateUser(update *users.User) error {
    return p.operator.UpdateUser(update, p.RequestUser)
}

// Updates the device referenced in the path with the given device
func (p *Path) UpdateDevice(update *users.Device) error {
    return p.operator.UpdateDevice(update, p.RequestDevice)
}

// Updates the stream referenced in the path with the given stream
func (p *Path) UpdateStream(update *users.Stream) error {
    return p.operator.UpdateStream(p.RequestDevice, update, p.RequestStream )
}
