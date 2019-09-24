package api

import (
	"errors"

	"github.com/heedy/heedy/backend/database"
	"github.com/jmoiron/sqlx"
)

var SQLVersion = 1

const sqlSchema = `
	CREATE TABLE notifications (

		-- Notification target - all need to be filled
		user VARCHAR NOT NULL,
		connection VARCHAR,
		source VARCHAR,

		-- Notification content
		type VARCHAR NOT NULL, -- type of notification (info,warning,error)
		title VARCHAR NOT NULL, -- The notification title
		description VARCHAR DEFAULT '', -- Details

		CONSTRAINT user_c
			FOREIGN KEY (user)
			REFERENCES users(username)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT connection_c
			FOREIGN KEY (connection)
			REFERENCES connections(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		CONSTRAINT source_c
			FOREIGN KEY (source)
			REFERENCES sources(id)
			ON UPDATE CASCADE
			ON DELETE CASCADE,

		-- There can only be one notification of each title for each target. This allows
		-- connections to keep giving errors, which won't be duplicated
		PRIMARY KEY (user,connection,source,type,title)
	)
`

func CreateSQLData(db *sqlx.DB) error {
	_, err := db.Exec(sqlSchema)
	return err
}

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, curversion int) error {
	if curversion != 0 {
		return errors.New("Notifications database version too new")
	}
	return CreateSQLData(db.DB)
}
