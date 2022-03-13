package assets

import (
	"fmt"
	"os"
	"path"

	"strings"

	"github.com/sirupsen/logrus"
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

type CreateOptions struct {
	Directory  string         `json:"directory,omitempty"`
	Config     *Configuration `json:"config,omitempty"`
	ConfigFile string         `json:"config_file,omitempty"`
	Plugins    []string       `json:"plugins,omitempty"`
}

// Create takes a loaded configuration, as well as a target
// directory, and creates an associated heedy database.
// Can also optionally pass in the location of a configuration file
// which will be copied in, replacing the default config.
func Create(opt CreateOptions) (*Assets, error) {
	err := EnsureEmptyDatabaseFolder(opt.Directory)
	if err != nil {
		return nil, err
	}

	osFs := afero.NewOsFs()

	// Setting up the database: first we dump the newdb folder there
	builtinFs := BuiltinAssets()

	err = CopyDir(builtinFs, "/new", osFs, opt.Directory)
	if err != nil {
		osFs.RemoveAll(opt.Directory)
		return nil, err
	}

	configFilePath := path.Join(opt.Directory, "heedy.conf")

	if opt.ConfigFile != "" {
		if err = osFs.Remove(configFilePath); err != nil {
			osFs.RemoveAll(opt.Directory)
			return nil, err
		}
		if err = CopyFile(osFs, opt.ConfigFile, osFs, configFilePath); err != nil {
			osFs.RemoveAll(opt.Directory)
			return nil, err
		}
	}

	if len(opt.Plugins) > 0 {
		ppath := path.Join(opt.Directory, "plugins")
		if err = osFs.MkdirAll(ppath, os.ModePerm); err != nil {
			osFs.RemoveAll(opt.Directory)
			return nil, err
		}
		for _, p := range opt.Plugins {
			pname := path.Base(p)
			if strings.HasPrefix(p, "ln-s:") {
				p = p[5:]
				logrus.Debugf("Linking %s -> %s", p, path.Join(ppath, pname))
				err = os.Symlink(p, path.Join(ppath, pname))
			} else {
				logrus.Debugf("Copying %s -> %s", p, path.Join(ppath, pname))
				err = CopyDir(osFs, p, osFs, path.Join(ppath, pname))
			}
			if err != nil {
				osFs.RemoveAll(opt.Directory)
				return nil, err
			}
		}

	}

	// Finally, overwrite the config file with overloaded configuration options, which were specified
	// during setup
	if opt.Config != nil {
		err = WriteConfig(configFilePath, opt.Config)
		if err != nil {
			osFs.RemoveAll(opt.Directory)
			return nil, err
		}
	}

	// And now load the assets
	ass, err := Open(opt.Directory, nil)
	if err != nil {
		// Must clean up on failure
		osFs.RemoveAll(opt.Directory)
	}
	return ass, err
}
