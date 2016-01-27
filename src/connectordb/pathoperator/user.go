package pathoperator

// UpdateUser updates the user with the given name to have the given updates
func (w Wrapper) UpdateUser(username string, updates map[string]interface{}) error {
	u, err := w.ReadUser(username)
	if err != nil {
		return err
	}
	return w.UpdateUserByID(u.UserID, updates)
}

//DeleteUser deletes a user given the user's name
func (w Wrapper) DeleteUser(username string) error {
	u, err := w.ReadUser(username)
	if err != nil {
		return err
	}
	return w.DeleteUserByID(u.UserID)
}
