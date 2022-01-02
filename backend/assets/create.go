package assets

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
)

func EnsureEmptyDatabaseFolder(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(files) > 0 {
		return fmt.Errorf("%s is not empty", dir)
	}
	return nil
}

// Create takes a loaded configuration, as well as a target
// directory, and creates an associated heedy database.
// Can also optionally pass in the location of a configuration file
// which will be copied in, replacing the default config.
func Create(directory string, cfg *Configuration, configFile string) (*Assets, error) {
	err := EnsureEmptyDatabaseFolder(directory)
	if err != nil {
		return nil, err
	}

	osFs := afero.NewOsFs()

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
