package shell

import (
	"fmt"
)

// The command to add a user
type AddUser struct {
}

func (h AddUser) Help() string {
	return "Creates a new user"
}

func (h AddUser) Usage() string {
	return ""
}

func (h AddUser) Execute(shell *Shell, args []string) {
	// TODO grant admin
	fmt.Print("Enter the name for the new user: ")
	name := shell.ReadLine()

	fmt.Print("Enter the email for the new user: ")
	email := shell.ReadLine()

	// Do the password check
	passdiff := true
	pass1 := ""
	for passdiff {
		fmt.Println("Enter password for new user:")
		fmt.Print(Password)
		pass1 = shell.ReadLine()
		fmt.Println(Reset + "Re-enter password:" + Black)
		pass2 := shell.ReadLine()
		fmt.Print(Reset)

		if pass1 == pass2 {
			passdiff = false
		} else {
			fmt.Println("Passwords did not match, type 'yes' to try again")
			decision := shell.ReadLine()
			if decision != "yes" {
				return
			}
		}
	}

	fmt.Printf("Creating User %v at %v\n", name, email)

	err := shell.operator.CreateUser(name, email, pass1)
	if err != nil {
		fmt.Printf(Red+"Error: %v\n"+Reset, err.Error())
	}
}

func (h AddUser) Name() string {
	return "adduser"
}
