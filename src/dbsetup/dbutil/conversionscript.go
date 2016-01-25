/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package dbutil

const dbconversion = `
-- Properties Needed for golang template
-- DBVersion, string, the current DB version or "00000000" if none
-- DBType, string, postgres
-- Reset, boolean, should we drop old tables?


{{/* These variables need to be defined for all databases */}}



-- {{.DBType}} specific template features
-- Primary Key Expression: {{.pkey_exp}}

{{ if eq .Reset "true"}}
-- Delete the full database
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
DROP SCHEMA v1 CASCADE;
SET search_path = public;

{{end}}

{{if eq .DBVersion "00000000"}}

-- This table won't exist for the first one. It is in the public schema
CREATE TABLE connectordbmeta (
	 Key VARCHAR UNIQUE NOT NULL,
	 Value VARCHAR NOT NULL);

-- The index is also in the public schema
CREATE INDEX cdb_meta ON connectordbmeta (Key);


-- All tables/things are created under the v1 schema, so that they can
-- be manipulated as a unit. This helps upgrading
CREATE SCHEMA v1;

-- Set the schema search path to our current tables - create new tables in v1
ALTER DATABASE connectordb SET search_path = v1,public;
SET search_path = v1,public;

CREATE TABLE Users (
	UserID {{.pkey_exp}},
	Name VARCHAR UNIQUE NOT NULL,
	Nickname VARCHAR DEFAULT '',
	Email VARCHAR UNIQUE NOT NULL,
	Description VARCHAR(1000) DEFAULT '',
	Icon		VARCHAR(4096) DEFAULT '', -- DATA URI

	Public BOOLEAN DEFAULT FALSE,
	Roles VARCHAR NOT NULL,

	Password VARCHAR NOT NULL,
	PasswordSalt VARCHAR NOT NULL,
	PasswordHashScheme VARCHAR NOT NULL);

CREATE UNIQUE INDEX UserNameIndex ON Users (Name);

CREATE TABLE Devices (
	DeviceID {{.pkey_exp}},
	Name VARCHAR NOT NULL,
	Nickname VARCHAR DEFAULT '',
	Description VARCHAR(1000) DEFAULT '',
	Icon		VARCHAR(4096) DEFAULT '', -- DATA URI

	UserID INTEGER,
	APIKey VARCHAR NOT NULL,
	Enabled BOOLEAN DEFAULT TRUE,

	Public BOOLEAN DEFAULT FALSE,

	Roles VARCHAR DEFAULT '',


	IsVisible BOOLEAN DEFAULT TRUE,
	UserEditable BOOLEAN DEFAULT TRUE,
	UNIQUE(UserID, Name),
	FOREIGN KEY(UserID) REFERENCES Users(UserID) ON DELETE CASCADE);



CREATE INDEX DeviceNameIndex ON Devices (Name);
CREATE UNIQUE INDEX DeviceAPIIndex ON Devices (APIKey) WHERE APIKey!='';
CREATE INDEX DeviceUserIndex ON Devices (UserID);

CREATE TABLE Streams (
	StreamID {{.pkey_exp}},
	Name VARCHAR NOT NULL,
	Nickname VARCHAR NOT NULL DEFAULT '',
	Description VARCHAR(1000) DEFAULT '',
	Icon		VARCHAR(4096) DEFAULT '', -- DATA URI
	Schema VARCHAR NOT NULL,
	DeviceID INTEGER,
	Ephemeral BOOLEAN DEFAULT FALSE,
	Downlink BOOLEAN DEFAULT FALSE,
	UNIQUE(Name, DeviceID),
	FOREIGN KEY(DeviceID) REFERENCES Devices(DeviceID) ON DELETE CASCADE);


CREATE INDEX StreamNameIndex ON Streams (Name);
CREATE INDEX StreamDeviceIndex ON Streams (DeviceID);


CREATE TABLE Datastream (
	StreamID BIGINT NOT NULL,
	Substream VARCHAR,
	EndTime DOUBLE PRECISION,
	EndIndex BIGINT,
	Version INTEGER,
	Data BYTEA,
	UNIQUE (StreamID, Substream, EndIndex),
	PRIMARY KEY (StreamID, Substream, EndIndex)
);

CREATE INDEX datastreamtime ON Datastream (StreamID,Substream,EndTime ASC);


-- Create the user and meta Devices for the user when a user is created
CREATE FUNCTION initial_user_setup() RETURNS TRIGGER AS $_$
DECLARE
	var_deviceid INTEGER;
BEGIN
	INSERT INTO Devices (Name, UserID, APIKey,Roles)
		VALUES ('user', NEW.UserID, NEW.PasswordSalt, 'user');

	INSERT INTO Devices (Name, UserID, APIKey, UserEditable, IsVisible) VALUES ('meta', NEW.UserID, '', FALSE, FALSE);

	SELECT DeviceID INTO var_deviceid FROM Devices
		WHERE UserID = NEW.UserID AND Name = 'meta';

	INSERT INTO Streams (Name, Schema, DeviceID)
		VALUES ('log',
			'{"type": "object", "properties": {"cmd": {"type": "string"},"arg": {"type": "string"}},"required": ["cmd","arg"]}',
			var_deviceid);

	RETURN NEW;
END $_$ LANGUAGE 'plpgsql';

CREATE TRIGGER initialize_user AFTER INSERT ON Users FOR EACH ROW
	EXECUTE PROCEDURE initial_user_setup();


-- Set the database version
INSERT INTO public.connectordbmeta VALUES ('DBVersion', '20160120');


-- ID scrambling for use if ids are ever exposed.
-- Note that
CREATE FUNCTION permuteQPR(x BIGINT) RETURNS INTEGER AS $$
DECLARE
	prime 	INTEGER;
	residue	BIGINT;
BEGIN
	-- see:
	-- http://preshing.com/20121224/how-to-generate-a-sequence-of-unique-random-integers/
	-- for more information on these calculations
	prime = {{ .IDScramblePrime }}; -- congruence to 3 % 4 holds to ensure 1:1 mapping

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
	NEW.UserID = id_scramble(NEW.UserID);
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER userid_scramble_trigger
	BEFORE INSERT ON Users
	FOR EACH ROW
	EXECUTE PROCEDURE userid_scramble();


CREATE FUNCTION deviceid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.DeviceID = id_scramble(NEW.DeviceID);
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER deviceid_scramble_trigger
	BEFORE INSERT ON Devices
	FOR EACH ROW
	EXECUTE PROCEDURE deviceid_scramble();


CREATE FUNCTION streamid_scramble()
RETURNS trigger AS $$
BEGIN
	NEW.StreamID = id_scramble(NEW.StreamID);
	RETURN NEW;
END$$ LANGUAGE 'plpgsql';

CREATE TRIGGER streamid_scramble_trigger
	BEFORE INSERT ON Streams
	FOR EACH ROW
	EXECUTE PROCEDURE streamid_scramble();
{{end}}

{{if lt .DBVersion "20160120"}}
-- We can perform upgrade operations here

-- UPDATE connectordbmeta SET Version = '20160120' WHERE Key = 'DBVersion';
{{end}}
`
