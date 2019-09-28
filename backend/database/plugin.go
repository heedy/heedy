package database

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

// PluginFunc is the function that, when given an AdminDB, generates the necessary database tables
// based on the version, and does all relevant initialization
type PluginFunc func(*AdminDB, int) error

// PluginInfo contains the information necessary to create a plugin's database tables
type PluginInfo struct {
	Name    string
	Version int
	OnOpen PluginFunc
}

var pluginArray = make([]*PluginInfo, 0)

// RegisterPlugin adds the plugin so that its database tables are auto-created when the database is created
func RegisterPlugin(pluginName string, dbversion int, onOpen PluginFunc) {
	pluginArray = append(pluginArray, &PluginInfo{
		Name:    pluginName,
		Version: dbversion,
		OnOpen: onOpen,
	})
}

// InitPlugin calls Updater if either the plugin doesn't exist, or the version doesn't match
func InitPlugin(db *AdminDB, p *PluginInfo) error {
	var curVersion int
	err := db.Get(&curVersion, `SELECT version FROM heedy WHERE name=?`, p.Name)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		curVersion = 0
	}
	err = p.OnOpen(db, curVersion)
	if err != nil {
		return err
	}
	// Now update the version in heedy
	if p.Version!=curVersion {
		_, err = db.Exec(`INSERT OR REPLACE INTO heedy(name,version) VALUES (?,?)`, p.Name, p.Version)
	}
	return err
}

func initRegisteredPlugins(db *AdminDB) error {
	for _, v := range pluginArray {
		logrus.Debugf("Initializing %s plugin backend (v%d)",v.Name,v.Version)
		if err := InitPlugin(db, v); err != nil {
			return err
		}
	}
	return nil
}
