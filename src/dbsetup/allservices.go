package dbsetup

import (
	"config"
	"connectordb/users"
	"dbsetup/dbutil"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"util"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrDirectoryExists is thrown if try to init over existing directory
	ErrDirectoryExists = errors.New("Cannot initialize database in an existing directory")
	//ErrDirectoryDNE is thrown if try start/run in existing directory
	ErrDirectoryDNE = errors.New("The given directory does not exist")
	//ErrAlreadyRunning is thrown if the database is already running
	ErrAlreadyRunning = errors.New("It looks like the ConnectorDB backend is already running.")
	//ErrNotRunning is thrown if the pid file does not exist
	ErrNotRunning = errors.New("Could not find pid file for ConnectorDB backend. Looks like the servers are off already.")
)

//Create generates a new ConnectorDB database
func Create(o *Options) error {
	c := o.Config

	if util.PathExists(o.DatabaseDirectory) {
		return ErrDirectoryExists
	}

	log.Infof("Creating new ConnectorDB database at '%s'", o.DatabaseDirectory)
	if err := os.MkdirAll(o.DatabaseDirectory, FolderPermissions); err != nil {
		return err
	}

	var umaker *users.UserMaker
	if o.InitialUser != nil && o.InitialUser.Name != "" {
		umaker = &users.UserMaker{User: users.User{
			Name:        o.InitialUser.Name,
			Email:       o.InitialUser.Email,
			Password:    o.InitialUser.Password,
			Description: o.InitialUser.Description,
			Icon:        o.InitialUser.Icon,
			Role:        o.InitialUser.Role,
			Nickname:    o.InitialUser.Nickname,
			Public:      o.InitialUser.Public,
		}}
	}

	//Now generate the conf file for the full configuration
	dbconf := filepath.Join(o.DatabaseDirectory, "connectordb.conf")
	err := c.Save(dbconf)
	if err != nil {
		return err
	}

	// Set that conf file as the globalConfiguration
	config.SetPath(dbconf)

	if c.Redis.Enabled && o.RedisEnabled {
		if err = NewRedisService(o.DatabaseDirectory, c).Create(); err != nil {
			return err
		}
	}
	if c.Nats.Enabled && o.GnatsdEnabled {
		if err = NewGnatsdService(o.DatabaseDirectory, c).Create(); err != nil {
			return err
		}
	}
	if c.Sql.Enabled && o.SQLEnabled {
		var p Service

		if c.Sql.Type == "postgres" {
			p = NewPostgresService(o.DatabaseDirectory, c)
		} else if c.Sql.Type == "sqlite3" {
			c.Sql.URI = filepath.Join(o.DatabaseDirectory, "db.sqlite3")
			p = NewSqliteService(c.Sql)
		} else {
			return fmt.Errorf("Unrecognized sql database type %s", c.Sql.Type)
		}

		if err = p.Create(); err != nil {
			return err
		}
		//Stop the database once finished with it
		defer p.Stop()

		//Now that the databases are all created, and postgres is running, we check if we are to create a default user
		if umaker != nil {
			log.Infof("Creating user %s (%s)", o.InitialUser.Name, o.InitialUser.Email)
			db, err := dbutil.OpenDatabase(c.Sql.Type, c.Sql.GetSqlConnectionString())
			if err != nil {
				return err
			}
			defer db.Close()

			udb := users.NewUserDatabase(db, false, 0, 0, 0, 0)

			err = udb.CreateUser(umaker)
			if err != nil {
				return err
			}
		}
	}

	// The frontend server does not need any creation stuff - since we are IN the frontend server

	return nil
}

//Start starts a ConnectorDB database
func Start(o *Options) error {
	if !util.PathExists(o.DatabaseDirectory) {
		return ErrDirectoryDNE
	}
	pidfile := filepath.Join(o.DatabaseDirectory, "connectordb.pid")
	//Check if connectordb.pid exists - if it does, it means the servers are running.
	if util.PathExists(pidfile) {
		return ErrAlreadyRunning
	}

	c, err := config.Load(filepath.Join(o.DatabaseDirectory, "connectordb.conf"))
	if err != nil {
		return err
	}
	// Set the config
	o.Config = c

	var r Service
	var g Service
	var p Service
	if c.Redis.Enabled && o.RedisEnabled {
		r = NewRedisService(o.DatabaseDirectory, c)
		if err := r.Start(); err != nil {
			return err
		}
	}
	if c.Nats.Enabled && o.GnatsdEnabled {
		g = NewGnatsdService(o.DatabaseDirectory, c)
		if err := g.Start(); err != nil {
			if r != nil {
				r.Stop()
			}
			return err
		}
	}
	if c.Sql.Enabled && o.SQLEnabled {
		if c.Sql.Type == "postgres" {
			p = NewPostgresService(o.DatabaseDirectory, c)
			if err := p.Start(); err != nil {
				if r != nil {
					r.Stop()
				}
				if g != nil {
					g.Stop()
				}
				return err
			}
		}

	}

	if c.Frontend.Enabled && o.FrontendEnabled {
		f := NewFrontendService(o.DatabaseDirectory, c, o)
		if err := f.Start(); err != nil {
			if r != nil {
				r.Stop()
			}
			if g != nil {
				g.Stop()
			}
			if p != nil {
				p.Stop()
			}
			return err
		}
	}

	//Now we save the current options to the pid file
	o.Save(pidfile)
	return nil
}

//Stop stops a ConnectorDB database
func Stop(opt *Options) error {
	if !util.PathExists(opt.DatabaseDirectory) {
		return ErrDirectoryDNE
	}
	pidfile := filepath.Join(opt.DatabaseDirectory, "connectordb.pid")
	//Check if connectordb.pid exists - if it does, it means the servers are running
	if !util.PathExists(pidfile) {
		return ErrNotRunning
	}

	o, err := LoadOptions(pidfile)
	if err != nil {
		return err
	}
	c := o.Config

	var errR error
	var errG error
	var errP error
	var errF error

	if c.Frontend.Enabled && o.FrontendEnabled {
		// We close the frontend First
		errF = NewFrontendService(o.DatabaseDirectory, c, o).Stop()
	}

	if c.Redis.Enabled && o.RedisEnabled {
		errR = NewRedisService(o.DatabaseDirectory, c).Stop()
	}
	if c.Nats.Enabled && o.GnatsdEnabled {
		errG = NewGnatsdService(o.DatabaseDirectory, c).Stop()
	}
	if c.Sql.Enabled && o.SQLEnabled && c.Sql.Type == "postgres" {
		errP = NewPostgresService(o.DatabaseDirectory, c).Stop()
	}
	if errF != nil {
		return errF
	}
	if errR != nil {
		return errR
	}
	if errG != nil {
		return errG
	}

	os.Remove(pidfile)

	return errP
}

/*
//Kill stops a ConnectorDB database
func Kill(dbfolder string) error {
	if !util.PathExists(dbfolder) {
		return ErrDirectoryDNE
	}
	pidfile := filepath.Join(dbfolder, "connectordb.pid")
	//Check if connectordb.pid exists - if it does, it means the servers are running
	if !util.PathExists(pidfile) {
		return ErrNotRunning
	}

	c, err := config.Load(pidfile)
	if err != nil {
		return err
	}
	errF := NewFrontendService(o.DatabaseDirectory, c).Stop()
	errR := NewRedisService(o.DatabaseDirectory, &c.Redis).Stop()
	errG := NewGnatsdService(o.DatabaseDirectory, &c.Nats).Stop()
	errP := NewPostgresService(o.DatabaseDirectory, &c.Sql).Stop()
	if errF != nil {
		return errF
	}
	if errR != nil {
		return errR
	}
	if errG != nil {
		return errG
	}

	os.Remove(pidfile)
	return errP
}
*/
