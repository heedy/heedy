package dbmaker

import (
	"path/filepath"
	"streamdb/config"
	"streamdb/dbutil"
	"streamdb/util"

	log "github.com/Sirupsen/logrus"
)

var sqliteDatabaseName = "streamdb.sqlite3"

// A service representing the postgres database
type SqliteService struct {
	ServiceHelper     // We get stop, status, kill, and Name from this
	streamdbDirectory string
	sqliteFilepath    string
}

// Creates and returns a new postgres service in a pre-init state
// with default values loaded from config
func NewDefaultSqliteService() *SqliteService {
	return NewConfigSqliteSerivce(config.GetConfiguration())
}

func NewConfigSqliteSerivce(config *config.Configuration) *SqliteService {
	dir := config.StreamdbDirectory
	return NewSqliteService(dir)
}

// Creates and returns a new postgres service in a pre-init state
func NewSqliteService(streamdbDirectory string) *SqliteService {
	var ps SqliteService
	ps.sqliteFilepath = filepath.Join(streamdbDirectory, sqliteDatabaseName)
	ps.streamdbDirectory = streamdbDirectory

	ps.InitServiceHelper(streamdbDirectory, config.Sqlite)
	return &ps
}

//InitializeSqlite creates an sqlite database and subsequently sets it up to work with streamdb
func (srv *SqliteService) Setup() error {
	log.Printf("Initializing sqlite database '%s'\n", srv.sqliteFilepath)

	// because sqlite doesn't always like being started on a file that
	// doesn't exist
	util.Touch(srv.sqliteFilepath)

	//Initialize the database tables
	log.Printf("Setting up initial tables\n")
	return dbutil.UpgradeDatabase(srv.sqliteFilepath, true)
}

func (srv *SqliteService) Init() error {
	srv.Stat = StatusInit
	return nil
}

//StartRedis runs the redis server
func (srv *SqliteService) Start() error {
	srv.Stat = StatusRunning
	return nil
}

func (srv *SqliteService) Stop() error {
	srv.Stat = StatusInit
	return nil
}

func (srv *SqliteService) Kill() error {
	srv.Stat = StatusInit
	return nil
}
