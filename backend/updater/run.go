package updater

import (
	"os"
	"path"
	"path/filepath"

	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

type Options struct {
	ConfigDir   string
	AddonConfig *assets.Configuration
	Runner      func(a *assets.Assets) error
	Revert      bool
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
	if hadUpdate {
		o.Revert = true
	}

	// Actually run it
	a, err := assets.Open(o.ConfigDir, o.AddonConfig)
	if err == nil {
		assets.SetGlobal(a)
		err = o.Runner(a)
	}

	if o.Revert && err != nil {
		logrus.Error(err)
		err = Revert(o.ConfigDir, err)
		if err != nil {
			return err
		}

		return RunHeedy(o.ConfigDir)
	}

	return err
}
