package dbsetup

import (
	"config"
	"connectordb/users"
	"errors"
	"os"
	"path/filepath"
	"util"
	"util/dbutil"

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
	err := c.Save(filepath.Join(c.DatabaseDirectory, "connectordb.conf"))
	if err != nil {
		return err
	}

	if err = NewRedisService(c.DatabaseDirectory, &c.Redis).Create(); err != nil {
		return err
	}
	if err = NewGnatsdService(c.DatabaseDirectory, &c.Nats).Create(); err != nil {
		return err
	}

	p := NewPostgresService(c.DatabaseDirectory, &c.Sql)
	if err = p.Create(); err != nil {
		return err
	}
	//Stop the database once finished with it
	defer p.Stop()

	//Now that the databases are all created (and postgres is running), we check if we are to create a default user
	if c.Username != "" {
		log.Infof("Creating user %s (%s)", c.InitialUsername, c.InitialUserEmail)
		db, driver, err := dbutil.OpenSqlDatabase(c.GetSqlConnectionString())
		if err != nil {
			return err
		}
		defer db.Close()

		udb := users.NewUserDatabase(db, driver, false)
		err = udb.CreateUser(c.InitialUsername, c.InitialUserEmail, c.InitialUserPassword)
		if err != nil {
			return err
		}
		usr, err := udb.ReadUserByName(c.InitialUsername)
		if err != nil {
			return err
		}
		usr.Admin = true
		err = udb.UpdateUser(usr)
		if err != nil {
			return err
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

	r := NewRedisService(c.DatabaseDirectory, &c.Redis)
	if err := r.Start(); err != nil {
		return err
	}
	g := NewGnatsdService(c.DatabaseDirectory, &c.Nats)
	if err := g.Start(); err != nil {
		r.Stop()
		return err
	}
	p := NewPostgresService(c.DatabaseDirectory, &c.Sql)
	if err := p.Start(); err != nil {
		r.Stop()
		g.Stop()
		return err
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
