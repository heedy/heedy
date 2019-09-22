package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/heedy/heedy/backend/assets"
	"github.com/jmoiron/sqlx"

	// Make sure we include sqlite support
	_ "github.com/mattn/go-sqlite3"
)

const schema = `

-- This is a meta-table, which specifies the versions of database tables
-- Every plugin that includes tables in the core database must add itself to the table
CREATE TABLE heedy (
	name VARCHAR(36) PRIMARY KEY NOT NULL,
	version INTEGER
);

-- This makes sure that the heedy version is specified, so that future upgrades will know
-- whether a schema modification is necessary
INSERT INTO heedy VALUES ("heedy",1);

-- A user is a group with an additional password. The id is a group id, we will
-- add the foreign key constraint once the groups table is created.
CREATE TABLE users (
	username VARCHAR(36) PRIMARY KEY NOT NULL,
	name VARCHAR NOT NULL DEFAULT '',
	description VARCHAR NOT NULL DEFAULT '',
	avatar VARCHAR NOT NULL DEFAULT '',

	-- whether the public or users can read the user
	public_read BOOLEAN NOT NULL DEFAULT FALSE,
	users_read BOOLEAN NOT NULL DEFAULT FALSE,

	password VARCHAR NOT NULL,

	UNIQUE(username)
);

CREATE INDEX useraccess ON users(public_read,users_read);

CREATE TABLE connections (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,

	name VARCHAR NOT NULL,
	description VARCHAR NOT NULL DEFAULT '',
	avatar VARCHAR NOT NULL DEFAULT '',

	owner VARACHAR(36) NOT NULL,

	-- Can (but does not have to) have an access token
	access_token VARCHAR UNIQUE DEFAULT NULL,

	-- Permissions are granted to a connection through scopes
	scopes VARCHAR NOT NULL DEFAULT '[]',

	settings VARCHAR DEFAULT '{}',
	setting_schema VARCHAR DEFAULT '{}',

	enabled BOOLEAN NOT NULL DEFAULT TRUE,

	-- the "plugin key" of the connection if it was generated for a plugin
	plugin VARCHAR DEFAULT NULL,

	UNIQUE(owner,plugin),

	CONSTRAINT valid_settings CHECK (json_valid(settings)),
	CONSTRAINT valid_settings_schema CHECK (json_valid(setting_schema)),

	CONSTRAINT connectionowner
		FOREIGN KEY(owner) 
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
-- We will want to list connections by owner 
CREATE INDEX connectionowner ON connections(owner,name);
-- A lot of querying will happen by API key
CREATE INDEX connectiontoken ON connections(access_token);


CREATE TABLE sources (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,
	name VARCHAR NOT NULL,
	description VARCHAR NOT NULL DEFAULT '',
	avatar VARCHAR NOT NULL DEFAULT '',
	connection VARCHAR(36) DEFAULT NULL,
	owner VARCHAR(36) NOT NULL,

	-- A key is used for connections to easily map sources to physical things
	key VARCHAR(36) DEFAULT NULL,

	type VARCHAR NOT NULL, 	            -- The source type
	meta VARCHAR NOT NULL DEFAULT '{}', -- Metadata for the source

	-- Maximal scopes that can be given. The * represents all scopes possible for the given source type
	scopes VARCHAR NOT NULL DEFAULT '["*"]',

	CONSTRAINT sourceconnection
		FOREIGN KEY(connection) 
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	CONSTRAINT sourceowner
		FOREIGN KEY(owner) 
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	CONSTRAINT valid_scopes CHECK (json_valid(scopes)),
	CONSTRAINT valid_meta CHECK (json_valid(meta))
);

-- Sources can be queried by key
CREATE INDEX source_key ON sources(key,connection);

------------------------------------------------------------------------------------
-- SHARING
------------------------------------------------------------------------------------

CREATE TABLE shared_sources (
	username VARCHAR(36) NOT NULL,
	sourceid VARCHAR(36) NOT NULL,
	scopes VARCHAR NOT NULL DEFAULT '["read"]',

	PRIMARY KEY (username,sourceid),
	UNIQUE (username,sourceid),

	CONSTRAINT sourceuser
		FOREIGN KEY(username)
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	CONSTRAINT sharedsource
		FOREIGN KEY(sourceid)
		REFERENCES sources(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	CONSTRAINT valid_scopes CHECK (json_valid(scopes))
);

CREATE INDEX share_sourceid on shared_sources(sourceid);


------------------------------------------------------------------
-- User Login Tokens
------------------------------------------------------------------
-- These are used to control manually logged in devices,
-- so that we don't need to put passwords in cookies

CREATE TABLE user_logintokens (
	username VARCHAR(36) NOT NULL,
	token VARCHAR UNIQUE NOT NULL,

	CONSTRAINT fk_user
		FOREIGN KEY(username) 
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

-- This will be requested on every single query
CREATE INDEX login_tokens ON user_logintokens(token);

------------------------------------------------------------------
-- Key-Value Storage for Plugins & Frontend
------------------------------------------------------------------

-- The given storage allows the frontend to save settings and such
CREATE TABLE frontend_kv (
	user VARCHAR(36) NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR DEFAULT '',
	include BOOLEAN DEFAULT FALSE, -- whether or not the key is included when the map is returned, or whether it needs to be queried.

	PRIMARY KEY(user,key),

	CONSTRAINT kvuser
		FOREIGN KEY(user) 
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE plugin_kv (
	plugin VARCHAR,
	-- Plugins can optionally save keys by user, where the key
	-- is automatically life-cycled with the user
	user VARCHAR DEFAULT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR DEFAULT '',
	include BOOLEAN DEFAULT FALSE, -- whether or not the key is included when the map is returned, or whether it should be queried

	PRIMARY KEY(plugin,user,key),
	UNIQUE(plugin,user,key),

	CONSTRAINT kvuser
		FOREIGN KEY(user) 
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

------------------------------------------------------------------
-- Database Views
------------------------------------------------------------------

CREATE VIEW user_source_scopes(user,source,scope) AS
	SELECT sources.owner,sources.id,'*' FROM sources WHERE sources.connection IS NULL
	UNION ALL
	SELECT sources.owner,sources.id,value FROM sources,json_each(sources.scopes) WHERE sources.connection IS NOT NULL
	UNION ALL
	SELECT shared_sources.username,sources.id,ss.value FROM sources,shared_sources,json_each(shared_sources.scopes) AS ss WHERE shared_sources.sourceid=sources.id AND ss.value<>'*' AND EXISTS (SELECT sss.value FROM json_each(sources.scopes) AS sss WHERE sss.value=ss.value OR sss.value='*')
	UNION ALL
	SELECT shared_sources.username,sources.id,sss.value FROM sources,shared_sources,json_each(sources.scopes) AS sss WHERE shared_sources.sourceid=sources.id AND EXISTS (SELECT 1 FROM json_each(shared_sources.scopes) AS ss WHERE ss.value='*')
	;


------------------------------------------------------------------
-- Database Default Users
------------------------------------------------------------------

INSERT INTO users (username,name,description,avatar,password) VALUES 
-- The public/users virtual users are created by default, and cannot be deleted,
-- as they represent the database view that someone not logged in will get,
-- and the sources accessible to a user who is logged in
(
	'users',
	'Users',
	'All logged-in users',
	'perm_identity',
	'-'
),(
	'public',
	'Public',
	'Represents everyone - any visitor to this heedy server',
	'share',
	'-'
),
-- The heedy user represents the database internals. It is used as the actor
-- when the software or plugins do something 
(
	"heedy",
	"Heedy",
	"",
	"remove_red_eye",
	"-"
);

`

// Create sets up a new heedy instance
func Create(a *assets.Assets) error {

	if a.Config.SQL == nil {
		return errors.New("Configuration does not specify an sql database")
	}

	// Split the sql string into database type and connection string
	sqlInfo := strings.SplitAfterN(*a.Config.SQL, "://", 2)
	if len(sqlInfo) != 2 {
		return errors.New("Invalid sql connection string")
	}
	sqltype := strings.TrimSuffix(sqlInfo[0], "://")

	if sqltype != "sqlite3" {
		return fmt.Errorf("Database type '%s' not supported", sqltype)
	}

	// We use the sql as location of our sqlite database
	sqlpath := a.DataAbs(sqlInfo[1])

	// Create any necessary directories
	sqlfolder := filepath.Dir(sqlpath)
	if err := os.MkdirAll(sqlfolder, 0750); err != nil {
		return err
	}

	// Notice we don't open the _heedy database here, since we don't want events
	// to fire while creating the database!
	db, err := sqlx.Open(sqltype, sqlpath)
	if err != nil {
		return err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	adb := &AdminDB{
		a: a,
	}
	adb.SqlxCache.InitCache(db)

	// Need to initialize all registered plugins
	err = initRegisteredPlugins(adb)
	if err != nil {
		return err
	}

	return adb.Close()
}
