package python

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PythonCandidate struct {
	Path    string
	PipArgs []string
}

const testScript = `
import sys
try:
	if sys.version_info.major != 3 or sys.version_info.minor < 7:
		raise Exception("Heedy's Python support requires at least Python 3.7")

	import venv
	import ensurepip

	print("OK")
except Exception as e:
	print(e)
	sys.exit(1)
`

// Paths to search for the executable. These are commonly used names when python is in PATH
var PathNames = []string{"python", "python3", "pypy3"}

// SearchPython finds a valid installed python version
func SearchPython() (string, error) {
	logrus.Debug("Searching for compatible Python interpreter")
	for i := range PathNames {
		exepath, err := exec.LookPath(PathNames[i])
		if err == nil {
			err = ValidatePython(exepath)
			if err == nil {
				return filepath.Abs(exepath)
			}
			logrus.Debug(err)
		}
	}
	return "", errors.New("No supported Python found")
}

// ValidatePython checks if the given python version satisfies all requirements
func ValidatePython(exepath string) error {
	if settings.DB != nil && settings.DB.Verbose {
		logrus.Debugf("Checking python at %s with script: %s", exepath, testScript)
	} else {
		logrus.Debugf("Checking python at %s", exepath)
	}
	testResult, err := exec.Command(exepath, "-c", testScript).CombinedOutput()
	if err != nil {
		if len(testResult) > 0 {
			return fmt.Errorf("Python at %s not supported: %s, (%w)", exepath, string(testResult), err)
		}
		return err
	}
	cmdout := strings.TrimSpace(string(testResult))
	if cmdout != "OK" {
		return fmt.Errorf("Python at %s not supported: %s", exepath, cmdout)
	}
	return nil
}

func RunCommand(pypath string, args []string) error {
	if settings.DB.Verbose {
		l.Debugf("%s %s", pypath, strings.Join(args, " "))
	}
	cmd := exec.Command(pypath, args...)
	lvl := logrus.GetLevel()
	if lvl == logrus.DebugLevel || lvl == logrus.InfoLevel {
		logdir := settings.DB.Assets().LogDir()
		if logdir == "stdout" {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			logfile := path.Join(logdir, "python.log")
			f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return err
			}
			_, err = f.WriteString(fmt.Sprintf("\n\n%s >>> %s %s\n", time.Now().Format(time.RFC3339), pypath, strings.Join(args, " ")))
			if err != nil {
				return err
			}
			cmd.Stdout = f
			cmd.Stderr = f
			defer f.Close()
		}
	}
	return cmd.Run()
}

// EnsureVenv makes sure that a venv exists in the given folder. If not, it creates one there.
// Returns the path of the venv's python executable.
func EnsureVenv(pypath, folder string) (string, error) {
	newpypath := path.Join(folder, "bin", "python")
	if runtime.GOOS == "windows" {
		newpypath = path.Join(folder, "Scripts", "python.exe")
	}
	if _, err := os.Stat(newpypath); !os.IsNotExist(err) {

		// If the venv exists, check if it is compatible
		if err = ValidatePython(newpypath); err == nil {
			return newpypath, err
		}

		l.Warnf("Found existing venv at %s, but it failed to initialize. Removing it and creating a new one.", newpypath)

		// If it is not compatible, delete it, and attempt to recreate it
		if err = os.RemoveAll(folder); err != nil {
			return "", err
		}
	}

	// No venv exists. Let's create it.

	// Create the holding folder if not exist
	if err := os.MkdirAll(path.Dir(folder), os.ModePerm); err != nil {
		return pypath, err
	}

	// Create the venv!
	err := RunCommand(pypath, append([]string{"-m", "venv", folder}, settings.VenvArgs...))

	if err != nil {
		// If there was an error creating the venv, clear it!
		os.RemoveAll(folder)
	}

	return newpypath, err
}
