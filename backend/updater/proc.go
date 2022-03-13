package updater

import (
	"os"
	"os/exec"
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
	// We use the current executable
	heedyPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Set up the args
	args := make([]string, 0, len(extraArgs)+len(os.Args)-1)
	args = append(args, os.Args[1:]...)

	// Only add the extraArgs that are not yet added - this is hacky,
	// but works for boolean flags (like --update)
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
	// TODO: This is a nasty hack, but
outerloop:
	for i := range args {
		switch args[i] {
		case "run":
			break outerloop
		case "start":
			args[i] = "run"
			break outerloop
		case "create":

			args[i] = "run"

			// TODO: This is a hacky and crappy solution to the problem of rewriting the args,
			// but it is not a priority to make it robust right now...

			// Key is arg to replace, value is whether it has arg value itself
			argmap := map[string]bool{
				"--noserver": false,
				"--username": true,
				"--password": true,
				"--testapp":  true,
				"--config":   true,
				"--plugin":   true,
			}

			// Fix the flags to remove flags used when creating
			for j := i + 1; j < len(args); j++ {
				if strings.HasPrefix(args[j], "-") {
					// Now args can be of form --flag="value", split
					fv := strings.SplitN(args[j], "=", 2)
					val, ok := argmap[fv[0]]
					if ok {
						// Remove the arg, and its value if it has one
						args = append(args[:j], args[j+1:]...)
						if len(fv) == 1 && val {
							args = append(args[:j], args[j+1:]...)
							j--
						}
						j--

					}
				}
			}
			break outerloop
		}
	}
	if replace {
		return ReplaceOrStart(heedyPath, args...)
	}
	return StartProcess(heedyPath, args...)
}
