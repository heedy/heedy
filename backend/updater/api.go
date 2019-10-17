package updater

import (
	"errors"
	"os"
	"path"

	"io/ioutil"

	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

type UpdateInfo struct {
	Heedy   bool     `json:"heedy"`
	Config  bool     `json:"config"`
	Plugins []string `json:"plugins"`
}

func GetInfo(configDir string) (ui UpdateInfo) {
	var err error
	updateDir := path.Join(configDir, "updates")
	_, err = os.Stat(path.Join(updateDir, "heedy.conf"))
	ui.Config = err == nil
	_, err = os.Stat(path.Join(updateDir, "heedy"))
	ui.Heedy = err == nil

	d, err := ioutil.ReadDir(path.Join(updateDir, "plugins"))
	if err != nil {
		ui.Plugins = make([]string, 0)
		return
	}
	s := make([]string, len(d))
	for i := range d {
		s[i] = d[i].Name()
	}
	ui.Plugins = s
	return
}

func Available(configDir string) bool {
	updateDir := path.Join(configDir, "updates")
	_, err := os.Stat(updateDir)
	return err == nil
}

func ReadConfigFile(configDir string) ([]byte, error) {
	b, err := ioutil.ReadFile(path.Join(configDir, "updates", "heedy.conf"))
	if err == nil {
		return b, err
	}
	return ioutil.ReadFile(path.Join(configDir, "heedy.conf"))
}

func ModifyConfigFile(configDir string, c *assets.Configuration) error {
	configHeedy := path.Join(configDir, "heedy.conf")
	updateHeedy := path.Join(configDir, "updates", "heedy.conf")

	if _, err := os.Stat(updateHeedy); os.IsNotExist(err) {
		// The file does not exist, copy over heedy.conf
		err = CopyFile(configHeedy, updateHeedy)
		if err != nil {
			return err
		}
	}

	return assets.WriteConfig(updateHeedy, c)
}

func SetConfigFile(configDir string, b []byte) error {
	updateHeedy := path.Join(configDir, "updates", "heedy.conf")

	_, err := assets.LoadConfigBytes(b, "updates/heedy.conf")
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(updateHeedy), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(updateHeedy, b, os.ModePerm)
}

func ReadConfig(a *assets.Assets) (*assets.Configuration, error) {
	configHeedy := path.Join(a.FolderPath, "heedy.conf")
	updateHeedy := path.Join(a.FolderPath, "updates", "heedy.conf")
	if _, err := os.Stat(updateHeedy); os.IsNotExist(err) {
		return assets.LoadConfigFile(configHeedy)
	}

	return assets.LoadConfigFile(updateHeedy)
}

func Status(configDir string) error {
	errFile := path.Join(configDir, "updates.reverted", "ERROR")
	if _, err := os.Stat(errFile); os.IsNotExist(err) {
		return nil
	}

	b, err := ioutil.ReadFile(errFile)
	if err != nil {
		return err
	}
	return errors.New(string(b))
}

// Lists ALL plugins (including those that are not active, and those that are currently pending restart)
func ListPlugins(configDir string) (map[string]*assets.Plugin, error) {
	pluginDir := path.Join(configDir, "plugins")
	pluginUpdateDir := path.Join(configDir, "updates", "plugins")

	p := make(map[string]*assets.Plugin)

	d, err := ioutil.ReadDir(pluginDir)
	if err == nil {
		for _, v := range d {
			pFile := path.Join(pluginDir, v.Name(), "heedy.conf")
			c, err := assets.LoadConfigFile(pFile)
			if err == nil {
				pv, ok := c.Plugins[v.Name()]
				if ok {
					p[v.Name()] = pv
				}
			}
		}
	}
	d, err = ioutil.ReadDir(pluginUpdateDir)
	if err == nil {
		for _, v := range d {
			pFile := path.Join(pluginUpdateDir, v.Name(), "heedy.conf")
			c, err := assets.LoadConfigFile(pFile)
			if err == nil {
				pv, ok := c.Plugins[v.Name()]
				if ok {
					p[v.Name()] = pv
				}
			}
		}
	}
	return p, nil
}

func UpdatePlugin(configDir string, zipFile string) error {
	// Extract the file into a temporary directory
	tmpDir, err := ioutil.TempDir(configDir, "tmp-plugin-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	logrus.Debugf("Unzipping %s -> %s", zipFile, tmpDir)
	if err = UnzipDirectory(zipFile, tmpDir); err != nil {
		return err
	}

	// The plugin is unzipped. Check the folder
	d, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		return err
	}
	if len(d) == 0 {
		return errors.New("Empty zip file")
	}
	if len(d) > 1 {
		// WHY ON EARTH does mac have to include garbage in its zip files
		if len(d) > 2 || d[0].Name() != "__MACOSX" && d[1].Name() != "__MACOSX" {
			return errors.New("Only a single plugin folder per zip file is supported")
		}
		if d[0].Name() == "__MACOSX" {
			d[0] = d[1]
		}
	}

	if !d[0].IsDir() {
		// HACK: update the main heedy executable
		if d[0].Name() != "heedy" && d[0].Name() != "heedy.exe" {
			return errors.New("The plugin must be in a folder")
		}

		if err = os.MkdirAll(path.Join(configDir, "updates"), os.ModePerm); err != nil {
			return err
		}
		outName := path.Join(configDir, "updates", "heedy")
		if _, err := os.Stat(outName); !os.IsNotExist(err) {
			logrus.Debugf("Removing %s", outName)
			if err = os.Remove(outName); err != nil {
				return err
			}
		}
		tmpName := path.Join(tmpDir, "heedy")

		// Make it executable
		f, err := os.Open(tmpName)
		if err != nil {
			return err
		}
		if err = f.Chmod(0554); err != nil {
			return err
		}

		logrus.Debugf("Moving %s -> %s", tmpName, outName)
		return os.Rename(tmpName, outName)

	}
	pn := d[0].Name()
	pfile := path.Join(tmpDir, pn, "heedy.conf")
	c, err := assets.LoadConfigFile(pfile)
	if err != nil {
		return err
	}
	if _, ok := c.Plugins[d[0].Name()]; !ok {
		return errors.New("The plugin folder and name must match")
	}

	// OK, looks like the plugin passed sanity checks. Let's copy it over to the updates folder
	if err = os.MkdirAll(path.Join(configDir, "updates", "plugins"), os.ModePerm); err != nil {
		return err
	}
	outFolder := path.Join(configDir, "updates", "plugins", d[0].Name())
	if _, err := os.Stat(outFolder); !os.IsNotExist(err) {
		logrus.Debugf("Removing %s", outFolder)
		if err = os.RemoveAll(outFolder); err != nil {
			return err
		}
	}
	tmpFolder := path.Join(tmpDir, pn)
	logrus.Debugf("Moving %s -> %s", tmpFolder, outFolder)
	return os.Rename(tmpFolder, outFolder)
}
