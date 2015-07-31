package shell

import (
	"fmt"
)

// The command to add a device
type AddStream struct {
}

func (h AddStream) Help() string {
	return "Creates a new Stream at the given path, default is numerical."
}

func (h AddStream) Usage() string {
	return "addstream path [type]"
}

func (h AddStream) Execute(shell *Shell, args []string) {
	path := ""
	streamType := `{"type":"number"}`

	switch len(args) {
	default:
		fmt.Printf(Red + "Error: Wrong number of args\n" + Reset)
	case 2:
		path = args[1]
	case 3:
		path = args[1]
		streamType = args[2]
	}

	path = shell.ResolvePath(path)

	fmt.Printf("Creating Stream %v\n", path)

	err := shell.operator.CreateStream(path, streamType)

	if err != nil {
		fmt.Printf(Red+"Error: %v\n"+Reset, err.Error())
	}
}

func (h AddStream) Name() string {
	return "addstream"
}
