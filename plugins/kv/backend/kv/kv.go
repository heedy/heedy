package kv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
)

const SQLVersion = 1

const sqlSchema = `

CREATE TABLE kv_user (
	user VARCHAR NOT NULL,

	namespace VARCHAR NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR NOT NULL DEFAULT 'null',

	CONSTRAINT pk PRIMARY KEY (user,namespace,key),
	CONSTRAINT valid_value CHECK(json_valid(value)),

	CONSTRAINT fk
		FOREIGN KEY (user)
		REFERENCES users(username)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE kv_app (
	app VARCHAR NOT NULL,

	namespace VARCHAR NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR NOT NULL DEFAULT 'null',

	CONSTRAINT pk PRIMARY KEY (app,namespace,key),
	CONSTRAINT valid_value CHECK(json_valid(value)),

	CONSTRAINT fk
		FOREIGN KEY (app)
		REFERENCES apps(id)
		ON UPDATE CASCADE
		ON DELETE CASCADE
);

CREATE TABLE kv_object (
	object VARCHAR NOT NULL,

	namespace VARCHAR NOT NULL,
	key VARCHAR NOT NULL,
	value VARCHAR NOT NULL DEFAULT 'null',

	CONSTRAINT pk PRIMARY KEY (object,namespace,key),
	CONSTRAINT valid_value CHECK(json_valid(value)),

	CONSTRAINT fk
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

type KV interface {
	Get() (map[string]interface{}, error)
	Set(data map[string]interface{}) error
	Update(data map[string]interface{}) error

	SetKey(key string, value interface{}) error
	GetKey(key string) (interface{}, error)
	DelKey(key string) error
}

func UserAuth(ctx *rest.Context, username string, namespace string) (KV, error) {
	if ctx.DB.Type() != database.AdminType {
		return nil, errors.New("access_denied: KV is currently only available to plugins.")
	}
	return &AdminUserKV{
		DB:        ctx.DB.AdminDB(),
		ID:        username,
		Namespace: namespace,
	}, nil
}

func AppAuth(ctx *rest.Context, appid string, namespace string) (KV, error) {
	if ctx.DB.Type() != database.AdminType {
		if ctx.DB.Type() == database.AppType {
			dbid := strings.Split(ctx.DB.ID(), "/")[1]
			if (appid == dbid || appid == "self") && (dbid == namespace || namespace == "self") {
				// TODO: This should be a scope, but for now, we just include it as a temporary hack to allow testing KV from python code without setting up a plugin
				return &AdminAppKV{
					DB:        ctx.DB.AdminDB(),
					ID:        dbid,
					Namespace: dbid,
				}, nil
			}
			return nil, errors.New("access_denied: apps don't currently support KV in other apps.")
		}

		return nil, errors.New("access_denied: KV is currently only available to plugins.")
	}
	return &AdminAppKV{
		DB:        ctx.DB.AdminDB(),
		ID:        appid,
		Namespace: namespace,
	}, nil
}

func ObjectAuth(ctx *rest.Context, oid string, namespace string) (KV, error) {
	if ctx.DB.Type() != database.AdminType {
		return nil, errors.New("access_denied: KV is currently only available to plugins.")
	}
	return &AdminObjectKV{
		DB:        ctx.DB.AdminDB(),
		ID:        oid,
		Namespace: namespace,
	}, nil
}

func getKV(adb *database.AdminDB, selectStatement string, args ...interface{}) (map[string]interface{}, error) {
	var res []struct {
		Key   string
		Value string
	}

	err := adb.Select(&res, selectStatement, args...)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for _, resv := range res {
		var v interface{}
		err = json.Unmarshal([]byte(resv.Value), &v)
		if err != nil {
			return nil, err
		}

		m[resv.Key] = v
	}

	return m, err

}

func getKVr(w http.ResponseWriter, r *http.Request, selectStatement string, args ...interface{}) {
	ctx := rest.CTX(r)
	m, err := getKV(ctx.DB.AdminDB(), selectStatement, args...)
	rest.WriteJSON(w, r, m, err)
}

func setKV(adb *database.AdminDB, data map[string]interface{}, deleteStatement string, deleteArgs []interface{}, setKeyStatement string, args ...interface{}) error {
	tx, err := adb.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(deleteStatement, deleteArgs...)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Add the key and value elements
	kindex := len(args)
	vindex := kindex + 1
	args = append(args, "", "")
	for k, v := range data {
		args[kindex] = k
		b, err := json.Marshal(v)
		if err != nil {
			tx.Rollback()
			return err
		}
		args[vindex] = string(b)
		res, err := tx.Exec(setKeyStatement, args...)
		err = database.GetExecError(res, err)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func updateKV(adb *database.AdminDB, data map[string]interface{}, setKeyStatement string, args ...interface{}) error {
	tx, err := adb.Beginx()
	if err != nil {
		return err
	}

	// Add the key and value elements
	kindex := len(args)
	vindex := kindex + 1
	args = append(args, "", "")
	for k, v := range data {
		args[kindex] = k
		b, err := json.Marshal(v)
		if err != nil {
			tx.Rollback()
			return err
		}
		args[vindex] = string(b)
		res, err := tx.Exec(setKeyStatement, args...)
		err = database.GetExecError(res, err)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func getKey(adb *database.AdminDB, selectStatement string, args ...interface{}) (interface{}, error) {
	var sv string
	err := adb.Get(&sv, selectStatement, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No rows just means that it is null
		}
		return nil, err
	}
	var v interface{}
	err = json.Unmarshal([]byte(sv), &v)
	return v, err
}

// The final argument of args in marshalled into json
func setKey(adb *database.AdminDB, key string, value interface{}, insertStatement string, args ...interface{}) error {

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	args = append(args, key, string(b))
	return runStatement(adb, insertStatement, args)
}

func runStatement(adb *database.AdminDB, statement string, args ...interface{}) error {
	res, err := adb.Exec(statement, args...)
	return database.GetExecError(res, err)
}
