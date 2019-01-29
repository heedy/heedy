package assets

import (
	"errors"
	"os"
	"path"

	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Assets holds the information that comes from loading the database folder,
// merging it with the built-in assets, and combining
type Assets struct {
	FolderPath    string
	ActivePlugins []string // List of active plugins

	Config *Configuration

	// The overlay filesystems that include the builtin assets, as well as all
	// overrides from active plugins, and user overrides
	AssetFS afero.Fs
}

// Reload the assets from scratch
func (a *Assets) Reload() error {

	builtinAssets := packr.New("assets", "../../assets")

	conf := NewConfiguration()

	// First, we load the configuration from the builtin assets
	configString, err := builtinAssets.FindString("/connectordb.conf")
	if err != nil {
		return err
	}
	if err = conf.Load(configString); err != nil {
		return err
	}

	// Next, we initialize the filesystem overlays from the builtin assets
	assetFs := NewAferoPackr(builtinAssets)

	// The os filesystem
	osfs := afero.NewOsFs()

	// Make sure that the root directory exists
	folderPathStats, err := os.Stat(a.FolderPath)
	if err != nil {
		return err
	}
	if !folderPathStats.IsDir() {
		return errors.New(pluginFolder + " is not a directory")
	}

	// Now we overlay the configuration and filesystems with the plugin assets
	for i, pluginName := range a.ActivePlugins {
		pluginFolder := path.Join(a.FolderPath, "plugins", pluginName)
		pluginFolderStats, err := os.Stat(configPath)
		if err != nil {
			return err
		}
		if !pluginFolderStats.IsDir() {
			return errors.New(pluginFolder + " is not a directory")
		}

		configPath := path.Join(pluginFolder, "connectordb.conf")
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			if err = conf.LoadFile(configPath); err != nil {
				return err
			}

		}

		assetPath := path.Join(pluginFolder, "assets")
		if assetPathStats, err := os.Stat(configPath); !os.IsNotExist(err) {
			if assetPathStats.IsDir() {
				log.Debugf("Adding %s to asset overlay")

				assetFs = afero.NewCopyOnWriteFs(assetFs, afero.NewBasePathFs(osfs, assetPath))
			}

		}
	}

	// Finally, we overlay the root directory

	configPath := path.Join(a.FolderPath, "connectordb.conf")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		if err = conf.LoadFile(configPath); err != nil {
			return err
		}

	}

	assetFs = afero.NewCopyOnWriteFs(assetFs, afero.NewBasePathFs(osfs, a.FolderPath))

	// Set the new config and assets
	a.Config = conf
	a.AssetFS = assetFs

	return nil
}

// GetAssets takes the given config directory, and returns
// a full filesystem which merges the built-in assets and
// their overwritten versions in the config folder.
// This is the core that permits distribution of ConnectorDB as a single binary
func GetAssets(configPath string, activePlugins []string) (afero.Fs, error) {

	// Assets compiled into the application
	builtinAssets := NewAferoPackr()

	if configPath == "" {
		// If no path is given for the configuration, return
		// just the builtin assets with a writable memory map over them.
		// This makes any writes ephemeral.
		afero.NewCopyOnWriteFs(builtinAssets, afero.MemMapFs())
	}

	// Now, if we are given a configuration directory,
	// we open it as an overlay over the builtin assets
	fileInfo, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		// Create the folder
		err = os.MkdirAll(configPath, 0700)
		if err != nil {
			return nil, err
		}
	}
	if !fileInfo.IsDir() {
		return nil, errors.New("A file with that name already exists")
	}

	// And now, we overlay the actual config folder
	diskAssets := afero.NewBasePathFs(afero.NewOsFs(), configPath)

	return afero.NewCopyOnWriteFs(builtinAssets, diskAssets)
}
