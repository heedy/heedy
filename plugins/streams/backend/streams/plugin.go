package streams

import "github.com/heedy/heedy/backend/plugins/run"

const PluginName = "streams"

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {

	run.Builtin.Add(&run.BuiltinRunner{
		Key:     PluginName,
		Start:   run.WithVersion(PluginName, SQLVersion, SQLUpdater),
		Handler: Handler,
	})
}
