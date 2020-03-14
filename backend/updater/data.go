package updater

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

func BackupData(configDir, updateDir, backupDir string) error {

	backupFileName := path.Join(backupDir, "data.zip")

	logrus.Infof("Creating data backup %s", backupFileName)
	backupFolderName := path.Join(configDir, "data")
	return ZipDirectory(backupFileName, backupFolderName)
}

func RevertData(configDir, backupDir, revertDir string) error {

	backupFileName := path.Join(backupDir, "data.zip")
	if _, err := os.Stat(backupFileName); os.IsNotExist(err) {
		return nil // Apparently no data backup exists.
	}

	logrus.Infof("Reverting data from backup %s", backupFileName)
	dataFolderName := path.Join(configDir, "data")
	revertDataFolder := path.Join(revertDir, "data")

	if err := os.Rename(dataFolderName, revertDataFolder); err != nil {
		return err
	}

	return UnzipDirectory(backupFileName, dataFolderName)

}
