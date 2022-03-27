package database

import (
	"database/sql"
	"fmt"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database/dbutil"
	"github.com/heedy/heedy/backend/events"
)

func readUser(adb *AdminDB, name string, o *ReadUserOptions, selectStatement string, args ...interface{}) (*User, error) {
	u := &User{}
	err := adb.Get(u, selectStatement, args...)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if o == nil || !o.Icon {
		u.Icon = nil
	}
	return u, err
}

// updateUser updates the user if the given scopeSQL returns a result
func updateUser(adb *AdminDB, u *User, scopeSQL string, args ...interface{}) error {
	userColumns, userValues, err := userUpdateQuery(u)
	if err != nil {
		return err
	}

	tx, err := adb.BeginImmediatex()
	if err != nil {
		return err
	}

	rows, err := tx.Query(scopeSQL, args...)

	if err != nil {
		return err
	}
	canEdit := rows.Next()
	rows.Close()
	if !canEdit {
		tx.Rollback()
		return ErrAccessDenied("You do not have sufficient access to edit this user")
	}

	result, err := tx.Exec(fmt.Sprintf("UPDATE users SET %s WHERE username=?;", userColumns), userValues...)
	err = GetExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err == nil && u.UserName != nil {
		// The username was changed - make sure to update the configuration
		err = adb.Assets().SwapAdmin(u.ID, *u.UserName)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
func delUser(adb *AdminDB, name string, sqlStatement string, args ...interface{}) error {
	result, err := adb.Exec(sqlStatement, args...)
	err = GetExecError(result, err)
	if err == nil {
		// When deleting a user, we also remove the user from the list of admins
		err = adb.Assets().RemAdmin(name)
	}
	return err
}

func readObject(adb *AdminDB, objectid string, o *ReadObjectOptions, selectStatement string, args ...interface{}) (*Object, error) {
	s := &Object{}
	err := adb.Get(s, selectStatement, args...)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if !s.Access.HasScope("read") {
		return nil, ErrNotFound
	}

	if o == nil || !o.Icon {
		s.Icon = nil
	}

	return s, err
}

func readApp(adb *AdminDB, cid string, o *ReadAppOptions, selectStatement string, args ...interface{}) (*App, error) {
	c := &App{}
	err := adb.Get(c, selectStatement, args...)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if o == nil || !o.Icon {
		c.Icon = nil
	}
	if o == nil || !o.AccessToken {
		if c.AccessToken != nil {
			c.AccessToken = nil
		} else {
			// Make empty access token show up as empty, so services can know
			// that no access token is available
			emptyString := ""
			c.AccessToken = &emptyString

		}

	} else {
		// Make empty access token show up as empty
		if c.AccessToken == nil {
			emptyString := ""
			c.AccessToken = &emptyString
		}
	}

	return c, err
}

// updateObject uses a select statement that returns the object type if editing is permitted
func updateObject(adb *AdminDB, s *Object, selectStatement string, args ...interface{}) error {
	// Get the object type and scope
	var sv struct {
		Stype  string     `db:"type"`
		Access ScopeArray `db:"access"`
	}
	err := adb.Get(&sv, selectStatement, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	// Now check which scope we require for the update to succeed
	if s.Name != nil || s.Owner != nil || s.App != nil || s.OwnerScope != nil {
		if !sv.Access.HasScope("update") {
			return ErrNotFound
		}
	} else {
		if !sv.Access.HasScope("update") && !sv.Access.HasScope("update:basic") {
			return ErrNotFound
		}
	}

	sColumns, sValues, err := objectUpdateQuery(adb.Assets().Config, s, sv.Stype)
	if err != nil {
		return err
	}

	sValues = append(sValues, s.ID)

	// Allow updating groups that are not users
	result, err := adb.Exec(fmt.Sprintf("UPDATE objects SET %s WHERE id=?;", sColumns), sValues...)
	return GetExecError(result, err)
}

func updateApp(adb *AdminDB, c *App, whereStatement string, args ...interface{}) (err error) {
	var tx *TxWrapper
	tx, err = adb.BeginImmediatex()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err == nil && c.Settings != nil {
				e := &events.Event{
					App:   c.ID,
					Event: "app_settings_update",
				}
				if FillEvent(adb, e) == nil {
					events.Fire(e)
				}
			}
		}
	}()

	// If either settings or schema are being updated, we need to make sure that the settings and schema are compatible
	if c.Settings != nil || c.SettingsSchema != nil {
		// If the schema was changed, we need to overwrite the settings ANYWAYS, because the defaults/extra fields might have been added
		if c.Settings == nil {
			var vsettings dbutil.JSONObject
			c.Settings = &vsettings
			err = tx.Get(c.Settings, fmt.Sprintf("SELECT settings FROM apps WHERE %s", whereStatement), args...)
		}
		var vschema dbutil.JSONObject
		if c.SettingsSchema == nil {
			err = tx.Get(&vschema, fmt.Sprintf("SELECT settings_schema FROM apps WHERE %s", whereStatement), args...)
		} else {
			vschema = *c.SettingsSchema
		}
		if err != nil {
			if err == sql.ErrNoRows {
				err = ErrNotFound
			}
			return err
		}
		ss, err := assets.NewSchema(vschema)
		if err != nil {
			return err
		}
		err = ss.ValidateAndInsertDefaults(*c.Settings)
		if err != nil {
			return err
		}
	}

	cColumns, cValues, err := appUpdateQuery(c)
	if err != nil {
		return err
	}
	cValues = append(cValues, args...)
	result, err := tx.Exec(fmt.Sprintf("UPDATE apps SET %s WHERE %s", cColumns, whereStatement), cValues...)
	return GetExecError(result, err)
}

// Here db is different, since it calls unshare
func shareObject(db DB, objectid, username string, sa *ScopeArray, scopeSQL string, args ...interface{}) error {
	adb := db.AdminDB()
	if len(sa.Scope) == 0 {
		return db.UnshareObjectFromUser(objectid, username)
	}

	if !sa.HasScope("read") {
		return ErrBadQuery("To share a object, it needs to have the read scope active")
	}

	tx, err := adb.BeginImmediatex()
	if err != nil {
		return err
	}

	rows, err := tx.Query(scopeSQL, args...)

	if err != nil {
		return err
	}
	canShare := rows.Next()
	rows.Close()
	if !canShare {
		tx.Rollback()
		return ErrAccessDenied("You do not have sufficient access to share this object")
	}

	result, err := tx.Exec("INSERT OR REPLACE INTO shared_objects(username,objectid,scope) VALUES (?,?,?);", username, objectid, sa)
	err = GetExecError(result, err)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()

}

func unshareObjectFromUser(adb *AdminDB, objectid, userid string, selectStatement string, args ...interface{}) error {
	res, err := adb.Exec(selectStatement, args...)
	return GetExecError(res, err)
}

func unshareObject(adb *AdminDB, objectid, selectStatement string, args ...interface{}) error {
	res, err := adb.Exec(selectStatement, args...)
	return GetExecError(res, err)
}

func getObjectShares(adb *AdminDB, objectid, selectStatement string, args ...interface{}) (m map[string]*ScopeArray, err error) {
	var res []struct {
		Username string
		Scope    *ScopeArray
	}

	err = adb.Select(&res, selectStatement, args...)
	if err != nil {
		return nil, err
	}

	m = make(map[string]*ScopeArray)
	for _, v := range res {
		m[v.Username] = v.Scope
	}

	return m, err
}

func listObjects(adb *AdminDB, o *ListObjectsOptions, selectStatement string, args ...interface{}) ([]*Object, error) {
	var res []*Object
	q, v, err := listObjectsQuery(o)
	if err != nil {
		return nil, err
	}

	v = append(v, args...)
	limitString := ""
	if o != nil && o.Limit != nil {
		limitString = fmt.Sprintf("LIMIT %d", *o.Limit)
	} else {
		// If no limit is given, use limit of 1000
		limitString = fmt.Sprintf("LIMIT %d", 1000)
	}
	qstring := fmt.Sprintf(selectStatement, q, limitString)

	err = adb.Select(&res, qstring, v...)
	if err != nil {
		return nil, err
	}

	// Clear icons if not needed
	if o == nil || !o.Icon {
		for r := range res {
			res[r].Icon = nil
		}
	}
	return res, nil
}

func listApps(adb *AdminDB, o *ListAppOptions, selectStatement string, args ...interface{}) ([]*App, error) {
	var res []*App
	err := adb.Select(&res, selectStatement, args...)
	if err != nil {
		return nil, err
	}
	if o == nil || !o.Icon {
		for r := range res {
			res[r].Icon = nil
		}
	}
	if o == nil || !o.AccessToken {
		for _, c := range res {
			if c.AccessToken != nil {
				c.AccessToken = nil
			} else {
				// Make empty access token show up as empty, so services can know
				// that no access token is available
				emptyString := ""
				c.AccessToken = &emptyString
			}
		}
	} else {
		for _, cc := range res {
			if cc.AccessToken == nil {
				emptyString := ""
				cc.AccessToken = &emptyString
			}
		}
	}

	return res, nil
}
