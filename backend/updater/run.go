package updater

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"
	"github.com/sirupsen/logrus"
)

type Options struct {
	ConfigDir   string
	AddonConfig *assets.Configuration
	RunOptions  *server.RunOptions
	Revert      bool
}

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

func Run(o Options) error {
	hadUpdate, err := Update(o.ConfigDir)
	if err != nil {
		return err
	}

	// Check if the config directory contains a heedy executable

	heedyPath, err := filepath.Abs(path.Join(o.ConfigDir, "heedy"))
	if err != nil {
		return err
	}
	_, err = os.Stat(heedyPath)
	restartHeedy := !os.IsNotExist(err)

	curPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}

	if restartHeedy && (curPath != heedyPath || hadUpdate) {
		// TODO: check the signature
		// We run the heedy executable.
		a := []string{}
		if hadUpdate {
			a = append(a, "--revert")
		}
		a = append(a, os.Args[1:]...)
		return StartProcess(heedyPath, a...)

	}

	// Actually run it
	a, err := assets.Open(o.ConfigDir, o.AddonConfig)
	if err == nil {
		assets.SetGlobal(a)
		err = server.Run(a, o.RunOptions)
	}

	if o.Revert && err != nil {
		logrus.Error(err)
		err = Revert(o.ConfigDir)
		if err != nil {
			return err
		}

		_, err = os.Stat(heedyPath)
		restartHeedy = !os.IsNotExist(err)
		if restartHeedy {
			return StartProcess(heedyPath, os.Args[1:]...)
		}

	}

	return err
}
