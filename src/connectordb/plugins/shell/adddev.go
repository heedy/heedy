package shell

import (
	"fmt"
)

// The command to add a device
type AddDev struct {
}

func (h AddDev) Help() string {
	return "Creates a new Device"
}

func (h AddDev) Usage() string {
	return "adddev user/dev"
}

func (h AddDev) Execute(shell *Shell, args []string) {
	if len(args) != 2 {
		fmt.Printf(Red + "Error: Wrong number of args\n" + Reset)
	}

	path := args[1]

	fmt.Printf("Creating Device %v\n", path)

	err := shell.operator.CreateDevice(path)

	if err != nil {
		fmt.Printf(Red+"Error: %v\n"+Reset, err.Error())
	}
}

func (h AddDev) Name() string {
	return "adddev"
}
