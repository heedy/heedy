package assets

import (
	"sync"

	"github.com/spf13/afero"
)

var assetLock sync.RWMutex
var globalAssets *Assets

// Get returns global asset holder
func Get() *Assets {
	assetLock.RLock()
	defer assetLock.RUnlock()
	if globalAssets == nil {
		var err error
		globalAssets, err = NewAssets("", nil)
		if err != nil {
			panic(err.Error())
		}
	}
	return globalAssets
}

// Config returns the global configuration
func Config() *Configuration {
	return Get().Config
}

// FS returns the current filesystem
func FS() afero.Fs {
	return Get().FS
}

// SetGlobal sets the global assets to the given values
func SetGlobal(a *Assets) {
	assetLock.Lock()
	globalAssets = a
	assetLock.Unlock()
}
