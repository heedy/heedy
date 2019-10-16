package updater

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

func UpdateConfig(configDir, updateDir, backupDir string) error {
	configHeedy := path.Join(configDir, "heedy.conf")
	updateHeedy := path.Join(updateDir, "heedy.conf")
	backupHeedy := path.Join(backupDir, "heedy.conf")

	if _, err := os.Stat(updateHeedy); os.IsNotExist(err) {
		// The file does not exist, we're done here
		return nil
	}

	logrus.Info("Updating heedy.conf")
	return ShiftFiles(updateHeedy, configHeedy, backupHeedy)
}

func RevertConfig(configDir, backupDir, revertDir string) error {
	configHeedy := path.Join(configDir, "heedy.conf")
	revertHeedy := path.Join(revertDir, "heedy.conf")
	backupHeedy := path.Join(backupDir, "heedy.conf")

	if _, err := os.Stat(backupHeedy); os.IsNotExist(err) {
		// The file does not exist, we're done here
		return nil
	}

	logrus.Info("Reverting heedy.conf")

	return ShiftFiles(backupHeedy, configHeedy, revertHeedy)
}
