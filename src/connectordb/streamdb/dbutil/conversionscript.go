package dbutil

const dbconversion = `
-- Properties Needed for golang template
-- DBVersion, string, the current DB version or "00000000" if none
-- DBType, string, postgres
-- DroppingTables, boolean, should we drop old tables?


{{/* These variables need to be defined for all databases */}}



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
	Superdevice BOOLEAN DEFAULT FALSE,
	OwnerId INTEGER,

	CanWrite BOOLEAN DEFAULT TRUE,
	CanWriteAnywhere BOOLEAN DEFAULT TRUE,
	UserProxy BOOLEAN DEFAULT FALSE,

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
    Ephemeral BOOLEAN DEFAULT FALSE,
    Output BOOLEAN DEFAULT FALSE,
    UNIQUE(Name, OwnerId),
    FOREIGN KEY(OwnerId) REFERENCES Device(Id) ON DELETE CASCADE
    );


CREATE INDEX StreamNameIndex ON Stream (Name);
CREATE INDEX StreamOwnerIndex ON Stream (OwnerId);


CREATE TABLE IF NOT EXISTS datastream (
	StreamId BIGINT NOT NULL,
	Substream VARCHAR,
	EndTime DOUBLE PRECISION,
	EndIndex BIGINT,
	Version INTEGER,
	Data BYTEA,
	UNIQUE (StreamId, Substream, EndIndex),
	PRIMARY KEY (StreamId, Substream, EndIndex)
);

--Index creation should only be run once.
CREATE INDEX datastreamtime ON datastream (StreamId,Substream,EndTime ASC);

{{end}}

{{/*========================================================================*/}}
{{/*Changes: updates to all tables in the database for newer schemas*/}}
{{/*========================================================================*/}}


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
    IsAdmin BOOLEAN DEFAULT FALSE,
    CanWrite BOOLEAN DEFAULT TRUE,
    CanWriteAnywhere BOOLEAN DEFAULT FALSE,
    CanActAsUser BOOLEAN DEFAULT FALSE,
    IsVisible BOOLEAN DEFAULT TRUE,
    UserEditable BOOLEAN DEFAULT TRUE,
    UNIQUE(UserId, Name),
    FOREIGN KEY(UserId) REFERENCES Users(UserId) ON DELETE CASCADE);


CREATE TABLE Streams (
    StreamId {{.pkey_exp}},
    Name VARCHAR NOT NULL,
    Nickname VARCHAR NOT NULL DEFAULT '',
    Type VARCHAR NOT NULL,
    DeviceId INTEGER,
    Ephemeral BOOLEAN DEFAULT FALSE,
    Downlink BOOLEAN DEFAULT FALSE,
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

{{ if eq .DBType "postgres"}}

CREATE FUNCTION AddUserdev20150328Func() RETURNS TRIGGER AS $_$
BEGIN
	INSERT INTO Devices (Name, UserId, ApiKey, CanActAsUser, UserEditable, IsAdmin) VALUES ('user', NEW.UserId, NEW.Name || '-' || NEW.PasswordSalt, TRUE, FALSE, NEW.Admin);
    RETURN NEW;
END $_$ LANGUAGE 'plpgsql';


CREATE TRIGGER AddUserdev20150328 AFTER INSERT ON Users FOR EACH ROW
    EXECUTE PROCEDURE AddUserdev20150328Func();




CREATE FUNCTION ModifyUserdev20150328Func() RETURNS TRIGGER AS $_$
BEGIN
	UPDATE Devices SET IsAdmin = NEW.Admin WHERE UserId = NEW.UserId AND IsAdmin = TRUE;
	UPDATE Devices SET IsAdmin = NEW.Admin WHERE UserId = NEW.UserId AND Name = 'user';
    RETURN NEW;
END $_$ LANGUAGE 'plpgsql';


CREATE TRIGGER ModifyUserdev20150328 AFTER UPDATE ON Users FOR EACH ROW
    EXECUTE PROCEDURE ModifyUserdev20150328Func();


{{end}}


-- Construct Indexes

CREATE UNIQUE INDEX UserNameIndex20150328 ON Users (Name);
CREATE INDEX DeviceNameIndex20150328 ON Devices (Name);
CREATE UNIQUE INDEX DeviceAPIIndex20150328 ON Devices (ApiKey);
CREATE INDEX DeviceOwnerIndex20150328 ON Devices (UserId);
CREATE INDEX StreamNameIndex20150328 ON Streams (Name);
CREATE INDEX StreamOwnerIndex20150328 ON Streams (DeviceId);

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
-- do one by one because some databases don't like multiple
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



{{/*========================================================================

Changelog: 2015062826

* drop the unused phone carriers table
* drop the key/value tables (we don't want to be storing random people's data)
* create a log stream when the user is created
* create a table for defunct stream ids. These need to be manually scanned and
  deleted from redis.

========================================================================*/}}

{{if lt .DBVersion "2015062826"}}

UPDATE StreamdbMeta SET Value = '2015062826' WHERE Key = 'DBVersion';

DROP TABLE PhoneCarriers;
DROP TABLE StreamKeyValues;
DROP TABLE UserKeyValues;
DROP TABLE DeviceKeyValues;

-- Holds deleted streams
CREATE TABLE DeletedStreamIds (id INTEGER);

-- Updates the DeletedStreamIds with the deleted stream id
CREATE FUNCTION stream_deleted() RETURNS TRIGGER AS $_$
BEGIN
	INSERT INTO DeletedStreamIds VALUES (OLD.StreamId);
    RETURN OLD;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER StreamDeleteTrigger
    AFTER DELETE ON Streams
    FOR EACH ROW
    EXECUTE PROCEDURE stream_deleted();

-- Create a new trigger for inserting a device that is for a user.

DROP TRIGGER AddUserdev20150328 ON Users;

CREATE FUNCTION initial_user_setup() RETURNS TRIGGER AS $_$
DECLARE
	var_deviceid INTEGER;
BEGIN
	INSERT INTO Devices (Name, UserId, ApiKey, CanActAsUser, UserEditable, IsAdmin)
	    VALUES ('user', NEW.UserId, NEW.Name || '-' || NEW.PasswordSalt, TRUE, FALSE, NEW.Admin);

	SELECT DeviceId INTO var_deviceid FROM Devices
	    WHERE UserId = NEW.UserId AND Name = 'user';

	INSERT INTO Streams (Name, Type, DeviceId)
		VALUES ('log',
			'{"type": "object", "properties": {"cmd": {"type": "string"},"arg": {"type": "string"}},"required": ["cmd","arg"]}',
			var_deviceid);

	RETURN NEW;
END $_$ LANGUAGE 'plpgsql';


CREATE TRIGGER initialize_user AFTER INSERT ON Users FOR EACH ROW
    EXECUTE PROCEDURE initial_user_setup();

-- Now update the old user Devices

INSERT INTO Streams (Name, Type, DeviceId)
    SELECT 'log' as "Name",
		'{"type": "object", "properties": {"cmd": {"type": "string"},"arg": {"type": "string"}},"required": ["cmd","arg"]}' as "Type",
		DeviceId FROM Devices d WHERE d.Name = 'user';

-- Update users to add a nickname
ALTER TABLE Users
	ADD COLUMN nickname VARCHAR;

-- Update tables to use a random id

CREATE FUNCTION userid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.UserId = NEW.UserId * 928559 % 4294967296;
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER userid_scramble_trigger
	BEFORE INSERT ON Users
	FOR EACH ROW
	EXECUTE PROCEDURE userid_scramble();


CREATE FUNCTION deviceid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.DeviceId = NEW.DeviceId * 928553 % 4294967296;
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER deviceid_scramble_trigger
	BEFORE INSERT ON Devices
	FOR EACH ROW
	EXECUTE PROCEDURE deviceid_scramble();


CREATE FUNCTION streamid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.StreamId = NEW.StreamId * 928521 % 4294967296;
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER streamid_scramble_trigger
	BEFORE INSERT ON Streams
	FOR EACH ROW
	EXECUTE PROCEDURE streamid_scramble();


{{/* End 2015062826 */}}
{{end}}

`
