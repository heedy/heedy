package updater

import (
	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

type Options struct {
	ConfigDir   string
	AddonConfig *assets.Configuration
	Runner      func(a *assets.Assets) error
	Revert      bool
	Update      bool
}

func Run(o Options) error {
	hadUpdate := false
	if o.Update {
		var err error
		hadUpdate, err = Update(o.ConfigDir)
		if err != nil {
			return err
		}
	}

	if hadUpdate {
		o.Revert = true
	}

	// Actually run it
	a, err := assets.Open(o.ConfigDir, o.AddonConfig)
	if err == nil {
		assets.SetGlobal(a)
		defer a.Close()
		err = o.Runner(a)
	}

	if o.Revert && err != nil {
		logrus.Error(err)
		err = Revert(o.ConfigDir, err)
		if err != nil {
			return err
		}

		return StartHeedy(o.ConfigDir, true)
	}

	return err
}
