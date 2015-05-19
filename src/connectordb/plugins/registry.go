package plugins

import (
	"connectordb/streamdb"
	"errors"
)

var (
	knownPlugins = []plugin{}
	ErrNoPlugin  = errors.New("No command was found with the given name.")
)

// executes the plugin
type MainFunc func(sdb *streamdb.Database, args []string) error

// Prints the usage string of a plugin
type UsageFunc func()

// The plugin is a generic addition to connectordb that isn't
type plugin struct {
	name  string
	usage UsageFunc
	exec  MainFunc
}

// Registers a plugin with the plugin system.
func Register(name string, usage UsageFunc, exec MainFunc) {
	p := plugin{name, usage, exec}

	knownPlugins = append(knownPlugins, p)
}

// Runs a plugin with the given name
func Run(name string, sdb *streamdb.Database, args []string) error {
	for _, plug := range knownPlugins {
		if plug.name == name {
			return plug.exec(sdb, args)
		}
	}

	return ErrNoPlugin
}

// Returns the usage strings of all known plugins
func Usage() {
	for _, plug := range knownPlugins {
		plug.usage()
	}
}
