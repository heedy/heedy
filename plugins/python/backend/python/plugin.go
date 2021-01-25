package python

import (
	"path"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/sirupsen/logrus"
)

const PluginName = "python"

// setupDefaultPython searches the system for a supported python interpreter,
// so that heedy can automatically have working python support without requiring
// manual intervention. No packages will be installed unless a plugin is run
// explicitly requiring those packages, so the user can change the interpreter
// to anything they want (including a virtualenv)
func setupDefaultPython(db *database.AdminDB) error {
	pypath, err := SearchPython()
	if err != nil {
		logrus.Warn("No supported Python interpreter found - you will need to configure one manually.")
		return nil
	}
	a := db.Assets()
	logrus.Infof("Using python from %s", pypath)
	return assets.WriteConfig(path.Join(a.FolderPath, "heedy.conf"), &assets.Configuration{
		Plugins: map[string]*assets.Plugin{
			"python": &assets.Plugin{
				Settings: map[string]interface{}{
					"path": pypath,
				},
			},
		},
	})
}

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {

	run.Builtin.Add(&run.BuiltinRunner{
		Key:     PluginName,
		Start:   Start,
		Handler: Handler,
	})

	// When creating the database, try to find a supported interpreter
	database.AddCreateHook(setupDefaultPython)
}
