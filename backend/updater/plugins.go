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

func RemovePlugins(configDir, updateDir, backupDir string, plugins []string) error {
	configPluginDir := path.Join(configDir, "plugins")
	updatePluginDir := path.Join(updateDir, "plugins")
	backupPluginDir := path.Join(backupDir, "plugins")

	// If the plugin is in the update directory, remove it from the directory

	if _, err := ioutil.ReadDir(updatePluginDir); err == nil {
		for _, p := range plugins {
			err = os.RemoveAll(path.Join(updatePluginDir, p))
			if err != nil {
				return err
			}
		}
	}
	// If the plugin is in the config directory, move it to backup
	if _, err := ioutil.ReadDir(configPluginDir); err == nil {

		// Make sure the backup plugin directory exists
		_, err = ioutil.ReadDir(backupPluginDir)
		if os.IsNotExist(err) {
			// The plugins folder does not exist yet. Create it.
			if err = os.MkdirAll(backupPluginDir, os.ModePerm); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		for _, p := range plugins {
			pdir := path.Join(configPluginDir, p)
			bdir := path.Join(backupPluginDir, p)
			if _, err = os.Stat(pdir); err == nil {
				logrus.Debugf("Moving %s -> %s", pdir, bdir)
				err = os.Rename(pdir, bdir)
				if err != nil {
					return err
				}
			}
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
