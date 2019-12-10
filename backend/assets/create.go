package assets

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
)

// Create takes a loaded configuration, as well as a target
// directory, and creates an associated heedy database.
// Can also optionally pass in the location of a configuration file
// which will be copied in, replacing the default config.
func Create(directory string, cfg *Configuration, configFile string) (*Assets, error) {

	osFs := afero.NewOsFs()
	f, err := osFs.Open(directory)
	if !os.IsNotExist(err) && err != nil {
		return nil, err
	}
	if err == nil {
		// There is a folder there already. Check if it is empty
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}
		if !fi.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", directory)
		}

		finfo, err := f.Readdir(1)
		if err != nil {
			return nil, err
		}
		if len(finfo) > 0 {
			return nil, fmt.Errorf("%s already has files in it. Must be empty to initialize heedy in it", directory)
		}
	}

	// Setting up the database: first we dump the newdb folder there
	builtinFs := BuiltinAssets()

	err = CopyDir(builtinFs, "/new", osFs, directory)
	if err != nil {
		osFs.RemoveAll(directory)
		return nil, err
	}

	configFilePath := path.Join(directory, "heedy.conf")

	// Now, if a config file was specified, overwrite the current one with it
	if configFile != "" {

		if err = osFs.Remove(configFilePath); err != nil {
			osFs.RemoveAll(directory)
			return nil, err
		}
		if err = CopyFile(osFs, configFile, osFs, configFilePath); err != nil {
			osFs.RemoveAll(directory)
			return nil, err
		}
	}

	// Finally, overwrite the config file with overloaded configuration options, which were specified
	// during setup
	if cfg != nil {
		err = WriteConfig(configFilePath, cfg)
		if err != nil {
			osFs.RemoveAll(directory)
			return nil, err
		}
	}

	// And now load the assets
	ass, err := Open(directory, nil)
	if err != nil {
		// Must clean up on failure
		osFs.RemoveAll(directory)
	}
	return ass, err
}
