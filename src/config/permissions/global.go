package permissions

import (
	"path/filepath"
	"util"
)

// globalPermissions works in exactly the same way as globalConfiguration
var globalPermissions *PermissionsLoader

// Get returns the global permissions of the system
func Get() *Permissions {
	if globalPermissions == nil {
		return &Default
	}
	return globalPermissions.Get()
}

// SetPath attempts to set the global permissions to the given file
func SetPath(permissions string) error {
	if permissions != "default" {
		pl, err := NewPermissionsLoader(permissions)
		if err != nil {
			return err
		}
		if globalPermissions != nil {
			globalPermissions.Close()
		}
		globalPermissions = pl

	} else {
		if globalPermissions != nil {
			globalPermissions.Close()
		}
		globalPermissions = nil
	}
	return nil
}

// PermissionsLoader is basically same as ConfigurationLoader but without callbacks
type PermissionsLoader struct {
	Permissions *Permissions      // The currently loaded permissions
	Watcher     *util.FileWatcher // The watcher watches for changes to the permissions file
}

// Get returns the current permissions
func (pl *PermissionsLoader) Get() *Permissions {
	if pl == nil {
		// This is a rare instance where permissions were changed to default in between instructions
		return &Default
	}

	pl.Watcher.RLock()
	defer pl.Watcher.RUnlock()
	return pl.Permissions
}

// NewPermissionsLoader loads permissions from file, and watches them for changes, auto-updating
// when they are modified
func NewPermissionsLoader(filename string) (*PermissionsLoader, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	p, err := Load(filename)
	if err != nil {
		return nil, err
	}

	pl := &PermissionsLoader{
		Permissions: p,
	}
	pl.Watcher, err = util.NewFileWatcher(filename, pl)
	return pl, err
}

// Reload attempts to reload the permissions from file
func (pl *PermissionsLoader) Reload() error {
	p, err := Load(pl.Watcher.FileName)
	if err != nil {
		return err
	}

	pl.Watcher.Lock()
	pl.Permissions = p
	pl.Watcher.Unlock()

	return nil
}

// Close stops file listening
func (pl *PermissionsLoader) Close() {
	pl.Watcher.Close()
}
