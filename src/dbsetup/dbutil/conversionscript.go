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


-- This table won't exist for the first one.
CREATE TABLE StreamdbMeta (
     Key VARCHAR UNIQUE NOT NULL,
     Value VARCHAR NOT NULL);

CREATE INDEX sdb_meta ON StreamdbMeta (Key);


CREATE TABLE Users (
    UserId {{.pkey_exp}},
	Name VARCHAR UNIQUE NOT NULL,
	Nickname VARCHAR DEFAULT '',
	Email VARCHAR UNIQUE NOT NULL,
    Description VARCHAR(1000) DEFAULT '',
    Icon        VARCHAR(4096) DEFAULT '', -- DATA URI

	Password VARCHAR NOT NULL,
	PasswordSalt VARCHAR NOT NULL,
	PasswordHashScheme VARCHAR NOT NULL,

	Admin BOOLEAN DEFAULT FALSE,

	UploadLimit_Items INTEGER DEFAULT 24000,
	ProcessingLimit_S INTEGER DEFAULT 86400,
	StorageLimit_Gb INTEGER DEFAULT 4);

CREATE UNIQUE INDEX UserNameIndex ON Users (Name);

CREATE TABLE Devices (
    DeviceId {{.pkey_exp}},
    Name VARCHAR NOT NULL,
    Nickname VARCHAR DEFAULT '',
    Description VARCHAR(1000) DEFAULT '',
    Icon        VARCHAR(4096) DEFAULT '', -- DATA URI

    UserId INTEGER,
    ApiKey VARCHAR NOT NULL,
    Enabled BOOLEAN DEFAULT TRUE,
    IsAdmin BOOLEAN DEFAULT FALSE,
    CanWrite BOOLEAN DEFAULT TRUE,
    CanWriteAnywhere BOOLEAN DEFAULT FALSE,
    CanActAsUser BOOLEAN DEFAULT FALSE,
    IsVisible BOOLEAN DEFAULT TRUE,
    UserEditable BOOLEAN DEFAULT TRUE,
    UNIQUE(UserId, Name),
    FOREIGN KEY(UserId) REFERENCES Users(UserId) ON DELETE CASCADE);



CREATE INDEX DeviceNameIndex ON Devices (Name);
CREATE UNIQUE INDEX DeviceAPIIndex ON Devices (ApiKey) WHERE ApiKey!='';
CREATE INDEX DeviceUserIndex ON Devices (UserId);

CREATE TABLE Streams (
    StreamId {{.pkey_exp}},
    Name VARCHAR NOT NULL,
    Nickname VARCHAR NOT NULL DEFAULT '',
    Description VARCHAR(1000) DEFAULT '',
    Icon        VARCHAR(4096) DEFAULT '', -- DATA URI
    Type VARCHAR NOT NULL,
    DeviceId INTEGER,
    Ephemeral BOOLEAN DEFAULT FALSE,
    Downlink BOOLEAN DEFAULT FALSE,
    UNIQUE(Name, DeviceId),
    FOREIGN KEY(DeviceId) REFERENCES Devices(DeviceId) ON DELETE CASCADE);


CREATE INDEX StreamNameIndex ON Streams (Name);
CREATE INDEX StreamDeviceIndex ON Streams (DeviceId);


CREATE TABLE datastream (
	StreamId BIGINT NOT NULL,
	Substream VARCHAR,
	EndTime DOUBLE PRECISION,
	EndIndex BIGINT,
	Version INTEGER,
	Data BYTEA,
	UNIQUE (StreamId, Substream, EndIndex),
	PRIMARY KEY (StreamId, Substream, EndIndex)
);

CREATE INDEX datastreamtime ON datastream (StreamId,Substream,EndTime ASC);

CREATE FUNCTION ModifyUserDeviceFunc() RETURNS TRIGGER AS $_$
BEGIN
	UPDATE Devices SET IsAdmin = NEW.Admin WHERE UserId = NEW.UserId AND IsAdmin = TRUE;
	UPDATE Devices SET IsAdmin = NEW.Admin WHERE UserId = NEW.UserId AND Name = 'user';
    RETURN NEW;
END $_$ LANGUAGE 'plpgsql';


CREATE TRIGGER ModifyUserDevice AFTER UPDATE ON Users FOR EACH ROW
    EXECUTE PROCEDURE ModifyUserDeviceFunc();

CREATE FUNCTION initial_user_setup() RETURNS TRIGGER AS $_$
DECLARE
	var_deviceid INTEGER;
BEGIN
	INSERT INTO Devices (Name, UserId, ApiKey, CanActAsUser, UserEditable, IsAdmin)
	    VALUES ('user', NEW.UserId, NEW.Name || '-' || NEW.PasswordSalt, TRUE, FALSE, NEW.Admin);

	INSERT INTO Devices (Name, UserId, ApiKey, CanActAsUser, UserEditable, IsAdmin) VALUES ('meta', NEW.UserId, '', TRUE, FALSE, FALSE);

	SELECT DeviceId INTO var_deviceid FROM Devices
	    WHERE UserId = NEW.UserId AND Name = 'meta';

	INSERT INTO Streams (Name, Type, DeviceId)
		VALUES ('log',
			'{"type": "object", "properties": {"cmd": {"type": "string"},"arg": {"type": "string"}},"required": ["cmd","arg"]}',
			var_deviceid);

	RETURN NEW;
END $_$ LANGUAGE 'plpgsql';


CREATE TRIGGER initialize_user AFTER INSERT ON Users FOR EACH ROW
    EXECUTE PROCEDURE initial_user_setup();


CREATE FUNCTION permuteQPR(x BIGINT) RETURNS INTEGER AS $$
DECLARE
	prime 	INTEGER;
	residue	BIGINT;
BEGIN
	-- see:
	-- http://preshing.com/20121224/how-to-generate-a-sequence-of-unique-random-integers/
	-- for more information on these calculations
	prime = 2147483423; -- congruence to 3 % 4 holds to ensure 1:1 mapping

	IF x >= prime THEN
		RETURN x; -- last 5 digits map to themselves
	ELSE
		residue = (x * x) % prime;
		IF residue <= prime / 2 THEN
			RETURN residue;
		ELSE
			RETURN (prime - residue);
		END IF;
	END IF;
END
$$ LANGUAGE 'plpgsql';


CREATE FUNCTION id_scramble(id INTEGER) RETURNS INTEGER AS $$
DECLARE
	xor   	INTEGER;
	add   	INTEGER;
BEGIN
	xor   = 0x5bf03635;
	add   = 0xDEADBEEF;
	RETURN permuteQPR(permuteQPR(id) + add) # xor;
END
$$ LANGUAGE 'plpgsql';


CREATE FUNCTION userid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.UserId = id_scramble(NEW.UserId);
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER userid_scramble_trigger
	BEFORE INSERT ON Users
	FOR EACH ROW
	EXECUTE PROCEDURE userid_scramble();


CREATE FUNCTION deviceid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.DeviceId = id_scramble(NEW.DeviceId);
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER deviceid_scramble_trigger
	BEFORE INSERT ON Devices
	FOR EACH ROW
	EXECUTE PROCEDURE deviceid_scramble();


CREATE FUNCTION streamid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.StreamId = id_scramble(NEW.StreamId);
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER streamid_scramble_trigger
	BEFORE INSERT ON Streams
	FOR EACH ROW
	EXECUTE PROCEDURE streamid_scramble();

{{end}}

{{if lt .DBVersion "20150829"}}
-- If we use this format date, we can just test < using lexicographic comparisons
INSERT INTO StreamdbMeta VALUES ('DBVersion', '20150829');
{{end}}
`
