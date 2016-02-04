package authoperator

import (
	"connectordb/authoperator/permissions"
	"connectordb/users"
	"errors"

	pconfig "config/permissions"
)

// CountUsers returns the total number of users of the entire database
func (a *AuthOperator) CountUsers() (int64, error) {
	perm := pconfig.Get()
	usr, dev, err := a.UserAndDevice()
	if err != nil {
		return 0, err
	}
	urole := permissions.GetUserRole(perm, usr)
	drole := permissions.GetDeviceRole(perm, dev)
	if !urole.CanCountUsers || !drole.CanCountUsers {
		return 0, errors.New("Don't have permissions necesaary to count users")
	}
	return a.Operator.CountUsers()
}

// ReadAllUsers reads all of the users who this device has permissions to read
func (a *AuthOperator) ReadAllUsers() ([]*users.User, error) {
	_, _, _, ua, da, err := a.getAccessLevels(-1, false, false)
	if err != nil {
		return nil, err
	}
	if !ua.CanListUsers || !da.CanListUsers {
		return nil, errors.New("You do not have permissions necessary to list users.")
	}

	// This is not particularly efficient, but it has the correct behavior, so
	// screw efficiency when I just need this working. I leave efficiency to future
	// coders.
	usrs, err := a.Operator.ReadAllUsers()
	if err != nil {
		return nil, err
	}
	result := make([]*users.User, 0, len(usrs))
	for i := range usrs {
		u, err := a.ReadUserByID(usrs[i].UserID)
		if err == nil {
			result = append(result, u)
		}
	}
	return result, nil
}

// ReadAllUsersToMap reads all of the users who this device has permissions to read to a map
func (a *AuthOperator) ReadAllUsersToMap() ([]map[string]interface{}, error) {
	_, _, _, ua, da, err := a.getAccessLevels(-1, false, false)
	if err != nil {
		return nil, err
	}
	if !ua.CanListUsers || !da.CanListUsers {
		return nil, errors.New("You do not have permissions necessary to list users.")
	}

	// See ReadAllUsers
	usrs, err := a.Operator.ReadAllUsers()
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(usrs))
	for i := range usrs {
		u, err := a.ReadUserToMap(usrs[i].Name)
		if err == nil {
			result = append(result, u)
		}
	}
	return result, nil
}

// CreateUser creates the given user if the device has user creating permissions
func (a *AuthOperator) CreateUser(name, email, password, role string, public bool) error {
	perm, u, _, ua, da, err := a.getAccessLevels(-1, public, false)
	if err != nil {
		return err
	}

	if !ua.CanCreateUser || !da.CanCreateUser {
		return errors.New("You do not have permissions necessary to create a user.")
	}

	if u.Role != role {
		uw := permissions.GetWriteAccess(perm, ua)
		dw := permissions.GetWriteAccess(perm, da)
		if !uw.UserRole || !dw.UserRole {
			return errors.New("Don't have permission to create user with different role than creator")
		}
	}

	return a.Operator.CreateUser(name, email, password, role, public)
}

// ReadUser reads the user with the given username. Any fields for which
// the device does not have permission are stripped from the resulting
func (a *AuthOperator) ReadUser(username string) (*users.User, error) {
	usr, err := a.Operator.ReadUser(username)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	// Don't repeat code unnecessarily
	return a.ReadUserByID(usr.UserID)
}

// ReadUserByID attmepts to read the user as the given device. Any fields for which
// the device does not have permission are stripped
func (a *AuthOperator) ReadUserByID(userID int64) (*users.User, error) {
	usr, err := a.Operator.ReadUserByID(userID)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}

	// A user is never self
	perm, _, _, ua, da, err := a.getAccessLevels(usr.UserID, usr.Public, false)
	if err != nil {
		return nil, err
	}
	err = permissions.DeleteDisallowedFields(perm, ua, da, "user", usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}

// ReadUserToMap reads the given user into a map, where only the permitted fields are present in the map
func (a *AuthOperator) ReadUserToMap(username string) (map[string]interface{}, error) {
	usr, err := a.Operator.ReadUser(username)
	if err != nil {
		return nil, permissions.ErrNoAccess
	}
	perm, _, _, ua, da, err := a.getAccessLevels(usr.UserID, usr.Public, false)
	if err != nil {
		return nil, err
	}
	return permissions.ReadObjectToMap(perm, ua, da, "user", usr)
}

// UpdateUserByID updates the user - fails if an attempt is made at updating fields
// for which the device does not have permission
func (a *AuthOperator) UpdateUserByID(userID int64, updates map[string]interface{}) error {
	usr, err := a.Operator.ReadUserByID(userID)
	if err != nil {
		return permissions.ErrNoAccess
	}
	perm, _, _, ua, da, err := a.getAccessLevels(usr.UserID, usr.Public, false)
	if err != nil {
		return err
	}
	err = permissions.CheckIfUpdateFieldsPermitted(perm, ua, da, "user", updates)
	if err != nil {
		return err
	}
	return a.Operator.UpdateUserByID(userID, updates)
}

// DeleteUserByID removes the given user if the device has the associated permissions
func (a *AuthOperator) DeleteUserByID(userID int64) error {
	usr, err := a.Operator.ReadUserByID(userID)
	if err != nil {
		return permissions.ErrNoAccess
	}
	// A user is never self
	_, _, _, ua, da, err := a.getAccessLevels(usr.UserID, usr.Public, false)
	if err != nil {
		return err
	}

	if !ua.CanDeleteUser || !da.CanDeleteUser {
		return errors.New("You do not have permissions necessary to delete this user.")
	}

	return a.Operator.DeleteUserByID(userID)
}
