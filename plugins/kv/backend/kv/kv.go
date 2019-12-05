package kv

import (
	"errors"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
)

const SQLVersion = 1

const sqlSchema = `

CREATE TABLE kv_plugin_user (
	user VARCHAR NOT NULL,
	plugin VARCHAR NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR NOT NULL DEFAULT '{}',

	CONSTRAINT pk PRIMARY KEY (plugin,user,key),

	CONSTRAINT valid_value CHECK(json_valid(value)),
	CONSTRAINT user_c
		FOREIGN KEY (user)
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE kv_plugin_app (
	app VARCHAR NOT NULL,
	plugin VARCHAR NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR NOT NULL DEFAULT '{}',

	CONSTRAINT pk PRIMARY KEY (plugin,app,key),

	CONSTRAINT valid_value CHECK(json_valid(value)),
	CONSTRAINT app_c
		FOREIGN KEY (app)
		REFERENCES apps(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE kv_plugin_object (
	object VARCHAR NOT NULL,
	plugin VARCHAR NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR NOT NULL DEFAULT '{}',

	CONSTRAINT pk PRIMARY KEY (plugin,object,key),

	CONSTRAINT valid_value CHECK(json_valid(value)),
	CONSTRAINT object_c
		FOREIGN KEY (object)
		REFERENCES objects(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

`

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion != 0 {
		return errors.New("KV database version too new")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

var Handler = func() *chi.Mux {
	mux := chi.NewMux()
	return mux
}()
