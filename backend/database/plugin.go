package database

import "database/sql"

// PluginFunc is the function that, when given an AdminDB, generates the necessary database tables
type PluginFunc func(*AdminDB, int) error

// PluginInfo contains the information necessary to create a plugin's database tables
type PluginInfo struct {
	Name    string
	Version int
	Updater PluginFunc
}

var pluginArray = make([]*PluginInfo, 0)

// RegisterPlugin adds the plugin so that its database tables are auto-created when the database is created
func RegisterPlugin(pluginName string, dbversion int, updater PluginFunc) {
	pluginArray = append(pluginArray, &PluginInfo{
		Name:    pluginName,
		Version: dbversion,
		Updater: updater,
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
	if curVersion == p.Version {
		return nil
	}

	err = p.Updater(db, curVersion)
	if err != nil {
		return err
	}
	// Now update the version in heedy
	_, err = db.Exec(`INSERT OR REPLACE INTO heedy(name,version) VALUES (?,?)`, p.Name, p.Version)
	return err
}

func initRegisteredPlugins(db *AdminDB) error {
	for _, v := range pluginArray {
		if err := InitPlugin(db, v); err != nil {
			return err
		}
	}
	return nil
}
