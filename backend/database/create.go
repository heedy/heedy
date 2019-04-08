package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/heedy/heedy/backend/assets"

	// Make sure we include sqlite support
	_ "github.com/mattn/go-sqlite3"
)

const schema = `

-- A user is a group with an additional password. The id is a group id, we will
-- add the foreign key constraint once the groups table is created.
CREATE TABLE users (
	name VARCHAR(36) PRIMARY KEY NOT NULL,
	fullname VARCHAR DEFAULT '',
	description VARCHAR DEFAULT '',
	avatar VARCHAR DEFAULT '',

	-- 100 is read user - there is currently no further access available for users
	public_access INTEGER DEFAULT 0, -- access of the public to the user
	user_access INTEGER DEFAULT 0, -- access of all users to the user

	password VARCHAR NOT NULL,

	UNIQUE(name)
);

CREATE INDEX useraccess ON users(public_access,user_access);

-- Groups are the underlying container for access control and sharing
CREATE TABLE groups (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,

	name VARCHAR NOT NULL,
	fullname VARCHAR DEFAULT '',
	description VARCHAR DEFAULT '',
	avatar VARCHAR DEFAULT '',
	
	public_access INTEGER DEFAULT 0, -- access of the public to the group
	user_access INTEGER DEFAULT 0, -- access of all users to the group

	owner VARCHAR(36) NOT NULL,

	CONSTRAINT groupowner
		FOREIGN KEY(owner) 
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE INDEX groupaccess ON groups(public_access,user_access);
CREATE INDEX groupowner ON groups(owner);
	

CREATE TABLE connections (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,

	name VARCHAR NOT NULL,
	fullname VARCHAR DEFAULT '',
	description VARCHAR DEFAULT '',
	avatar VARCHAR DEFAULT '',

	owner VARACHAR(36) NOT NULL,

	-- Can (but does not have to) have an API key
	apikey VARCHAR UNIQUE DEFAULT NULL,

	settings VARCHAR DEFAULT '{}',
	setting_schema VARCHAR DEFAULT '{}',

	CONSTRAINT connectionowner
		FOREIGN KEY(owner) 
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
-- We will want to list connections by owner 
CREATE INDEX connectionowner ON connections(owner,name);
-- A lot of querying will happen by API key
CREATE INDEX connectionapikey ON connections(apikey);


CREATE TABLE streams (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,
	name VARCHAR NOT NULL,
	fullname VARCHAR DEFAULT '',
	description VARCHAR DEFAULT '',
	avatar VARCHAR DEFAULT '',
	connection VARCHAR(36) DEFAULT NULL,
	owner VARCHAR(36) NOT NULL,

	-- json schema
	schema VARCHAR DEFAULT '{}',
	type VARCHAR DEFAULT '',

	-- Set to '' when the stream is internal, and gives the rest url/plugin uri for querying if external
	external VARCHAR DEFAULT '',

	actor BOOLEAN DEFAULT FALSE, -- Whether the stream is also an actor, ie, it can take action, meaning that it performs interventions

	-- What access is given to the user and others who have access to the stream
	access INTEGER DEFAULT 200, -- 0 hidden, 100 read, 200 insert actions, 300 modify, 400 insert data, 500 remove data, 600 delete

	CONSTRAINT streamconnection
		FOREIGN KEY(connection) 
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	CONSTRAINT streamowner
		FOREIGN KEY(owner) 
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

------------------------------------------------------------------------------------
-- USER SCOPES
------------------------------------------------------------------------------------

-- Scopesets are sets of scopes which correspond to a scopeset
CREATE TABLE scopesets (
	name VARCHAR(36) NOT NULL,
	scope VARCHAR(36) NOT NULL,

	PRIMARY KEY (name,scope)
);

-- A user is given a scope set
CREATE TABLE user_scopesets (
	user VARCHAR(36) NOT NULL,
	scopeset VARCHAR NOT NULL,
	PRIMARY KEY (user,scopeset),
	CONSTRAINT fk_userss
		FOREIGN KEY(user)
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

------------------------------------------------------------------------------------
-- GROUP ACCESS
------------------------------------------------------------------------------------

CREATE TABLE group_members (
	groupid VARCHAR(36),
	user VARCHAR(36),

	access INTEGER DEFAULT 200, -- 100 is read group, 200 is readonly all, 300 gives stream insert access, 400 allows adding/removing own streams/sources, 500 allows removing streams/sources, 600 allows adding/removing members (except owner)
	
	PRIMARY KEY (groupid,user),

	CONSTRAINT idid
		FOREIGN KEY(user)
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT groupid
		FOREIGN KEY(groupid)
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE group_streams (
	groupid VARCHAR(36),
	streamid VARCHAR(36),

	access INTEGER DEFAULT 100, -- Same as stream access

	UNIQUE(groupid,streamid),
	PRIMARY KEY (groupid,streamid),

	CONSTRAINT idid
		FOREIGN KEY(streamid)
		REFERENCES streams(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT groupid
		FOREIGN KEY(groupid)
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

------------------------------------------------------------------------------------
-- CONNECTION ACCESS
------------------------------------------------------------------------------------

-- The scopes available to the connection
CREATE TABLE connection_scopes (
	connectionid VARCHAR(36) NOT NULL,
	scope VARCHAR NOT NULL,
	PRIMARY KEY (connectionid,scope),
	UNIQUE (connectionid,scope),
	CONSTRAINT fk_connectionid
		FOREIGN KEY(connectionid)
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

-- Streams that the connection is permitted to access
CREATE TABLE connection_streams (
	connectionid VARCHAR(36),
	streamid VARCHAR(36),

	access INTEGER DEFAULT 1, -- Same as stream access

	UNIQUE(connectionid,streamid),
	PRIMARY KEY (connectionid,streamid),

	CONSTRAINT cstreamid
		FOREIGN KEY(streamid)
		REFERENCES streams(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT cconnectionid
		FOREIGN KEY(connectionid)
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

-- Other connections that the connection is permitted to access
CREATE TABLE connection_connections (
	connection VARCHAR(36),
	id VARCHAR(36),

	access INTEGER DEFAULT 100, -- Same as stream access

	UNIQUE(connection,id),
	PRIMARY KEY (connection,id),

	CONSTRAINT idid
		FOREIGN KEY(id)
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT connectionid
		FOREIGN KEY(connection)
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

-- Groups that the connection is permitted to access
CREATE TABLE connection_groups (
	connection VARCHAR(36),
	id VARCHAR(36),

	access INTEGER DEFAULT 100, -- Same as stream access

	UNIQUE(connection,id),
	PRIMARY KEY (connection,id),

	CONSTRAINT idid
		FOREIGN KEY(id)
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT connectionid
		FOREIGN KEY(connection)
		REFERENCES connections(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

------------------------------------------------------------------
-- User Login Tokens
------------------------------------------------------------------
-- These are used to control manually logged in devices,
-- so that we don't need to put passwords in cookies

CREATE TABLE user_logintokens (
	user VARCHAR(36) NOT NULL,
	token VARCHAR UNIQUE NOT NULL,

	CONSTRAINT fk_user
		FOREIGN KEY(user) 
		REFERENCES users(name)
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
		REFERENCES users(name)
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
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

------------------------------------------------------------------
-- Database Views
------------------------------------------------------------------

-- A table containing all the scopes for each user
CREATE VIEW user_scopes(user,scope) AS
	SELECT x.name, scope FROM scopesets 
		JOIN (SELECT users.name AS name,user_scopesets.scopeset AS scopeset FROM users 
					LEFT JOIN user_scopesets ON users.name=user_scopesets.user
				) AS x 
		ON (x.scopeset=scopesets.name OR scopesets.name IN ('users', 'public')) AND x.name<>'heedy';

-- A table of user * user, containing all the possible usernames that can read the given target username
CREATE VIEW user_can_read_user(user,target) AS
	SELECT users.name, t.name  FROM users,users t WHERE users.name<>'heedy' AND (
		t.public_access >= 100 OR t.user_access >=100 OR users.name=t.name 
		OR 'users:read' IN (SELECT scope FROM user_scopes WHERE user_scopes.user=users.name));

-- A table containing a list of users that can be read by the public
CREATE VIEW public_can_read_user(user) AS
	SELECT users.name FROM users WHERE
		users.public_access >= 100 OR 'users:read' IN (SELECT scope FROM scopesets WHERE scopesets.name='public');


-- A table of all streams that the given user can read based on group membership alone
CREATE VIEW user_can_read_groupstreams(user,stream) AS
	SELECT x.name,group_streams.streamid FROM groups 
		JOIN (SELECT users.name,group_members.groupid from users LEFT JOIN group_members ON (group_members.user=users.name)) AS x 
			ON (x.name<>'heedy' AND (x.groupid=groups.id OR groups.public_access>=200 OR groups.user_access>=200 OR groups.owner=x.name)) 
		JOIN group_streams ON (groups.id=group_streams.groupid AND group_streams.access>=100);


-- A table containing all the usernames that can read each stream. This encodes some internal processing:
-- A user can read a stream if...
-- 	1) The user owns the stream
--	2) The user is a member of a group which gives access to the stream
CREATE VIEW user_can_read_stream(user,stream) AS
	SELECT users.name,streams.id FROM users JOIN streams ON users.name=streams.owner 
		OR streams.id IN (SELECT stream FROM user_can_read_groupstreams WHERE user=users.name)
		OR ('streams:read' IN (SELECT scope FROM user_scopes WHERE user=users.name) 
			AND EXISTS (SELECT 1 FROM user_can_read_user WHERE user=users.name AND target=streams.owner));

CREATE VIEW public_can_read_stream(stream) AS
	SELECT streams.id FROM streams WHERE
		streams.id IN (SELECT streamid FROM group_streams JOIN groups ON (groups.id=group_streams.groupid) WHERE groups.public_access >= 200)
		OR (streams.owner IN (SELECT user FROM public_can_read_user) AND 'streams:read' IN (SELECT scope FROM scopesets WHERE scopesets.name='public'));
	

------------------------------------------------------------------
-- Database Default Users & Groups
------------------------------------------------------------------

-- The public group is created by default, and cannot be deleted,
-- as it represents the database view that someone not logged in will get.

-- The heedy user represents the database internals. It is used as the actor
-- when the software or plugins do something
INSERT INTO users (name,fullname,description,avatar,password) VALUES (
	"heedy",
	"Heedy",
	"",
	"mi:remove_red_eye",
	"-"
);

-- the public group has ID public
INSERT INTO groups (id,name,fullname,description,avatar,owner,public_access) VALUES (
	"public",
	"public",
	"Public",
	"Make accessible to all visitors, even if they're not logged in",
	"mi:share",
	"heedy",
	400 -- Allows each user to add/remove their own streams/connections
);

-- the users group has ID users
INSERT INTO groups (id,name,fullname,description,avatar,owner,user_access) VALUES (
	"users",
	"users",
	"Users",
	"Make accessible to all logged-in users",
	"mi:supervised_user_circle",
	"heedy",
	400 -- Allows each user to add/remove their own streams/connections
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
	sqlpath := a.Abs(sqlInfo[1])

	// Create any necessary directories
	sqlfolder := filepath.Dir(sqlpath)
	if err := os.MkdirAll(sqlfolder, 0750); err != nil {
		return err
	}

	db, err := sql.Open(sqltype, sqlpath)
	if err != nil {
		return err
	}

	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	if sqltype == "sqlite3" {
		_, err = db.Exec(sqliteAddonSchema)
		if err != nil {
			return err
		}
	}

	return db.Close()
}
