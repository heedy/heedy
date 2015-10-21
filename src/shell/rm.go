package shell

/* Facilitates removal of users/devices/streams

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Removes a user/device/stream: 'rm path'"
	usage := `Usage: rm path`
	name := "rm"

	main := func(shell *Shell, args []string) uint8 {
		if len(args) < 2 {
			fmt.Println(Red + "Must supply a user/dev/stream path" + Reset)
			return 1
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

		if shell.PrintError(err) {
			return 1
		}

		fmt.Println(Green + "Removed: " + removedName + Reset)
		return 0
	}

	registerShellCommand(help, usage, name, main)
}
