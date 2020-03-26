package dashboard

import (
	"errors"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
)

var SQLVersion = 1

const sqlSchema = `

`

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion != 0 {
		return errors.New("Dashboard database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}
