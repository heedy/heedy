package timeseries

import (
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/heedy/pipescript/datasets/interpolators"
	"github.com/heedy/pipescript/transforms"
)

const PluginName = "timeseries"

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {
	// Register all pipescript transforms
	transforms.Register()
	interpolators.Register()

	// Initialize the plugin
	withversion := run.WithVersion(PluginName, SQLVersion, SQLUpdater)
	run.Builtin.Add(&run.BuiltinRunner{
		Key:     PluginName,
		Start:   withversion,
		Handler: Handler,
	})
	// Runs schema creation on database create instead of on first start
	database.AddCreateHook(run.WithNilInfo(withversion))
}
