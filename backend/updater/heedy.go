package updater

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// UpdateHeedy updates the heedy executable
func UpdateHeedy(configDir, updateDir, backupDir string) error {
	configHeedy := path.Join(configDir, "heedy")
	updateHeedy := path.Join(updateDir, "heedy")
	backupHeedy := path.Join(backupDir, "heedy")

	if _, err := os.Stat(updateHeedy); os.IsNotExist(err) {
		// The file does not exist, we're done here
		return nil
	}

	logrus.Info("Updating heedy executable")

	/* Signatures not supported yet
	// Move the signature file over
	err := ShiftFiles(updateHeedy+".sig", configHeedy+".sig", backupHeedy+".sig")
	if err != nil {
		return err
	}
	*/

	return ShiftFiles(updateHeedy, configHeedy, backupHeedy)
}

// RevertHeedy reverts the heedy executable to the backed up version if such exists
func RevertHeedy(configDir, backupDir, revertDir string) error {
	configHeedy := path.Join(configDir, "heedy")
	revertHeedy := path.Join(revertDir, "heedy")
	backupHeedy := path.Join(backupDir, "heedy")

	if _, err := os.Stat(backupHeedy); os.IsNotExist(err) {
		// The file does not exist, we're done here
		return nil
	}

	logrus.Info("Reverting heedy executable")

	// Move the signature file over
	err := ShiftFiles(backupHeedy+".sig", configHeedy+".sig", revertHeedy+".sig")
	if err != nil {
		return err
	}
	return ShiftFiles(backupHeedy, configHeedy, revertHeedy)
}
