package database

type PublicDB struct {
	adb *AdminDB
}

func (db *PublicDB) ID() string {
	return ""
}

func (db *PublicDB) CreateUser(u *User) error {

	// Only create the user if the public group contains the user:create scope
	rows, err := db.adb.DB.Query("SELECT 1 FROM group_scopes WHERE groupid='public' and scope='user:create';")
	if err != nil {
		return err
	}
	canCreate := rows.Next()
	rows.Close()
	if !canCreate {
		return ErrAccessDenied
	}
	db.adb.CreateUser(u)
	return nil
}
