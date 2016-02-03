/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package dbsetup

import (
	"config"
	"connectordb/users"
	"dbsetup/dbutil"
	"errors"
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
func Create(c *config.Configuration) error {

	if util.PathExists(c.DatabaseDirectory) {
		return ErrDirectoryExists
	}

	log.Infof("Creating new ConnectorDB database at '%s'", c.DatabaseDirectory)
	if err := os.MkdirAll(c.DatabaseDirectory, FolderPermissions); err != nil {
		return err
	}

	//Now generate the conf file for the full configuration
	dbconf := filepath.Join(c.DatabaseDirectory, "connectordb.conf")
	err := c.Save(dbconf)
	if err != nil {
		return err
	}

	// Set that conf file as the globalConfiguration
	config.SetPath(dbconf)

	if c.Redis.Enabled {
		if err = NewRedisService(c.DatabaseDirectory, &c.Redis).Create(); err != nil {
			return err
		}
	}
	if c.Nats.Enabled {
		if err = NewGnatsdService(c.DatabaseDirectory, &c.Nats).Create(); err != nil {
			return err
		}
	}
	if c.Sql.Enabled {
		p := NewPostgresService(c.DatabaseDirectory, &c.Sql)
		if err = p.Create(); err != nil {
			return err
		}
		//Stop the database once finished with it
		defer p.Stop()

		//Now that the databases are all created (and postgres is running), we check if we are to create a default user
		if c.InitialUsername != "" {
			log.Infof("Creating user %s (%s)", c.InitialUsername, c.InitialUserEmail)
			db, driver, err := dbutil.OpenSqlDatabase(c.GetSqlConnectionString())
			if err != nil {
				return err
			}
			defer db.Close()

			udb := users.NewUserDatabase(db, driver, false, 0, 0, 0)
			err = udb.CreateUser(c.InitialUsername, c.InitialUserEmail, c.InitialUserPassword, c.InitialUserRole, c.InitialUserPublic, 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//Start starts a ConnectorDB database
func Start(dbfolder string) error {
	if !util.PathExists(dbfolder) {
		return ErrDirectoryDNE
	}
	pidfile := filepath.Join(dbfolder, "connectordb.pid")
	//Check if connectordb.pid exists - if it does, it means the servers are running
	if util.PathExists(pidfile) {
		return ErrAlreadyRunning
	}

	c, err := config.Load(filepath.Join(dbfolder, "connectordb.conf"))
	if err != nil {
		return err
	}
	//Overwrite the database directory
	c.DatabaseDirectory = dbfolder

	var r Service
	var g Service
	if c.Redis.Enabled {
		r = NewRedisService(c.DatabaseDirectory, &c.Redis)
		if err := r.Start(); err != nil {
			return err
		}
	}
	if c.Nats.Enabled {
		g = NewGnatsdService(c.DatabaseDirectory, &c.Nats)
		if err := g.Start(); err != nil {
			if r != nil {
				r.Stop()
			}
			return err
		}
	}
	if c.Sql.Enabled {
		p := NewPostgresService(c.DatabaseDirectory, &c.Sql)
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

	//Now we save the current config to the pid file
	c.Save(pidfile)
	return nil
}

//Stop stops a ConnectorDB database
func Stop(dbfolder string) error {
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
	//Overwrite the database directory
	c.DatabaseDirectory = dbfolder
	var errR error
	var errG error
	var errP error
	if c.Redis.Enabled {
		errR = NewRedisService(c.DatabaseDirectory, &c.Redis).Stop()
	}
	if c.Nats.Enabled {
		errG = NewGnatsdService(c.DatabaseDirectory, &c.Nats).Stop()
	}
	if c.Sql.Enabled {
		errP = NewPostgresService(c.DatabaseDirectory, &c.Sql).Stop()
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

	errR := NewRedisService(c.DatabaseDirectory, &c.Redis).Stop()
	errG := NewGnatsdService(c.DatabaseDirectory, &c.Nats).Stop()
	errP := NewPostgresService(c.DatabaseDirectory, &c.Sql).Stop()
	if errR != nil {
		return errR
	}
	if errG != nil {
		return errG
	}

	os.Remove(pidfile)
	return errP
}
