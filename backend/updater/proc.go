package updater

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

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

func StartHeedy(configDir string, extraArgs ...string) error {
	heedyPath, err := filepath.Abs(path.Join(configDir, "heedy"))
	if err != nil {
		return err
	}

	if _, err = os.Stat(heedyPath); os.IsNotExist(err) {
		// We use the current executable
		heedyPath, err = filepath.Abs(os.Args[0])
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

	return StartProcess(heedyPath, args...)
}
