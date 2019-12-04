package python

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
func SearchPython() (string, []string, error) {
	logrus.Debug("Searching for compatible Python interpreter")
	for i := range PathNames {
		exepath, err := exec.LookPath(PathNames[i])

		if err == nil {
			err = TestPython(exepath)
			if err != nil {
				logrus.Debug(err)
			} else {
				// The path was valid! Now check if we should add the --user arg, which
				// needs to be added if python is not in the user directory
				apath, err := filepath.Abs(exepath)
				if err != nil {
					return "", []string{}, err
				}
				d, err := os.UserHomeDir()
				if err != nil {
					return "", []string{}, err
				}
				if strings.HasPrefix(apath, d) || runtime.GOOS == "windows" {
					return exepath, []string{"--quiet"}, nil
				}

				return exepath, []string{"--quiet", "--user"}, nil
			}
		}
	}
	return "", []string{}, errors.New("No supported Python found")
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
