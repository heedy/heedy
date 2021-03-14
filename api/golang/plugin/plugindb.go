package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/schema"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
)

var ErrUnimplemented = errors.New("unimplemented")
var queryEncoder = schema.NewEncoder()

type PluginDB struct {
	P       *Plugin
	Entity  string
	Overlay int

	RequestID string

	client http.Client
}

func (db *PluginDB) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	host := db.P.Meta.Config.GetHost()
	if host == "" {
		host = "localhost"
	}
	host = "http://" + host + ":" + strconv.Itoa(int(db.P.Meta.Config.GetPort())) + path

	r, err := http.NewRequest(method, host, body)
	if err == nil {
		r.Header.Add("X-Heedy-As", db.Entity)
		r.Header.Add("X-Heedy-Key", db.P.Meta.APIKey)
		r.Header.Add("X-Heedy-Overlay", strconv.Itoa(db.Overlay))
		if db.RequestID != "" {
			r.Header.Add("X-Heedy-ID", db.RequestID)
		}

	}
	return r, err
}

// BasicRequest runs a basic query, and does not return the body unless there was an error
func (db *PluginDB) BasicRequest(method, api string, body io.Reader) error {
	r, err := db.NewRequest(method, api, body)
	if err != nil {
		return err
	}
	resp, err := db.client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		return nil
	}

	// Error
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// The response is an error, so unmarshal into the error struct
	var eresp rest.ErrorResponse
	err = json.Unmarshal(b, &eresp)
	if err != nil {
		return err
	}
	return &eresp

}

func (db *PluginDB) UnmarshalRequest(obj interface{}, method, api string, body io.Reader) error {
	r, err := db.NewRequest(method, api, body)
	if err != nil {
		return err
	}
	resp, err := db.client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		// The response is an error, so unmarshal into the error struct
		var eresp rest.ErrorResponse
		err = json.Unmarshal(b, &eresp)
		if err != nil {
			return err
		}
		return &eresp
	}

	// Unmarshal the result
	return json.Unmarshal(b, obj)
}

func (db *PluginDB) StringRequest(method, api string, body io.Reader) (string, error) {
	r, err := db.NewRequest(method, api, body)
	if err != nil {
		return "", err
	}
	resp, err := db.client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		// The response is an error, so unmarshal into the error struct
		var eresp rest.ErrorResponse
		err = json.Unmarshal(b, &eresp)
		if err != nil {
			return "", err
		}
		return "", &eresp
	}
	return string(b), nil
}

// Fire allows PluginDB to conform to the events.Handler interface, which is used to fire events
func (db *PluginDB) Fire(e *events.Event) {
	api := "/api/events"
	b, err := json.Marshal(e)
	if err != nil {
		db.P.Logger().Warnf("Failed to fire event: %s", err.Error())
		return
	}

	err = db.BasicRequest("POST", api, bytes.NewBuffer(b))
	if err != nil {
		db.P.Logger().Warnf("Failed to fire event: %s", err.Error())
	}
}

func (db *PluginDB) AdminDB() *database.AdminDB {
	adb, err := db.P.AdminDB()
	if err != nil {
		db.P.Logger().Errorf("Could not open AdminDB: %s", err.Error())
		return nil
	}
	return adb
}

func (db *PluginDB) ID() string {
	return db.Entity
}

func (db *PluginDB) Type() database.DBType {
	if db.Entity == "heedy" {
		return database.AdminType
	}
	if db.Entity == "public" {
		return database.PublicType
	}
	i := strings.Index(db.Entity, "/")
	if i > -1 {
		return database.AppType
	}
	return database.UserType
}

func (db *PluginDB) CreateUser(u *database.User) error {
	return database.ErrBadQuery("Can't create users through the REST API")
}

func (db *PluginDB) ReadUser(name string, o *database.ReadUserOptions) (*database.User, error) {
	api := fmt.Sprintf("/api/users/%s", url.PathEscape(name))

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	var u database.User

	err := db.UnmarshalRequest(&u, "GET", api, nil)
	return &u, err
}
func (db *PluginDB) UpdateUser(u *database.User) error {
	api := fmt.Sprintf("/api/users/%s", url.PathEscape(u.ID))
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) DelUser(name string) error {
	api := fmt.Sprintf("/api/users/%s", url.PathEscape(name))
	return db.BasicRequest("DELETE", api, nil)
}

func (db *PluginDB) ListUsers(o *database.ListUsersOptions) ([]*database.User, error) {
	var sl []*database.User
	api := "/api/users"

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	err := db.UnmarshalRequest(&sl, "GET", api, nil)
	return sl, err
}

func (db *PluginDB) CanCreateObject(s *database.Object) error {
	return ErrUnimplemented
}
func (db *PluginDB) CreateObject(s *database.Object) (string, error) {
	api := "/api/objects"
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	err = db.UnmarshalRequest(&s, "POST", api, bytes.NewBuffer(b))
	return s.ID, err
}
func (db *PluginDB) ReadObject(id string, o *database.ReadObjectOptions) (*database.Object, error) {
	api := fmt.Sprintf("/api/objects/%s", url.PathEscape(id))

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	var s database.Object

	err := db.UnmarshalRequest(&s, "GET", api, nil)
	return &s, err
}
func (db *PluginDB) UpdateObject(s *database.Object) error {
	api := fmt.Sprintf("/api/objects/%s", url.PathEscape(s.ID))
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) DelObject(id string) error {
	api := fmt.Sprintf("/api/objects/%s", url.PathEscape(id))
	return db.BasicRequest("DELETE", api, nil)
}

func (db *PluginDB) ShareObject(objectid, userid string, sa *database.ScopeArray) error {
	return ErrUnimplemented
}
func (db *PluginDB) UnshareObjectFromUser(objectid, userid string) error {
	return ErrUnimplemented
}
func (db *PluginDB) UnshareObject(objectid string) error {
	return ErrUnimplemented
}
func (db *PluginDB) GetObjectShares(objectid string) (m map[string]*database.ScopeArray, err error) {
	return nil, ErrUnimplemented
}

// ListObjects lists the given objects
func (db *PluginDB) ListObjects(o *database.ListObjectsOptions) ([]*database.Object, error) {
	var sl []*database.Object
	api := "/api/objects"

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	err := db.UnmarshalRequest(&sl, "GET", api, nil)
	return sl, err
}

func (db *PluginDB) CreateApp(c *database.App) (string, string, error) {
	api := "/api/apps"
	b, err := json.Marshal(c)
	if err != nil {
		return "", "", err
	}

	err = db.UnmarshalRequest(&c, "POST", api, bytes.NewBuffer(b))
	accessToken := ""
	if c.AccessToken != nil {
		accessToken = *c.AccessToken
	}
	return c.ID, accessToken, err
}
func (db *PluginDB) ReadApp(id string, o *database.ReadAppOptions) (*database.App, error) {
	api := fmt.Sprintf("/api/apps/%s", url.PathEscape(id))

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	var c database.App

	err := db.UnmarshalRequest(&c, "GET", api, nil)
	return &c, err
}
func (db *PluginDB) UpdateApp(c *database.App) error {
	api := fmt.Sprintf("/api/apps/%s", c.ID)
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) DelApp(id string) error {
	api := fmt.Sprintf("/api/apps/%s", url.PathEscape(id))
	return db.BasicRequest("DELETE", api, nil)
}
func (db *PluginDB) ListApps(o *database.ListAppOptions) ([]*database.App, error) {
	var cl []*database.App
	api := "/api/apps"

	if o != nil {
		form := url.Values{}
		queryEncoder.Encode(o, form)
		api = api + "?" + form.Encode()
	}
	err := db.UnmarshalRequest(&cl, "GET", api, nil)
	return cl, err
}

func (db *PluginDB) ReadUserSettings(username string) (v map[string]map[string]interface{}, err error) {
	api := fmt.Sprintf("/api/users/%s/settings", url.PathEscape(username))

	err = db.UnmarshalRequest(&v, "GET", api, nil)
	return
}
func (db *PluginDB) UpdateUserPluginSettings(username string, plugin string, settings map[string]interface{}) error {
	api := fmt.Sprintf("/api/users/%s/settings/%s", url.PathEscape(username), url.PathEscape(plugin))
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return db.BasicRequest("PATCH", api, bytes.NewBuffer(b))
}
func (db *PluginDB) ReadUserPluginSettings(username string, plugin string) (v map[string]interface{}, err error) {
	api := fmt.Sprintf("/api/users/%s/settings/%s", url.PathEscape(username), url.PathEscape(plugin))

	err = db.UnmarshalRequest(&v, "GET", api, nil)
	return
}

func (db *PluginDB) ListUserSessions(username string) (v []database.UserSession, err error) {
	api := fmt.Sprintf("/api/users/%s/sessions", url.PathEscape(username))

	err = db.UnmarshalRequest(&v, "GET", api, nil)
	return
}
func (db *PluginDB) DelUserSession(username, sessionid string) error {
	api := fmt.Sprintf("/api/users/%s/sessions/%s", url.PathEscape(username), url.PathEscape(sessionid))
	return db.BasicRequest("DELETE", api, nil)
}
