package shell

/**

Provides the ability to remove users/devices/streams

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"fmt"
)

// The clear command
type Rm struct {
}

func (h Rm) Help() string {
	return "Removes a user/device/stream: 'rm path'"
}

func (h Rm) Usage() string {
	return "rm usr[/dev[/stream]]"
}

func (h Rm) Execute(shell *Shell, args []string) {
	if len(args) < 2 {
		fmt.Println(Red + "Must supply a user/dev/stream path" + Reset)
		return
	}

	path := shell.ResolvePath(args[1])
	usr, dev, stream := shell.ReadPath(path)

	var err error
	var removedName string

	switch {
	case stream != nil:
		err = shell.operator.DeleteStreamByID(stream.StreamId, "")
		removedName = stream.Name
	case dev != nil:
		err = shell.operator.DeleteDeviceByID(dev.DeviceId)
		removedName = dev.Name
	case usr != nil:
		err = shell.operator.DeleteUserByID(usr.UserId)
		removedName = usr.Name

	}

	if err != nil {
		fmt.Println(Red + err.Error() + Reset)
		return
	}

	fmt.Println(Green + "Removed: " + removedName + Reset)
}

func (h Rm) Name() string {
	return "rm"
}
