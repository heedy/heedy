package run

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/database"
)

type BuiltinStartFunc func(db *database.AdminDB, i *Info) error

// Builtin is passed in to the BuiltinHandler with
type BuiltinRunner struct {
	Key     string
	Start   BuiltinStartFunc
	Stop    func(db *database.AdminDB, apikey string) error
	Handler http.Handler
}

func WithVersion(pluginName string, dbversion int, pstart func(*database.AdminDB, *Info, int) error) BuiltinStartFunc {
	return func(db *database.AdminDB, i *Info) error {
		var curVersion int
		err := db.Get(&curVersion, `SELECT version FROM heedy WHERE name=?`, pluginName)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == sql.ErrNoRows {
			curVersion = 0
		}
		err = pstart(db, i, curVersion)
		if err != nil {
			return err
		}
		if dbversion != curVersion {
			_, err = db.Exec(`INSERT OR REPLACE INTO heedy(name,version) VALUES (?,?)`, pluginName, dbversion)
		}
		return err
	}
}

// WithNilInfo can be used to convert a plugin start func that doesn't require an Info struct
// into a function compatible with database.AddCreateHook.
func WithNilInfo(bis BuiltinStartFunc) func(*database.AdminDB) error {
	return func(db *database.AdminDB) error {
		return bis(db, nil)
	}
}

type builtinRunnerMap map[string]*BuiltinRunner

type ChiClearer struct {
	Handler http.Handler
}

func (cc ChiClearer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cc.Handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, nil)))
}

func (b builtinRunnerMap) Add(r *BuiltinRunner) {
	if r.Handler != nil {
		r.Handler = ChiClearer{r.Handler}
	}
	b[r.Key] = r
}

var Builtin = make(builtinRunnerMap)

type BuiltinHandler struct {
	DB      *database.AdminDB
	Running map[string]string
}

func NewBuiltinHandler(db *database.AdminDB) *BuiltinHandler {
	return &BuiltinHandler{
		DB:      db,
		Running: make(map[string]string),
	}
}

func (bh *BuiltinHandler) Start(i *Info) (h http.Handler, err error) {
	err = bh.Run(i)
	if err == nil {
		// This was validated in Run
		bkey := i.Run.Settings["key"].(string)
		r := Builtin[bkey]
		h = r.Handler

		bh.Running[i.APIKey] = bkey
	}
	return
}

func (bh *BuiltinHandler) Stop(apikey string) error {
	bkey, ok := bh.Running[apikey]
	if !ok {
		return errors.New("The server is not running")
	}
	delete(bh.Running, apikey)
	r := Builtin[bkey]
	if r.Stop == nil {
		return nil
	}
	return r.Stop(bh.DB, apikey)
}

func (bg *BuiltinHandler) Kill(apikey string) error {
	// Builtin stuff can't actually be killed
	return nil
}

func (bh *BuiltinHandler) Run(i *Info) (err error) {
	k, ok := i.Run.Settings["key"]
	if !ok {
		err = errors.New("builtin runtype requires a 'key' attribute")
		return
	}
	bkey, ok := k.(string)
	if !ok {
		err = errors.New("'key' must be a string")
		return
	}

	r, ok := Builtin[bkey]
	if !ok {
		err = fmt.Errorf("builtin key '%s' not recognized", bkey)
		return
	}
	if r.Start != nil {
		err = r.Start(bh.DB, i)
	}

	return
}
