package notifications

import (
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
)

const PluginName = "notifications"

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {

	run.Builtin.Add(&run.BuiltinRunner{
		Key: PluginName,
		Start: run.WithVersion(PluginName, SQLVersion, func(db *database.AdminDB, i *run.Info, sqlVersion int) error {
			e := events.NewFilledHandler(db, events.GlobalHandler)
			RegisterNotificationHooks(e)

			return SQLUpdater(db, sqlVersion)
		}),
		Handler: Handler,
	})
}
