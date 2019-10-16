package updater

import (
	"errors"
	"os"
	"path"

	"io/ioutil"

	"github.com/heedy/heedy/backend/assets"
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
