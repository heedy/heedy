package assets

import (
	"io"
	"path/filepath"

	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
)

type PrintOpen struct {
	afero.Fs
}

func (r PrintOpen) Open(n string) (afero.File, error) {
	log.Debugf("Request: %s", n)
	return r.Fs.Open(n)
}

// Copied from https://raw.githubusercontent.com/spf13/afero/355ac117537e70d4e4f4880db91457bd350c6b35/util.go

func CopyFile(srcFs afero.Fs, srcFilePath string, destFs afero.Fs, destFilePath string) error {
	// Some code from https://www.socketloop.com/tutorials/golang-copy-directory-including-sub-directories-files
	srcFile, err := srcFs.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := destFs.Create(destFilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	if err != nil {
		err = destFs.Chmod(destFilePath, srcInfo.Mode())
	}

	return nil
}

func CopyDir(srcFs afero.Fs, srcDirPath string, destFs afero.Fs, destDirPath string) error {
	// Some code from https://www.socketloop.com/tutorials/golang-copy-directory-including-sub-directories-files

	// get properties of source dir
	srcInfo, err := srcFs.Stat(srcDirPath)
	if err != nil {
		return err
	}

	// create dest dir
	if err = destFs.MkdirAll(destDirPath, srcInfo.Mode()); err != nil {
		return err
	}

	directory, err := srcFs.Open(srcDirPath)
	if err != nil {
		return err
	}
	defer directory.Close()

	entries, err := directory.Readdir(-1)

	for _, e := range entries {
		srcFullPath := filepath.Join(srcDirPath, e.Name())
		destFullPath := filepath.Join(destDirPath, e.Name())

		if e.IsDir() {
			// create sub-directories - recursively
			if err = CopyDir(srcFs, srcFullPath, destFs, destFullPath); err != nil {
				return err
			}
		} else {
			// perform copy
			if err = CopyFile(srcFs, srcFullPath, destFs, destFullPath); err != nil {
				return err
			}
		}
	}

	return nil
}
