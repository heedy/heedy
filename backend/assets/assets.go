package assets

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Assets holds the information that comes from loading the database folder,
// merging it with the built-in assets, and combining
type Assets struct {
	// FolderPath is the path where the database is installed.
	// it can be "" if we are running heedy in setup mode,
	// in which case it runs solely on builtin assets
	// This is the only thing that needs to be manually initialized.
	FolderPath string

	// An override to the configuration. It is merged on top of the root configuration
	// before any special processing
	ConfigOverride *Configuration

	// The active configuration. This is loaded automatically
	Config *Configuration

	// The overlay stack. index 0 represents built-in assets. Each index is just that stack element.
	Stack []afero.Fs

	// The overlay filesystems that include the builtin assets, as well as all
	// overrides from active plugins, and user overrides. It is loaded automatically
	FS afero.Fs
}

// Reload the assets from scratch
func (a *Assets) Reload() error {

	assetStack := make([]afero.Fs, 1)

	builtinAssets := BuiltinAssets()
	assetStack[0] = builtinAssets

	// First, we load the configuration from the builtin assets
	baseConfigBytes, err := afero.ReadFile(builtinAssets, "/heedy.conf")
	if err != nil {
		return err
	}
	baseConfiguration, err := LoadConfigBytes(baseConfigBytes, "heedy.conf")
	if err != nil {
		return err
	}

	// Some plugins come built-in. Check for the built-in plugins
	if baseConfiguration.ActivePlugins != nil {
		for _, v := range *baseConfiguration.ActivePlugins {
			_, ok := baseConfiguration.Plugins[v]
			if !ok {
				return fmt.Errorf("Builtin configuration does not define plugin '%s'", v)
			}
		}
	}

	// Next, we initialize the filesystem overlays from the builtin assets
	FS := builtinAssets

	mergedConfiguration := baseConfiguration

	if a.FolderPath == "" {
		// If there is no folder path, we are running purely on built-in assets.
		//log.Debug("No asset folder specified - running on builtin assets.")

	} else {
		// Make sure the folder path is absolute
		a.FolderPath, err = filepath.Abs(a.FolderPath)
		if err != nil {
			return err
		}

		// The os filesystem
		osfs := afero.NewOsFs()

		// First, we load the root config file, which will specify which plugins to activate
		configPath := path.Join(a.FolderPath, "heedy.conf")
		rootConfiguration, err := LoadConfigFile(configPath)
		if err != nil {
			return err
		}

		if a.ConfigOverride != nil {
			rootConfiguration = MergeConfig(rootConfiguration, a.ConfigOverride)
		}

		// Next, we go through the plugin folder one by one, and add the active plugins to configuration
		// and overlay the plugin's filesystem over assets
		if rootConfiguration.ActivePlugins != nil {

			for _, pluginName := range *rootConfiguration.ActivePlugins {
				if !strings.HasPrefix(pluginName, "-") {
					if strings.HasPrefix(pluginName, "+") {
						pluginName = pluginName[1:len(pluginName)]
					}

					pluginFolder := path.Join(a.FolderPath, "plugins", pluginName)
					pluginFolderStats, err := os.Stat(pluginFolder)
					if err != nil {
						return err
					}
					if !pluginFolderStats.IsDir() {
						return fmt.Errorf("Could not find plugin %s at %s: not a directory", pluginName, pluginFolder)
					}

					configPath := path.Join(pluginFolder, "heedy.conf")
					pluginConfiguration, err := LoadConfigFile(configPath)
					if err != nil {
						return err
					}
					mergedConfiguration = MergeConfig(mergedConfiguration, pluginConfiguration)

					pluginFs := afero.NewBasePathFs(osfs, pluginFolder)
					assetStack = append(assetStack, pluginFs)
					FS = afero.NewCopyOnWriteFs(FS, pluginFs)
				}
			}

		}

		// Finally, we overlay the root directory and root config
		mergedConfiguration = MergeConfig(mergedConfiguration, rootConfiguration)
		mainFs := afero.NewBasePathFs(osfs, a.FolderPath)
		assetStack = append(assetStack, mainFs)
		FS = afero.NewCopyOnWriteFs(FS, mainFs)

		// Get the full list of active plugins here
		mergedConfiguration.ActivePlugins = MergeStringArrays(baseConfiguration.ActivePlugins, rootConfiguration.ActivePlugins)

	}

	// Set the new config and assets
	a.Config = mergedConfiguration
	a.FS = FS
	a.Stack = assetStack

	// Validate the configuration
	return a.Config.Validate()
}

// Abs returns config-relative absolute paths
func (a *Assets) Abs(p string) string {
	fp := filepath.Join(a.FolderPath, p)
	fpabs, err := filepath.Abs(fp)
	if err != nil {
		return fp
	}
	return fpabs
}

// Abs returns config-relative absolute paths
func (a *Assets) DataAbs(p string) string {
	fp := filepath.Join(a.DataDir(), p)
	fpabs, err := filepath.Abs(fp)
	if err != nil {
		return fp
	}
	return fpabs
}

// DataDir returns the directory where data is stored
func (a *Assets) DataDir() string {
	return path.Join(a.FolderPath, "data")
}

// PluginDir returns the directory where plugin data is stored
func (a *Assets) PluginDir() string {
	return path.Join(a.FolderPath, "plugins")
}

// Open opens the assets in a given configuration path
func Open(configPath string, override *Configuration) (*Assets, error) {
	a := &Assets{
		FolderPath:     configPath,
		ConfigOverride: override,
	}

	return a, a.Reload()
}
