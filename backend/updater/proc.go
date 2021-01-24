package updater

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

// StartProcess starts the given process in the background, and releases it, so that
// closing the current process doesn't close the child.
func StartProcess(heedyPath string, args ...string) error {
	logrus.Debugf("Starting Process: %s %s", heedyPath, strings.Join(args, " "))
	cmd := exec.Command(heedyPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	// Release the process, let it be freeeeee
	return cmd.Process.Release()
}

//  ReplaceProcess replaces the current process with the one given. It only works on unix
// so it will fail with an error on Windows.
func ReplaceProcess(heedyPath string, args ...string) error {
	logrus.Debugf("Replacing Process: %s %s", heedyPath, strings.Join(args, " "))

	argv := []string{filepath.Base(heedyPath)}
	argv = append(argv, args...)
	return syscall.Exec(heedyPath, argv, os.Environ())
}

func ReplaceOrStart(heedyPath string, args ...string) error {
	err := ReplaceProcess(heedyPath, args...)
	// ReplaceProcess does not return on success, so no need for conditionals.

	logrus.Debugf("Failed to replace process %s, starting in background instead: %s", heedyPath, err.Error())
	return StartProcess(heedyPath, args...)
}

// StartHeedy starts the heedy server set up for the given database. If replace is true,
// it tries to replace the current process with the new one, if not, or if replacement is not
// supported, it starts the new process in the background.
func StartHeedy(configDir string, replace bool, extraArgs ...string) error {
	heedyPath, err := filepath.Abs(path.Join(configDir, "heedy"))
	if err != nil {
		return err
	}

	if _, err = os.Stat(heedyPath); os.IsNotExist(err) {
		// We use the current executable
		heedyPath, err = os.Executable()
	}
	if err != nil {
		return err
	}

	// Set up the args
	args := make([]string, 0, len(extraArgs)+len(os.Args)-1)
	args = append(args, os.Args[1:]...)

	// Only add the extraArgs that are not yet added
	for _, ea := range extraArgs {
		hadArg := false
		for _, a := range args {
			if a == ea {
				hadArg = true
				break
			}
		}
		if !hadArg {
			args = append(args, ea)
		}
	}

	// Now replace whatever command was used with the run command
	for i := range args {
		switch args[i] {
		case "run":
			break
		case "start":
			args[i] = "run"
			break
		case "create":
			// TODO: fix flags, since create might have different flags!
			args[i] = "run"
			break
		}
	}
	if replace {
		return ReplaceOrStart(heedyPath, args...)
	}
	return StartProcess(heedyPath, args...)
}
