package dbutil

import (
	"bytes"
	"html/template"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	//The blank imports are used to automatically register the database handlers
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// This is the relevant schema used in ConnectorDB.
const dbSchema = `

CREATE TABLE connectordbmeta (
  key VARCHAR UNIQUE NOT NULL,
  value VARCHAR NOT NULL);

CREATE TABLE users (
	userid {{.pkey_exp}},
	name VARCHAR UNIQUE NOT NULL,
	nickname VARCHAR DEFAULT '',
	email VARCHAR UNIQUE NOT NULL,
	description VARCHAR(1000) DEFAULT '',
	icon		VARCHAR(4096) DEFAULT '', -- DATA URI

	public BOOLEAN DEFAULT FALSE,
	role VARCHAR NOT NULL,

	password VARCHAR NOT NULL,
	passwordsalt VARCHAR NOT NULL,
	passwordhashscheme VARCHAR NOT NULL);

CREATE UNIQUE INDEX UserNameIndex ON users (name);

CREATE TABLE devices (
	deviceid {{.pkey_exp}},
	name VARCHAR NOT NULL,
	nickname VARCHAR DEFAULT '',
	description VARCHAR(1000) DEFAULT '',
	icon		VARCHAR(4096) DEFAULT '', -- DATA URI

	userid INTEGER,
	apikey VARCHAR NOT NULL,
	enabled BOOLEAN DEFAULT TRUE,

	public BOOLEAN DEFAULT FALSE,

	role VARCHAR DEFAULT '',


	isvisible BOOLEAN DEFAULT TRUE,
	usereditable BOOLEAN DEFAULT TRUE,
	UNIQUE(userid, name),
	FOREIGN KEY(userid) REFERENCES users(userid) ON DELETE CASCADE);


CREATE INDEX DeviceNameIndex ON devices (name);
CREATE UNIQUE INDEX DeviceAPIIndex ON devices (apikey) WHERE apikey!='';
CREATE INDEX DeviceUserIndex ON devices (userid);

CREATE TABLE streams (
	streamid {{.pkey_exp}},
	name VARCHAR NOT NULL,
	nickname VARCHAR NOT NULL DEFAULT '',
	description VARCHAR(1000) DEFAULT '',
	icon		VARCHAR(4096) DEFAULT '',
	schema VARCHAR NOT NULL,
	datatype VARCHAR DEFAULT '',
	deviceid INTEGER,
	ephemeral BOOLEAN DEFAULT FALSE,
	downlink BOOLEAN DEFAULT FALSE,
	UNIQUE(name, deviceid),
	FOREIGN KEY(deviceid) REFERENCES devices(deviceid) ON DELETE CASCADE);


CREATE INDEX StreamNameIndex ON streams (name);
CREATE INDEX StreamDeviceIndex ON streams (deviceid);


CREATE TABLE datastream (
	streamid BIGINT NOT NULL,
	substream VARCHAR,
	endtime DOUBLE PRECISION,
	endindex BIGINT,
	version INTEGER,
	data BYTEA,
	UNIQUE (streamid, substream, endindex),
	PRIMARY KEY (streamid, substream, endindex)
);

CREATE INDEX datastreamtime ON datastream (streamID,substream,endtime ASC);

INSERT INTO connectordbmeta VALUES ('DBVersion', '20160820');
`

// postgresFunctions allow certain things to happen automatically in postgres,
// which is safer than doing them manually (as is done in sqlite)
const postgresFunctions = `

-- Create the user and meta Devices for the user when a user is created
CREATE FUNCTION initial_user_setup() RETURNS TRIGGER AS $_$
DECLARE
	var_deviceid INTEGER;
BEGIN
	INSERT INTO Devices (Name, UserID, APIKey,Role, Description, Icon)
		VALUES ('user', NEW.UserID, NEW.PasswordSalt, 'user', 'Holds manually inserted data for the user','material:person');

	INSERT INTO Devices (Name, UserID, APIKey, Description, UserEditable, IsVisible, Icon) VALUES ('meta', NEW.UserID, '','The meta device holds automatically generated streams', FALSE, FALSE, 'material:bug_report');

	RETURN NEW;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER initialize_user AFTER INSERT ON Users FOR EACH ROW
	EXECUTE PROCEDURE initial_user_setup();
`

// deleteSchema removes EVERYTHING from a postgres database.
const deletePostgresSchema = `
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
SET search_path = public;
`

func getSchemaString(dbtype string) (string, error) {
	templateParams := make(map[string]string)
	if dbtype == "postgres" {
		templateParams["pkey_exp"] = "SERIAL PRIMARY KEY"
	} else {
		templateParams["pkey_exp"] = "INTEGER PRIMARY KEY AUTOINCREMENT"
	}
	schemaTemplate, err := template.New("dbschema").Parse(dbSchema)
	if err != nil {
		return "", err
	}
	var doc bytes.Buffer
	schemaTemplate.Execute(&doc, templateParams)
	return doc.String(), nil
}

// SetupDatabase creates the ConnectorDB database schema
func SetupDatabase(dbtype, uri string) error {
	log.Debugf("Setting up %s database at %s", dbtype, uri)
	db, err := sqlx.Open(dbtype, uri)
	if err != nil {
		return err
	}

	schemaString, err := getSchemaString(dbtype)
	if err != nil {
		return err
	}

	_, err = db.Exec(schemaString)

	if err != nil {
		return err
	}

	// If it is a postgres database, set up the built-in functions
	if dbtype == "postgres" {
		_, err = db.Exec(postgresFunctions)
	}

	db.Close()

	return err
}

// ClearDatabase removes all data from the database
func ClearDatabase(dbtype, uri string) error {
	log.Warnf("Clearing %s database at %s", dbtype, uri)
	if dbtype == "sqlite3" {
		// If it is an sqlite database, we just delete the file :)
		return os.Remove(uri)
	}
	db, err := sqlx.Open(dbtype, uri)
	if err != nil {
		return err
	}
	_, err = db.Exec(deletePostgresSchema)
	return err

}
