-- Properties Needed for golang template
-- DBVersion, string, the current DB version or "00000000" if none
-- DBType, string, sqlite3 or postgres
-- DroppingTables, boolean, should we drop old tables?


{/* These variables need to be defined for all databases */}


{{ if eq .DBType "postgres" }}
    {{$pkey_exp := "SERIAL PRIMARY KEY"}}
{{ end }}

{{ if eq .DBType "sqlite3" }}
    {{$pkey_exp := "INTEGER PRIMARY KEY"}}

-- Turn on fkey support so we don't do shitty things without realizing it
PRAGMA foreign_keys = ON;
{{ end }}


-- {{.DBType}} specific template features
-- Primary Key Expression: {{$pkey_exp}}



{{if eq .DBVersion "init"}}

CREATE TABLE PhoneCarrier (
	    id {{$pkey_exp}},
	    name CHAR(100) UNIQUE NOT NULL,
	    emaildomain CHAR(50) UNIQUE NOT NULL);



CREATE TABLE IF NOT EXISTS Users (
    Id {{$pkey_exp}},
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
    Id INTEGER {{$pkey_exp}},
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

	FOREIGN KEY(OwnerId) REFERENCES Users(Id) ON DELETE CASCADE,
	UNIQUE(Name, OwnerId)
	);



CREATE UNIQUE INDEX DeviceNameIndex ON Device (Name);
CREATE UNIQUE INDEX DeviceAPIIndex ON Device (ApiKey);
CREATE INDEX DeviceOwnerIndex ON Device (OwnerId);


CREATE TABLE Stream (
    Id INTEGER PRIMARY KEY,
    Name VARCHAR NOT NULL,
    Active BOOLEAN DEFAULT TRUE,
    Public BOOLEAN DEFAULT FALSE,
    Type VARCHAR NOT NULL,
    OwnerId INTEGER,
    Ephemeral BOOL DEFAULT FALSE,
    Output BOOL DEFAULT FALSE,
    FOREIGN KEY(OwnerId) REFERENCES Device(Id) ON DELETE CASCADE,
    UNIQUE(Name, OwnerId)
    );


CREATE INDEX StreamNameIndex ON Stream (Name);
CREATE INDEX StreamOwnerIndex ON Stream (OwnerId);


CREATE TABLE IF NOT EXISTS timebatchtable (
    Key VARCHAR NOT NULL,
    EndTime BIGINT,
    EndIndex BIGINT,
    Data BYTEA,
    PRIMARY KEY (Key, EndIndex)
    );

--Index creation should only be run once.
CREATE INDEX keytime ON timebatchtable (Key,EndTime ASC);

{{end}}



{{if .DBVersion lt "20150328"}}

-- POSTGRES 9.1 should work with If Not Exists


-- This table won't exist for the first one.
CREATE TABLE IF NOT EXISTS StreamdbMeta (
     Key VARCHAR UNIQUE NOT NULL,
     Value VARCHAR NOT NULL);

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
    UserId INTEGER PRIMARY KEY,
	Name VARCHAR UNIQUE NOT NULL,
	Email VARCHAR UNIQUE NOT NULL,

	Password VARCHAR NOT NULL,
	PasswordSalt VARCHAR NOT NULL,
	PasswordHashScheme INTEGER NOT NULL,

	Admin BOOLEAN DEFAULT FALSE,

	UploadLimit_Items INTEGER DEFAULT 24000,
	ProcessingLimit_S INTEGER DEFAULT 86400,
	StorageLimit_Gb INTEGER DEFAULT 4);


CREATE TABLE Devices (
    DeviceId INTEGER PRIMARY KEY,
    Name STRING NOT NULL,
    Nickname STRING DEFAULT '',
    UserId INTEGER,
    ApiKey STRING UNIQUE NOT NULL,
    Enabled BOOLEAN DEFAULT TRUE,
    IsAdmin BOOL DEFAULT FALSE,
    CanWrite BOOL DEFAULT TRUE,
    CanWriteAnywhere BOOL DEFAULT TRUE,
    CanActAsUser BOOL DEFAULT FALSE,
    IsVisible BOOL DEFAULT TRUE,
    UserEditable BOOL DEFAULT TRUE,
    FOREIGN KEY(UserId) REFERENCES Users(UserId) ON DELETE CASCADE,
    UNIQUE(Name, UserId));


CREATE TABLE Streams (
    StreamId {{.pkey_expression}},
    Name STRING NOT NULL,
    Nickname VARCHAR NOT NULL DEFAULT '',
    Type VARCHAR NOT NULL,
    DeviceId INTEGER,
    Ephemeral BOOL DEFAULT FALSE,
    Downlink BOOL DEFAULT FALSE,
    FOREIGN KEY(DeviceId) REFERENCES Devices(DeviceId) ON DELETE CASCADE,
    UNIQUE(Name, DeviceId));

-- Construct Indexes

-- Transfer Data

INSERT INTO Users VALUES SELECT
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
              SELECT Id,       Name, Shortname, OwnerId, ApiKey, Enabled, Superuser, CanWrite, CanWriteAnywhere, UserProxy FROM Devices20150328;


INSERT INTO Streams (StreamId, Name, Type, DeviceId, Ephemeral, Downlink) SELECT Id, Name, Type, OwnerId, Ephemeral, Output FROM Streams20150328;

-- Insert Data

-- Default Carriers
INSERT INTO PhoneCarrier (name, emaildomain) VALUES
    ('Alltel US', '@message.alltel.com'),
    ('AT&T US', '@txt.att.net'),
    ('Nextel US', '@messaging.nextel.com'),
    ('T-mobile US', '@tmomail.net'),
    ('Verizon US', '@vtext.com');


{{if .DroppingTables}}
-- Comment these out if there's an issue.
DROP TABLE Streams20150328;
DROP TABLE Devices20150328;
DROP TABLE Users20150328;
{{end}}

{{end}} {/*end conversion 20150328*/}
