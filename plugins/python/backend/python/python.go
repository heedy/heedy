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

	"github.com/sirupsen/logrus"
)

type PythonCandidate struct {
	Path    string
	PipArgs []string
}

// Paths to search for the executable. These are commonly used names when python is in PATH
var PathNames = []string{"python", "python3", "pypy3"}

// SearchPython finds a valid installed python version
func SearchPython() (string, error) {
	logrus.Debug("Searching for compatible Python interpreter")
	for i := range PathNames {
		exepath, err := exec.LookPath(PathNames[i])
		if err == nil {
			err = TestPython(exepath)
			if err == nil {
				return filepath.Abs(exepath)
			}
			logrus.Debug(err)
		}
	}
	return "", errors.New("No supported Python found")
}

// TestPython checks if the given python version satisfies all requirements
func TestPython(exepath string) error {
	logrus.Debugf("Checking python at %s", exepath)
	var testScript = `
import sys
try:
	import pkg_resources

	if sys.version_info.major < 3 or sys.version_info.minor < 7:
		raise Exception("Heedy's Python support requires at least Python 3.7")

	requirements = [
		"pip"
	]
	pkg_resources.require(requirements)
	print("OK")
except Exception as e:
	print(e)
	sys.exit(1)
`
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
		return newpypath, err
	}

	// No venv exists. Let's create it.

	// Create the holding folder if not exist
	if err := os.MkdirAll(path.Dir(folder), os.ModePerm); err != nil {
		return pypath, err
	}

	// Create the venv!
	err := RunCommand(pypath, append([]string{"-m", "venv", folder}, settings.VenvArgs...))

	return newpypath, err
}
