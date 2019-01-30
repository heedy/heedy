package assets

import (
	"fmt"
	"os"
	"path"

	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"

	"github.com/connectordb/connectordb/assets/config"
)

// Assets holds the information that comes from loading the database folder,
// merging it with the built-in assets, and combining
type Assets struct {
	// FolderPath is the path where the database is installed.
	// it can be "" if we are running ConnectorDB in setup mode,
	// in which case it runs solely on builtin assets
	// This is the only thing that needs to be manually initialized.
	FolderPath string

	// An override to the configuration. It is merged on top of the root configuration
	// before any special processing
	ConfigOverride *config.Configuration

	// The active configuration. This is loaded automatically
	Config *config.Configuration

	// The overlay filesystems that include the builtin assets, as well as all
	// overrides from active plugins, and user overrides. It is loaded automatically
	AssetFS afero.Fs
}

// Reload the assets from scratch
func (a *Assets) Reload() error {

	builtinAssets := packr.New("assets", "../../assets")

	// First, we load the configuration from the builtin assets
	baseConfigBytes, err := builtinAssets.Find("connectordb.conf")
	if err != nil {
		return err
	}
	baseConfiguration, err := config.LoadConfig(baseConfigBytes, "connectordb.conf")

	// Next, we initialize the filesystem overlays from the builtin assets
	assetFs := NewAferoPackr(builtinAssets)

	mergedConfiguration := baseConfiguration

	if a.FolderPath == "" {
		// If there is no folder path, we are running purely on built-in assets.
		log.Warn("No asset folder specified - running on builtin assets.")

	} else {
		// The os filesystem
		osfs := afero.NewOsFs()

		// First, we load the root config file, which will specify which plugins to activate
		configPath := path.Join(a.FolderPath, "connectordb.conf")
		rootConfiguration, err := config.LoadConfigFile(configPath)
		if err != nil {
			return err
		}

		if a.ConfigOverride != nil {
			rootConfiguration = config.MergeConfig(rootConfiguration, a.ConfigOverride)
		}

		// Next, we go through the plugin folder one by one, and add the active plugins to configuration
		// and overlay the plugin's filesystem over assets
		if rootConfiguration.ActivePlugins != nil {

			for _, pluginName := range *rootConfiguration.ActivePlugins {
				pluginFolder := path.Join(a.FolderPath, "plugins", pluginName)
				pluginFolderStats, err := os.Stat(pluginFolder)
				if err != nil {
					return err
				}
				if !pluginFolderStats.IsDir() {
					return fmt.Errorf("Could not find plugin %s at %s: not a directory", pluginName, pluginFolder)
				}

				configPath := path.Join(pluginFolder, "connectordb.conf")
				pluginConfiguration, err := config.LoadConfigFile(configPath)
				if err != nil {
					return err
				}
				mergedConfiguration = config.MergeConfig(mergedConfiguration, pluginConfiguration)

				assetFs = afero.NewCopyOnWriteFs(assetFs, afero.NewBasePathFs(osfs, pluginFolder))
			}
		}

		// Finally, we overlay the root directory and root config
		mergedConfiguration = config.MergeConfig(mergedConfiguration, rootConfiguration)
		assetFs = afero.NewCopyOnWriteFs(assetFs, afero.NewBasePathFs(osfs, a.FolderPath))

	}

	// Set the new config and assets
	a.Config = mergedConfiguration
	a.AssetFS = assetFs

	return nil
}

// NewAssets creates a full new asset system, including configuration.
func NewAssets(configPath string, override *config.Configuration) (*Assets, error) {
	a := &Assets{
		FolderPath:     configPath,
		ConfigOverride: override,
	}

	return a, a.Reload()
}
