package api

import (
	"github.com/heedy/heedy/backend/plugins"
)

const PluginName = "registry"

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {
	// Add the main handler to the server's builtin routes. The builtin heedy.conf will refer
	// to builtin://<plugin name> to access the handler
	plugins.BuiltinRoutes[PluginName] = Handler
}
