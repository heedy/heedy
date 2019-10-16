package updater

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func Update(configDir string) (bool, error) {
	configDir, err := filepath.Abs(configDir)
	if err != nil {
		return false, err
	}
	updateDir := path.Join(configDir, "updates")
	if _, err := os.Stat(updateDir); os.IsNotExist(err) {
		// No updates are available
		return false, nil
	}

	backupDir := path.Join(configDir, "backup")
	if _, err := os.Stat(backupDir); !os.IsNotExist(err) {
		// The backup directory exists. Move it over to backup.old
		backupOld := path.Join(configDir, "backup.old")

		if _, err = os.Stat(backupOld); !os.IsNotExist(err) {
			if err = os.RemoveAll(backupOld); err != nil {
				return true, err
			}
		}

		if err = os.Rename(backupDir, backupOld); err != nil {
			return true, err
		}
	}

	if err = os.MkdirAll(backupDir, os.ModePerm); err != nil {
		return true, err
	}

	if err = BackupData(configDir, updateDir, backupDir); err != nil {
		return true, err
	}
	if err = UpdatePlugins(configDir, updateDir, backupDir); err != nil {
		return true, err
	}
	if err = UpdateHeedy(configDir, updateDir, backupDir); err != nil {
		return true, err
	}
	if err = UpdateConfig(configDir, updateDir, backupDir); err != nil {
		return true, err
	}

	// Remove the revert directory, to avoid confusion: the update will be successful
	revertDir := path.Join(configDir, "updates.reverted")
	if _, err := os.Stat(revertDir); !os.IsNotExist(err) {
		if err = os.RemoveAll(revertDir); err != nil {
			return true, err
		}
	}

	return true, os.RemoveAll(updateDir)

}

func Revert(configDir string, failure error) error {
	logrus.Warn("Reverting update")
	configDir, err := filepath.Abs(configDir)
	if err != nil {
		return err
	}

	backupDir := path.Join(configDir, "backup")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("Couldn't revert: no backup available (%s)", backupDir)
	}

	// Create the directory where reverted stuff will be stored (while deleting any old reverts)
	revertDir := path.Join(configDir, "updates.reverted")
	if _, err := os.Stat(revertDir); !os.IsNotExist(err) {
		if err = os.RemoveAll(revertDir); err != nil {
			return err
		}
	}
	if err = os.MkdirAll(revertDir, os.ModePerm); err != nil {
		return err
	}

	if err = ioutil.WriteFile(path.Join(revertDir, "ERROR"), []byte(failure.Error()), os.ModePerm); err != nil {
		return err
	}

	// Start by reverting the data directory to the backed-up version
	if err = RevertData(configDir, backupDir, revertDir); err != nil {
		return err
	}

	// Next revert plugins to their backed-up versions
	if err = RevertPlugins(configDir, backupDir, revertDir); err != nil {
		return err
	}

	// Finally, revert the heedy executable if possible
	if err = RevertHeedy(configDir, backupDir, revertDir); err != nil {
		return err
	}

	// Revert the heedy.conf file
	err = RevertConfig(configDir, backupDir, revertDir)
	if err == nil {
		return os.RemoveAll(backupDir)
	}
	return err
}
