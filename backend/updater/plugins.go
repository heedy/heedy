package updater

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// UpdatePlugins updates all the plugins
func UpdatePlugins(configDir, updateDir, backupDir string) error {
	configPluginDir := path.Join(configDir, "plugins")
	updatePluginDir := path.Join(updateDir, "plugins")
	backupPluginDir := path.Join(backupDir, "plugins")

	pinfo, err := ioutil.ReadDir(updatePluginDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = ioutil.ReadDir(configPluginDir)
	if os.IsNotExist(err) {
		// The plugins folder does not exist yet. Create it.
		if err = os.MkdirAll(configPluginDir, os.ModePerm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// Create the backup plugin directory
	if err = os.MkdirAll(backupPluginDir, os.ModePerm); err != nil {
		return err
	}
	for _, pfolder := range pinfo {
		pname := pfolder.Name()
		logrus.Infof("Updating plugin %s", pname)
		updatedPluginLocation := path.Join(updatePluginDir, pname)
		currentPluginLocation := path.Join(configPluginDir, pname)
		backupPluginLocation := path.Join(backupPluginDir, pname)

		if err = ShiftFiles(updatedPluginLocation, currentPluginLocation, backupPluginLocation); err != nil {
			return err
		}
	}

	return nil
}

func RevertPlugins(configDir, backupDir, revertDir string) error {
	configPluginDir := path.Join(configDir, "plugins")
	revertPluginDir := path.Join(revertDir, "plugins")
	backupPluginDir := path.Join(backupDir, "plugins")

	pinfo, err := ioutil.ReadDir(backupPluginDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// Create the backup plugin directory
	if err = os.MkdirAll(revertPluginDir, os.ModePerm); err != nil {
		return err
	}

	for _, pfolder := range pinfo {
		pname := pfolder.Name()
		logrus.Infof("Reverting update to plugin %s", pname)
		revertPluginLocation := path.Join(revertPluginDir, pname)
		currentPluginLocation := path.Join(configPluginDir, pname)
		backupPluginLocation := path.Join(backupPluginDir, pname)

		if err != ShiftFiles(backupPluginLocation, currentPluginLocation, revertPluginLocation) {
			return err
		}
	}
	return nil
}
