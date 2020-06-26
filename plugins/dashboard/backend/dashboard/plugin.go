package dashboard

import (
	"errors"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
)

const PluginName = "dashboard"

var dbUpdate = run.WithVersion(PluginName, SQLVersion, SQLUpdater)

func StartDashboard(db *database.AdminDB, i *run.Info, h run.BuiltinHelper) error {
	err := dbUpdate(db, i, h)
	if err != nil {
		return err
	}

	// Set up the global Dashboard object
	dplugin, ok := db.Assets().Config.Plugins["dashboard"]
	if !ok {
		return errors.New("Could not find dashboard plugin configuration")
	}

	Dashboard, err = NewDashboardProcessor(db, dplugin, h)
	if err != nil {
		return err
	}

	// Set up the event handler
	events.AddHandler(Dashboard)

	return nil
}

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {
	run.Builtin.Add(&run.BuiltinRunner{
		Key:     PluginName,
		Start:   StartDashboard,
		Handler: Handler,
	})
	// Runs schema creation on database create instead of on first start
	database.AddCreateHook(run.WithNilInfo(dbUpdate))
}
