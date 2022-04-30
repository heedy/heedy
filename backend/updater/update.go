package updater

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

var ErrBackup = errors.New("no backup available")

func PrepareBackupFolder(configDir string) (backupDir string, err error) {
	backupDir = path.Join(configDir, "backup")
	_, err = os.Stat(backupDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(backupDir, os.ModePerm)
	}
	if err != nil {
		return
	}
	backupDir = path.Join(backupDir, time.Now().Format("2006-01-02-15-04-05"))
	err = os.MkdirAll(backupDir, os.ModePerm)
	return
}

func RemoveOldBackups(configDir string) error {
	backupDir := path.Join(configDir, "backup")
	_, err := os.Stat(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Now check how many backups we have
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return err
	}

	// And get the maximum number configured.
	maxBackups := 1
	cfg := assets.Config()
	if cfg.MaxBackupCount != nil {
		maxBackups = *cfg.MaxBackupCount
	}
	cfg, err = assets.LoadConfigFile(configDir + "/heedy.conf")
	if err == nil && cfg.MaxBackupCount != nil {
		maxBackups = *cfg.MaxBackupCount
	}

	// Negative numbers/0 mean no limit
	if len(files) > maxBackups && maxBackups > 0 {
		// Delete the oldest backup
		files = files[:len(files)-maxBackups]
		for _, f := range files {
			err = os.RemoveAll(path.Join(backupDir, f.Name()))
		}

	}
	return err
}

func GetLatestBackup(configDir string) (backupDir string, err error) {
	backupDir = path.Join(configDir, "backup")
	_, err = os.Stat(backupDir)
	if err != nil {
		return
	}
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return "", ErrBackup
	}
	if len(files) == 0 {
		return "", ErrBackup
	}
	backupDir = path.Join(backupDir, files[len(files)-1].Name())
	return
}

type UpdateOptions struct {
	BackupData     bool     `json:"backup"`
	DeletedPlugins []string `json:"deleted"`
}

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

	// Read in update_options.json to get the settings for updates
	var o *UpdateOptions = nil
	if updateOptionsFile, err := ioutil.ReadFile(path.Join(updateDir, "update_options.json")); err == nil {
		err = json.Unmarshal(updateOptionsFile, &o)
		if err != nil {
			return true, err
		}
	}

	backupDir, err := PrepareBackupFolder(configDir)
	if err != nil {
		return true, err
	}
	if o != nil && o.BackupData {
		if err = BackupData(configDir, updateDir, backupDir); err != nil {
			return true, err
		}
	}
	if o != nil {
		if err = RemovePlugins(configDir, updateDir, backupDir, o.DeletedPlugins); err != nil {
			return true, err
		}
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

	files, err := ioutil.ReadDir(backupDir)
	if err == nil && len(files) == 0 {
		// Nothing was actually backed up... so remove the directory
		os.RemoveAll(backupDir)

		return false, os.RemoveAll(updateDir)
	}

	// Remove the revert directory, to avoid confusion: the update will be successful
	revertDir := path.Join(configDir, "updates.reverted")
	if _, err := os.Stat(revertDir); !os.IsNotExist(err) {
		if err = os.RemoveAll(revertDir); err != nil {
			return true, err
		}
	}

	RemoveOldBackups(configDir)

	return true, os.RemoveAll(updateDir)

}

func Revert(configDir string, failure error) error {
	logrus.Warn("Reverting from backup")
	configDir, err := filepath.Abs(configDir)
	if err != nil {
		return err
	}

	backupDir, err := GetLatestBackup(configDir)
	if err != nil {
		return err
	}

	// Create the directory where reverted stuff will be stored (while deleting any old reverts)
	revertDir := path.Join(configDir, "updates.reverted")
	if _, err = os.Stat(revertDir); !os.IsNotExist(err) {
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
