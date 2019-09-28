package api

import (
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins"
	"github.com/heedy/heedy/backend/events"
)

const PluginName = "notifications"

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {
	// Add the main handler to the server's builtin routes. The builtin heedy.conf will refer
	// to builtin://<plugin name> to access the handler
	plugins.BuiltinRoutes[PluginName] = Handler

	// Register the sql updater, so that the tables for the plugin are automatically created
	// and updated on database open
	database.RegisterPlugin(PluginName, SQLVersion, func(db *database.AdminDB,sqlVersion int) error {
		e := events.NewFilledHandler(db, events.GlobalHandler)
		RegisterNotificationHooks(e)

		return SQLUpdater(db,sqlVersion)
	})
}
