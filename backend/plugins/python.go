package plugins

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/blang/semver"
)

// PythonEnv represents a python environment
type PythonEnv struct {
	PythonPath string
	Version    string
	PipPath    string
	CondaPath  string
}

// Paths to search for the executable
var PythonNames = []string{"python", "python2", "python3", "pypy3"}

// The libraries that are required
var RequiredLibs = []string{"numpy", "scipy", "matplotlib"}

var PythonVersionRequired = ">=3.7.0"
var PythonVersionRange = semver.MustParseRange(PythonVersionRequired)

// SearchPython finds the path
func SearchPython() ([]PythonEnv, error) {
	envs := make([]PythonEnv, 0, len(PythonNames))
	for i := range PythonNames {
		exepath, err := exec.LookPath(PythonNames[i])

		if err == nil {
			fmt.Println(exepath)
			err = TestPython(exepath)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				fmt.Printf("Is ok!\n")
			}
		} else {
			fmt.Printf("Python %s not found\n", PythonNames[i])
		}
	}
	return envs, nil
}

// TestPython checks if the given python version satisfies all requirements
func TestPython(exepath string) error {

	// A version of Python was found. Let's check what its actual version is
	versionString, err := exec.Command(exepath, "--version").Output()
	if err == nil {
		vlist := strings.Split(string(versionString), " ")
		if len(vlist) == 2 {
			version, err := semver.Parse(strings.Trim(vlist[1], "\n "))
			if err == nil && PythonVersionRange(version) {
				fmt.Printf("-> %s\n", versionString)
				// Now check that the required packages are installed
				for i := range RequiredLibs {
					libversion, err := exec.Command(exepath, "-c", "import "+RequiredLibs[i]+";print("+RequiredLibs[i]+".__version__)").Output()
					if err != nil {
						return err
					}
					fmt.Printf("%s: %s", RequiredLibs[i], string(libversion))
				}

			} else {
				return fmt.Errorf("%s incompatible with requirement %s (%s)\n", string(versionString), PythonVersionRequired, strings.Trim(vlist[1], "\n "))

			}
		} else {
			return fmt.Errorf("%s: Couldn't read version (output: '%s')", exepath, string(versionString))
		}
	}
	return err
}
