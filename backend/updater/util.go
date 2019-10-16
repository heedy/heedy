package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// ShiftFiles moves middle->out (if exists), and then in->middle
func ShiftFiles(in, middle, out string) error {
	if _, err := os.Stat(middle); !os.IsNotExist(err) {
		logrus.Debugf("Moving %s -> %s", middle, out)
		if err = os.Rename(middle, out); err != nil {
			return err
		}
	}
	if _, err := os.Stat(in); !os.IsNotExist(err) {
		logrus.Debugf("Moving %s -> %s", in, middle)
		return os.Rename(in, middle)
	}

	return nil
}

func OverwriteFile(in, out string) error {
	logrus.Debugf("Overwriting %s -> %s", in, out)
	of, err := os.OpenFile(out, os.O_RDWR, 0766)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// The file doesn't exist - move it instead
		return os.Rename(in, out)
	}
	defer of.Close()
	of.Seek(0, 0)
	of.Truncate(0)

	inf, err := os.Open(in)
	if err != nil {
		return err
	}

	_, err = io.Copy(of, inf)
	inf.Close()
	if err == nil {
		return os.Remove(in)
	}
	return err
}

func CopyFile(in, out string) error {
	logrus.Debugf("Copying %s -> %s", in, out)

	err := os.MkdirAll(path.Dir(out), os.ModeDir)
	if err != nil {
		return err
	}

	inf, err := os.Open(in)
	if err != nil {
		return err
	}
	defer inf.Close()

	of, err := os.Create(out)
	if err != nil {
		return err
	}
	defer of.Close()

	_, err = io.Copy(of, inf)
	return err
}

func zipDirectory(w *zip.Writer, fpath, zipPath string) error {
	d, err := ioutil.ReadDir(fpath)
	if err != nil {
		return err
	}

	for _, f := range d {
		zipName := path.Join(zipPath, f.Name())
		fullName := path.Join(fpath, f.Name())
		if f.IsDir() {
			if err = zipDirectory(w, fullName, zipName); err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(fullName, ".sock") {
				logrus.Debugf("skipping socket %s", fullName)
				continue
			}
			logrus.Debugf("Zipping %s", fullName)
			fileToZip, err := os.Open(fullName)
			if err != nil {

				return err
			}
			defer fileToZip.Close()
			finfo, err := fileToZip.Stat()
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(finfo)
			if err != nil {
				return err
			}
			header.Name = zipName
			header.Method = zip.Deflate

			fwriter, err := w.CreateHeader(header)
			if err != nil {
				return err
			}

			_, err = io.Copy(fwriter, fileToZip)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ZipDirectory(zipFile, inputDir string) error {
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return err
	}

	zf, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer zf.Close()

	zipWriter := zip.NewWriter(zf)
	defer zipWriter.Close()

	return zipDirectory(zipWriter, inputDir, "/")
}

// UnzipDirectory will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// https://golangcode.com/unzip-files-in-go/
func UnzipDirectory(src string, dest string) error {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		logrus.Debugf("Extracting %s", fpath)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
