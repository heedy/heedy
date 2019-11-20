package assets

import (
	"io"
	"path/filepath"
	"reflect"
	"strings"

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

	// get properties of object dir
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

// MergeStringArrays allows merging arrays of strings, with the result having each element
// at most once, and special prefix of + being ignored, and - allowing removal from array
func MergeStringArrays(base *[]string, overlay *[]string) *[]string {
	if base == nil {
		return overlay
	}
	if overlay == nil {
		return base
	}

	output := make([]string, 0)
	for _, d := range *base {
		if !strings.HasPrefix(d, "-") {
			if strings.HasPrefix(d, "+") {
				d = d[1:len(d)]
			}

			// Check if the output aready contains it
			contained := false
			for _, bd := range output {
				if bd == d {
					contained = true
					break
				}
			}
			if !contained {
				output = append(output, d)
			}

		}
	}
	for _, d := range *overlay {
		if strings.HasPrefix(d, "-") {
			if len(output) <= 0 {
				break
			}
			d = d[1:len(d)]

			// Remove element if contained
			for j, bd := range output {
				if bd == d {
					if len(output) == j+1 {
						output = output[:len(output)-1]
					} else {
						output[j] = output[len(output)-1]
						output = output[:len(output)-1]
						break
					}

				}
			}
		} else {
			if strings.HasPrefix(d, "+") {
				d = d[1:len(d)]
			}

			// Check if the output aready contains it
			contained := false
			for _, bd := range output {
				if bd == d {
					contained = true
					break
				}
			}
			if !contained {
				output = append(output, d)
			}
		}
	}
	return &output
}

func preprocess(i interface{}) (reflect.Value, reflect.Kind) {
	v := reflect.ValueOf(i)
	k := v.Kind()
	for k == reflect.Ptr {
		v = reflect.Indirect(v)
		k = v.Kind()
	}
	return v, k
}

// CopyStructIfPtrSet copies all pointer params from overlay to base
// Does not touch arrays and things that don't have identical types
func CopyStructIfPtrSet(base interface{}, overlay interface{}) {

	bv, _ := preprocess(base)
	ov, _ := preprocess(overlay)

	tot := ov.NumField()
	for i := 0; i < tot; i++ {
		// Now check if the field is of type ptr
		fieldValue := ov.Field(i)

		if fieldValue.Kind() == reflect.Ptr {
			// Only if it is a ptr do we continue, since that's all that we care about
			fieldName := ov.Type().Field(i).Name
			//fmt.Println(fieldName)

			baseFieldValue := bv.FieldByName(fieldName)
			if baseFieldValue.IsValid() && baseFieldValue.Type() == fieldValue.Type() {
				if !fieldValue.IsNil() {
					//fmt.Printf("Setting %s\n", fieldName)
					baseFieldValue.Set(fieldValue)
				}

			}

		}
	}

}
