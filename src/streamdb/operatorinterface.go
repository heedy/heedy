package streamdb

import "streamdb/users"

// An operator is an object that wraps the active streamdb databases and allows
// operations to be done on them collectively. It differs from the straight
// timebatchdb/userdb as it allows some checking to be done with regards to
// permissions and such beforehand. If at all possible you should use this
// interface to perform operations because it will remain stable, secure and
// independent of future backends we implement.
type Operator interface {

	//Returns the underlying database
	Database() *Database

	//Gets the user and device associated with the current operator
	User() (*users.User, error)
	Device() (*users.Device, error)

	// Creates a user with the given name, email and password
	CreateUser(username, email, password string) error

	// The user read operations work pretty much as advertised
	ReadAllUsers() ([]users.User, error)

	ReadUser(username string) (*users.User, error)
	ReadUserByEmail(email string) (*users.User, error)
	ChangeUserPassword(username, newpass string) error
	UpdateUser(user *users.User, modifieduser users.User) error //TODO: I don't like that we have to pass the original user, but don't see how to make it better wihout making it get new user
	DeleteUser(username string) error

	//SetAdmin can set a user or a device
	SetAdmin(path string, isadmin bool) error
}
