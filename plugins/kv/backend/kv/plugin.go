package kv

import (
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
)

const PluginName = "kv"

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {
	withversion := run.WithVersion(PluginName, SQLVersion, func(db *database.AdminDB, i *run.Info, sqlVersion int) error {
		return SQLUpdater(db, i, sqlVersion)
	})
	run.Builtin.Add(&run.BuiltinRunner{
		Key:     PluginName,
		Start:   withversion,
		Handler: Handler,
	})
	// Runs schema creation on database create instead of on first start
	database.AddCreateHook(run.WithNilInfo(withversion))
}
