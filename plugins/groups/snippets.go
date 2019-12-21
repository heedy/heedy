package group

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const schema = `

-- Groups are the underlying container for access control and sharing
CREATE TABLE groups (
	id VARCHAR(36) UNIQUE NOT NULL PRIMARY KEY,

	name VARCHAR NOT NULL,
	fullname VARCHAR NOT NULL DEFAULT '',
	description VARCHAR NOT NULL DEFAULT '',
	icon VARCHAR NOT NULL DEFAULT '',

	owner VARCHAR(36) NOT NULL,

	-- json array of scopes given to the public and to users.
	-- We use the empty array
	public_scopes VARCHAR NOT NULL DEFAULT '[]',
	user_scopes VARCHAR NOT NULL DEFAULT '[]',

	CONSTRAINT groupowner
		FOREIGN KEY(owner) 
		REFERENCES users(name)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	-- For public and user access, must explicitly give group read permission,
	-- which automatically gives read access to all objects
	CONSTRAINT scopes_readaccess CHECK (
		(public_scopes='[]' OR public_scopes LIKE '%"group:read"%')
		AND (user_scopes='[]' OR user_scopes LIKE '%"group:read"%')
	)
	CONSTRAINT scopes_valid CHECK (
		json_valid(public_scopes) AND json_valid(user_scopes)
	)
);

-- This index simply checks if there exists any scope in the arrays. It allows quickly determining if
-- the group has given no special permissions to the public or to users.
CREATE INDEX groupscopes ON groups(public_scopes<>'[]',user_scopes<>'[]');
CREATE INDEX groupowner ON groups(owner);


------------------------------------------------------------------------------------
-- GROUP ACCESS
------------------------------------------------------------------------------------

CREATE TABLE group_members (
	groupid VARCHAR(36),
	user VARCHAR(36),

	-- json array of scopes given to the group members.
	-- the group read scope is implied
	scopes VARCHAR NOT NULL DEFAULT '[]',
	
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
		ON DELETE CASCADE,
	
	CONSTRAINT valid_scopes CHECK (json_valid(scopes))
);

CREATE INDEX member_users ON group_members(user);

CREATE TABLE group_objects (
	groupid VARCHAR(36),
	objectid VARCHAR(36),

	scopes VARCHAR NOT NULL DEFAULT '[]',

	UNIQUE(groupid,objectid),
	PRIMARY KEY (groupid,objectid),

	CONSTRAINT objectid
		FOREIGN KEY(objectid)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT groupid
		FOREIGN KEY(groupid)
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,

	CONSTRAINT valid_scopes CHECK (json_valid(scopes))
);


/* COMMENTED OUT FOR NOW

-- Timeseries that the app is permitted to access
CREATE TABLE app_timeseries (
	appid VARCHAR(36),
	tsid VARCHAR(36),

	access INTEGER DEFAULT 1, -- Same as timeseries access

	UNIQUE(appid,tsid),
	PRIMARY KEY (appid,tsid),

	CONSTRAINT ctsid
		FOREIGN KEY(tsid)
		REFERENCES timeseries(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT cappid
		FOREIGN KEY(appid)
		REFERENCES apps(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

-- Other apps that the app is permitted to access
CREATE TABLE app_apps (
	appid VARCHAR(36),
	otherid VARCHAR(36),

	scopes VARCHAR NOT NULL DEFAULT '[]',

	UNIQUE(appid,otherid),
	PRIMARY KEY (appid,otherid),

	CONSTRAINT idid
		FOREIGN KEY(appid)
		REFERENCES apps(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT appid
		FOREIGN KEY(otherid)
		REFERENCES apps(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

-- Groups that the app is permitted to access
CREATE TABLE app_groups (
	appid VARCHAR(36),
	groupid VARCHAR(36),

	scopes VARCHAR NOT NULL DEFAULT '[]',

	UNIQUE(app,id),
	PRIMARY KEY (app,id),

	CONSTRAINT groupid
		FOREIGN KEY(groupid)
		REFERENCES groups(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE,
	
	CONSTRAINT appid
		FOREIGN KEY(appid)
		REFERENCES apps(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);
*/


CREATE VIEW public_groupscopes(groupid,scope) AS
	SELECT groups.id,json_each.value FROM groups,json_each(public_scopes) WHERE public_scopes<>"[]";

-- Scopes available to each user when accessing each group
-- All users have all public and user scopes
-- Group members have additional scopes on the group
-- group owners have all scopes (*)
CREATE VIEW user_groupscopes(username,groupid,scope) AS
	SELECT users.name,id,value FROM users JOIN (
			SELECT groups.id, json_each.value FROM groups,json_each(user_scopes) WHERE user_scopes<>"[]"
		UNION 
			SELECT * FROM public_groupscopes
		)
	UNION
		SELECT user,groupid,value FROM group_members,json_each(scopes)
	UNION
		SELECT owner,id,'*' FROM groups;

`

// ReadGroupOptions gives options for reading
type ReadGroupOptions struct {
	Icon bool `json:"icon,omitempty" schema:"icon"`
}

// Group holds a group's details
type Group struct {
	Details
	Owner *string `json:"owner" db:"owner"`

	PublicScopes *ScopeArray `json:"public_scopes" db:"public_scopes"`
	UserScopes   *ScopeArray `json:"user_scopes" db:"user_scopes"`
}

func readGroup(adb *AdminDB, id string, o *ReadGroupOptions, selectStatement string, args ...interface{}) (*Group, error) {
	g := &Group{}
	err := adb.Get(g, selectStatement, args...)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if o == nil || !o.Icon {
		g.Icon = nil
	}
	return g, err
}

func extractGroup(g *Group) (groupColumns []string, groupValues []interface{}, err error) {
	if g.Owner != nil {
		if err = ValidName(*g.Owner); err != nil {
			return
		}
	}

	if g.PublicScopes != nil {
		if err = ValidGroupScopes(g.PublicScopes); err != nil {
			return
		}
	}
	if g.UserScopes != nil {
		if err = ValidGroupScopes(g.UserScopes); err != nil {
			return
		}
	}

	groupColumns, groupValues, err = extractDetails(&g.Details)
	if err != nil {
		return nil, nil, err
	}
	c2, g2 := extractPointers(g)
	groupColumns = append(groupColumns, c2...)
	groupValues = append(groupValues, g2...)

	return
}

func groupCreateQuery(g *Group) (string, []interface{}, error) {
	if g.Name == nil {
		return "", nil, ErrInvalidName
	}
	if g.Owner == nil {
		// A group must have an owner
		return "", nil, ErrInvalidQuery
	}
	groupColumns, groupValues, err := extractGroup(g)
	if err != nil {
		return "", nil, err
	}

	// Since we are creating the details, we also set up the id of the group
	// We guarantee that ID is last element
	groupColumns = append(groupColumns, "id")
	gid := uuid.New().String()
	groupValues = append(groupValues, gid)
	g.ID = gid // Set the object's ID

	return strings.Join(groupColumns, ","), groupValues, nil

}

func groupUpdateQuery(g *Group) (string, []interface{}, error) {
	groupColumns, groupValues, err := extractGroup(g)
	if len(groupColumns) == 0 {
		return "", nil, ErrNoUpdate
	}
	return strings.Join(groupColumns, "=?,") + "=?", groupValues, err
}

func (db *UserDB) CreateGroup(g *Group) (string, error) {
	if g.Owner != nil && db.user != *g.Owner {
		return "", ErrAccessDenied("you can only create groups for your own user")
	}
	g.Owner = &db.user
	return db.adb.CreateGroup(g)
}

func (db *UserDB) ReadGroup(id string, o *ReadGroupOptions) (*Group, error) {
	return readGroup(db.adb, id, o, `SELECT groups.*, json_group_array(scope) AS access FROM groups 
		JOIN user_groupscopes ON (groups.id=user_groupscopes.groupid) 
		WHERE groups.id=? AND user_groupscopes.username=?`, id, db.user)
}

// CreateGroup generates a group with the given owner groupID
func (db *AdminDB) CreateGroup(g *Group) (string, error) {
	groupColumns, groupValues, err := groupCreateQuery(g)
	if err != nil {
		return "", err
	}

	result, err := db.DB.Exec(fmt.Sprintf("INSERT INTO groups (%s) VALUES (%s);", groupColumns, qQ(len(groupValues))), groupValues...)
	err = getExecError(result, err)
	return g.ID, err
}

// ReadGroup reads a group by id
func (db *AdminDB) ReadGroup(id string, o *ReadGroupOptions) (*Group, error) {
	g := &Group{}
	err := db.Get(g, "SELECT * FROM groups WHERE (id=?) LIMIT 1;", id)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if o != nil && o.Icon {
		g.Icon = nil
	}

	return g, err
}

// UpdateGroup updates the given group (by ID)
func (db *AdminDB) UpdateGroup(g *Group) error {
	groupColumns, groupValues, err := groupUpdateQuery(g)
	if err != nil {
		return err
	}

	groupValues = append(groupValues, g.ID)

	// Allow updating groups that are not users
	result, err := db.Exec(fmt.Sprintf("UPDATE groups SET %s WHERE id=? AND id!=owner;", groupColumns), groupValues...)
	return getExecError(result, err)

}

// DelGroup deletes the given group. It does not permit deleting users.
func (db *AdminDB) DelGroup(id string) error {
	result, err := db.Exec("DELETE FROM groups WHERE id=? AND id!=owner;", id)
	return getExecError(result, err)
}
