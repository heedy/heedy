package dbutil


const dbconversion = `
-- Properties Needed for golang template
-- DBVersion, string, the current DB version or "00000000" if none
-- DBType, string, sqlite3 or postgres
-- DroppingTables, boolean, should we drop old tables?


{{/* These variables need to be defined for all databases */}}


{{ if eq .DBType "sqlite3" }}
-- Turn on fkey support so we don't do shitty things without realizing it
PRAGMA foreign_keys = ON;
{{ end }}


-- {{.DBType}} specific template features
-- Primary Key Expression: {{.pkey_exp}}



{{if eq .DBVersion "00000000"}}

CREATE TABLE PhoneCarrier (
	    id {{.pkey_exp}},
	    name VARCHAR UNIQUE NOT NULL,
	    emaildomain VARCHAR UNIQUE NOT NULL);



CREATE TABLE IF NOT EXISTS Users (
    Id {{.pkey_exp}},
	Name VARCHAR UNIQUE NOT NULL,
	Email VARCHAR UNIQUE NOT NULL,
	Password VARCHAR NOT NULL,
	PasswordSalt VARCHAR NOT NULL,
	PasswordHashScheme VARCHAR NOT NULL,
	Admin BOOLEAN DEFAULT FALSE,
	Phone VARCHAR DEFAULT '',
	PhoneCarrier INTEGER DEFAULT 0,
	UploadLimit_Items INTEGER DEFAULT 24000,
	ProcessingLimit_S INTEGER DEFAULT 86400,
	StorageLimit_Gb INTEGER DEFAULT 4,
	CreateTime INTEGER DEFAULT 0,
	ModifyTime INTEGER DEFAULT 0,
	UserGroup INTEGER DEFAULT 0,
	FOREIGN KEY(PhoneCarrier) REFERENCES PhoneCarrier(Id) ON DELETE SET NULL);

CREATE UNIQUE INDEX UserNameIndex ON Users (Name);


CREATE TABLE IF NOT EXISTS Device (
    Id {{.pkey_exp}},
	Name VARCHAR NOT NULL,
	ApiKey VARCHAR UNIQUE NOT NULL,
	Enabled BOOLEAN DEFAULT TRUE,
	Icon_PngB64 VARCHAR DEFAULT '',
	Shortname VARCHAR DEFAULT '',
	Superdevice BOOL DEFAULT FALSE,
	OwnerId INTEGER,

	CanWrite BOOL DEFAULT TRUE,
	CanWriteAnywhere BOOL DEFAULT TRUE,
	UserProxy BOOL DEFAULT FALSE,

	UNIQUE(Name, OwnerId),
	FOREIGN KEY(OwnerId) REFERENCES Users(Id) ON DELETE CASCADE
	);



CREATE UNIQUE INDEX DeviceNameIndex ON Device (Name);
CREATE UNIQUE INDEX DeviceAPIIndex ON Device (ApiKey);
CREATE INDEX DeviceOwnerIndex ON Device (OwnerId);


CREATE TABLE Stream (
    Id {{.pkey_exp}},
    Name VARCHAR NOT NULL,
    Active BOOLEAN DEFAULT TRUE,
    Public BOOLEAN DEFAULT FALSE,
    Type VARCHAR NOT NULL,
    OwnerId INTEGER,
    Ephemeral BOOL DEFAULT FALSE,
    Output BOOL DEFAULT FALSE,
    UNIQUE(Name, OwnerId),
    FOREIGN KEY(OwnerId) REFERENCES Device(Id) ON DELETE CASCADE
    );


CREATE INDEX StreamNameIndex ON Stream (Name);
CREATE INDEX StreamOwnerIndex ON Stream (OwnerId);


CREATE TABLE IF NOT EXISTS timebatchtable (
    Key VARCHAR NOT NULL,
    EndTime BIGINT,
    EndIndex BIGINT,
	Version INTEGER,
    Data BYTEA,
    UNIQUE (Key, EndIndex),
    PRIMARY KEY (Key, EndIndex)
    );

--Index creation should only be run once.
CREATE INDEX keytime ON timebatchtable (Key,EndTime ASC);

{{end}}



{{if lt .DBVersion "20150328"}}

-- POSTGRES 9.1 should work with If Not Exists


-- This table won't exist for the first one.
CREATE TABLE IF NOT EXISTS StreamdbMeta (
     Key VARCHAR UNIQUE NOT NULL,
     Value VARCHAR NOT NULL);

CREATE INDEX sdb_meta ON StreamdbMeta (Key);

-- If we use this format date, we can just test < using lexocographic comparisons
INSERT INTO StreamdbMeta VALUES ('DBVersion', '20150328');

-- Rename the tables to their temporary alternatives
-- This keeps them from cascading deletes until the end.

ALTER TABLE PhoneCarrier RENAME TO PhoneCarriers;
ALTER TABLE Users RENAME TO Users20150328;
ALTER TABLE Device RENAME TO Devices20150328;
ALTER TABLE Stream RENAME TO Streams20150328;

-- We don't use the INE because we need to fail if these exist

CREATE TABLE Users (
    UserId {{.pkey_exp}},
	Name VARCHAR UNIQUE NOT NULL,
	Email VARCHAR UNIQUE NOT NULL,

	Password VARCHAR NOT NULL,
	PasswordSalt VARCHAR NOT NULL,
	PasswordHashScheme VARCHAR NOT NULL,

	Admin BOOLEAN DEFAULT FALSE,

	UploadLimit_Items INTEGER DEFAULT 24000,
	ProcessingLimit_S INTEGER DEFAULT 86400,
	StorageLimit_Gb INTEGER DEFAULT 4);


CREATE TABLE Devices (
    DeviceId {{.pkey_exp}},
    Name VARCHAR NOT NULL,
    Nickname VARCHAR DEFAULT '',
    UserId INTEGER,
    ApiKey VARCHAR UNIQUE NOT NULL,
    Enabled BOOLEAN DEFAULT TRUE,
    IsAdmin BOOL DEFAULT FALSE,
    CanWrite BOOL DEFAULT TRUE,
    CanWriteAnywhere BOOL DEFAULT FALSE,
    CanActAsUser BOOL DEFAULT FALSE,
    IsVisible BOOL DEFAULT TRUE,
    UserEditable BOOL DEFAULT TRUE,
    UNIQUE(UserId, Name),
    FOREIGN KEY(UserId) REFERENCES Users(UserId) ON DELETE CASCADE);


CREATE TABLE Streams (
    StreamId {{.pkey_exp}},
    Name VARCHAR NOT NULL,
    Nickname VARCHAR NOT NULL DEFAULT '',
    Type VARCHAR NOT NULL,
    DeviceId INTEGER,
    Ephemeral BOOL DEFAULT FALSE,
    Downlink BOOL DEFAULT FALSE,
    UNIQUE(Name, DeviceId),
    FOREIGN KEY(DeviceId) REFERENCES Devices(DeviceId) ON DELETE CASCADE);


CREATE TABLE UserKeyValues (
    UserId INTEGER,
    Key VARCHAR NOT NULL,
    Value VARCHAR NOT NULL DEFAULT '',
    FOREIGN KEY(UserId) REFERENCES Users(UserId) ON DELETE CASCADE,
    UNIQUE(UserId, Key),
    PRIMARY KEY (UserId, Key)
);

CREATE TABLE DeviceKeyValues (
    DeviceId INTEGER,
    Key VARCHAR NOT NULL,
    Value VARCHAR NOT NULL DEFAULT '',
    FOREIGN KEY(DeviceId) REFERENCES Devices(DeviceId) ON DELETE CASCADE,
    UNIQUE(DeviceId, Key),
    PRIMARY KEY (DeviceId, Key)
);


CREATE TABLE StreamKeyValues (
    StreamId INTEGER,
    Key VARCHAR NOT NULL,
    Value VARCHAR NOT NULL DEFAULT '',
    FOREIGN KEY(StreamId) REFERENCES Streams(StreamId) ON DELETE CASCADE,
    UNIQUE(StreamId, Key),
    PRIMARY KEY (StreamId, Key)
);

{{ if eq .DBType "sqlite3"}}
CREATE TRIGGER AddUserdev20150328 AFTER INSERT ON Users FOR EACH ROW
BEGIN
INSERT INTO Devices (Name, UserId, ApiKey, CanActAsUser, UserEditable, IsAdmin) VALUES ('user', NEW.UserId, NEW.Name || '-' || NEW.PasswordSalt, 1, 0, NEW.Admin);
END;
{{end}}

{{ if eq .DBType "postgres"}}

CREATE FUNCTION AddUserdev20150328Func() RETURNS TRIGGER AS $_$
BEGIN
	INSERT INTO Devices (Name, UserId, ApiKey, CanActAsUser, UserEditable, IsAdmin) VALUES ('user', NEW.UserId, NEW.Name || '-' || NEW.PasswordSalt, TRUE, FALSE, NEW.Admin);
    RETURN NEW;
END $_$ LANGUAGE 'plpgsql';


CREATE TRIGGER AddUserdev20150328 AFTER INSERT ON Users FOR EACH ROW
    EXECUTE PROCEDURE AddUserdev20150328Func();

{{end}}

-- Construct Indexes

CREATE UNIQUE INDEX UserNameIndex20150328 ON Users (Name);
CREATE INDEX DeviceNameIndex20150328 ON Devices (Name);
CREATE UNIQUE INDEX DeviceAPIIndex20150328 ON Devices (ApiKey);
CREATE INDEX DeviceOwnerIndex20150328 ON Devices (UserId);
CREATE INDEX StreamNameIndex20150328 ON Streams (Name);
CREATE INDEX StreamOwnerIndex20150328 ON Streams (DeviceId);
CREATE INDEX keytime20150328 ON timebatchtable (Key, EndTime ASC);



-- Transfer Data

INSERT INTO Users SELECT
    Id,
    Name,
    Email,
    Password,
    PasswordSalt,
    PasswordHashScheme,
    Admin,
    UploadLimit_Items,
    ProcessingLimit_S,
    StorageLimit_Gb FROM Users20150328;

INSERT INTO Devices (DeviceId, Name, Nickname,  UserId,  ApiKey, Enabled, IsAdmin,   CanWrite, CanWriteAnywhere, CanActAsUser)
              SELECT Id,       Name, Shortname, OwnerId, ApiKey, Enabled, Superdevice, CanWrite, CanWriteAnywhere, UserProxy FROM Devices20150328;


INSERT INTO Streams (StreamId, Name, Type, DeviceId, Ephemeral, Downlink) SELECT Id, Name, Type, OwnerId, Ephemeral, Output FROM Streams20150328;

-- Insert Data

-- Default Carriers
-- do one by one because old sqlite doesn't like multiple
INSERT INTO PhoneCarriers (name, emaildomain) VALUES ('Alltel US', '@message.alltel.com');
INSERT INTO PhoneCarriers (name, emaildomain) VALUES ('AT&T US', '@txt.att.net');
INSERT INTO PhoneCarriers (name, emaildomain) VALUES ('Nextel US', '@messaging.nextel.com');
INSERT INTO PhoneCarriers (name, emaildomain) VALUES ('T-mobile US', '@tmomail.net');
INSERT INTO PhoneCarriers (name, emaildomain) VALUES ('Verizon US', '@vtext.com');


{{if .DroppingTables}}
-- Comment these out if there's an issue.
DROP TABLE Streams20150328;
DROP TABLE Devices20150328;
DROP TABLE Users20150328;
{{end}}



{{end}}
{{/*end conversion 20150328*/}}
`
