package plugins

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"time"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"

	"github.com/go-chi/chi"
	"github.com/robfig/cron"

	"github.com/sirupsen/logrus"
)

type Plugin struct {
	// The mux that holds the overlay
	Mux *chi.Mux

	// The assets that are used by the server
	Assets *assets.Assets

	// The database
	DB *database.AdminDB

	// Name of the plugin
	Name string

	// Processes holds all the exec programs being handled by heedy.
	// The map keys are their associated api keys
	Processes map[string]*Exec `json:"exec"`

	// The cron daemon to run in the background for cron processes
	cron *cron.Cron

	EventRouter *events.Router
}

func NewPlugin(db *database.AdminDB, a *assets.Assets, pname string) (*Plugin, error) {
	p := &Plugin{
		DB:          db,
		Processes:   make(map[string]*Exec),
		Assets:      a,
		Name:        pname,
		EventRouter: events.NewRouter(),
	}
	logrus.Debugf("Loading plugin '%s'", pname)

	psettings := a.Config.Plugins[pname]

	// Set up API forwards
	if psettings.Routes != nil && len(*psettings.Routes) > 0 {

		mux := chi.NewMux()

		for rname, redirect := range *psettings.Routes {
			revproxy, err := NewReverseProxy(a.DataDir(), redirect)
			if err != nil {
				return nil, err
			}
			logrus.Debugf("%s: Forwarding %s -> %s ", pname, rname, redirect)
			mux.Handle(rname, revproxy)
		}

		p.Mux = mux
	}

	// Set up events
	for ename, ev := range psettings.On {
		peh, err := PluginEventHandler(a, pname, ev)
		if err != nil {
			return nil, err
		}
		logrus.Debugf("%s: Forwarding event '%s' -> %s", pname, ename, *ev.Post)
		p.EventRouter.Subscribe(events.Event{
			Event: ename,
			User:  "*",
		}, peh)
	}
	for cplugin, cv := range psettings.Connections {
		for ename, ev := range cv.On {
			peh, err := PluginEventHandler(a, pname, ev)
			if err != nil {
				return nil, err
			}
			cpn := pname + ":" + cplugin
			logrus.Debugf("%s: Forwarding event '%s/%s' -> %s", pname, cpn, ename, *ev.Post)
			p.EventRouter.Subscribe(events.Event{
				Event:  ename,
				Plugin: &cpn,
			}, peh)
		}
		for skey, sv := range cv.Sources {
			for ename, ev := range sv.On {
				peh, err := PluginEventHandler(a, pname, ev)
				if err != nil {
					return nil, err
				}
				cpn := pname + ":" + cplugin
				logrus.Debugf("%s: Forwarding event '%s/%s/%s' -> %s", pname, cpn, skey, ename, *ev.Post)
				p.EventRouter.Subscribe(events.Event{
					Event:  ename,
					Plugin: &cpn,
					Key:    skey,
				}, peh)
			}
		}
	}
	return p, nil
}

// Start the backend executables
func (p *Plugin) Start() error {
	if len(p.Processes) > 0 {
		return errors.New("Must first stop running processes to restart Plugin")
	}
	p.cron = cron.New()
	p.cron.Start()
	pname := p.Name
	pv := p.Assets.Config.Plugins[pname]

	endpoints := make(map[string]*Exec)
	for ename, ev := range pv.Exec {
		if ev.Enabled == nil || ev.Enabled != nil && *ev.Enabled {
			keepAlive := false
			if ev.KeepAlive != nil {
				keepAlive = *ev.KeepAlive
			}
			if ev.Cmd == nil || len(*ev.Cmd) == 0 {
				p.Stop()
				return fmt.Errorf("%s/%s has empty command", pname, ename)
			}

			// Create an API key for the exec
			apikeybytes := make([]byte, 64)
			_, err := rand.Read(apikeybytes)

			e := &Exec{
				Plugin:    pname,
				Exec:      ename,
				APIKey:    base64.StdEncoding.EncodeToString(apikeybytes),
				Config:    p.Assets.Config,
				RootDir:   p.Assets.FolderPath,
				DataDir:   p.Assets.DataDir(),
				PluginDir: path.Join(p.Assets.PluginDir(), pname),
				keepAlive: keepAlive,
				cmd:       *ev.Cmd, // TODO: Handle Python
			}

			p.Processes[e.APIKey] = e

			if ev.Cron != nil && len(*ev.Cron) > 0 {
				logrus.Debugf("%s: Enabling cron job %s", pname, ename)
				err = p.cron.AddJob(*ev.Cron, e)
			} else {
				logrus.Debugf("%s: Running %s", pname, ename)
				err = e.Start()
			}
			if err != nil {
				p.Stop()
				return err
			}
			if ev.Endpoint != nil {
				endpoints[*ev.Endpoint] = e
			}

		}

	}

	// Now wait until all the endpoints are open
	for ep, e := range endpoints {
		logrus.Debugf("%s: Waiting for endpoint %s", pname, ep)
		method, host, err := GetEndpoint(p.Assets.DataDir(), ep)
		if err != nil {
			p.Stop()
			return err
		}
		if err = WaitForEndpoint(method, host, e); err != nil {
			p.Stop()
			return err
		}
		logrus.Debugf("%s: Endpoint %s open", pname, ep)
	}

	return nil
}

func processConnection(pluginKey string, owner string, cv *assets.Connection) *database.Connection {
	c := &database.Connection{
		Details: database.Details{
			Name:        &cv.Name,
			Description: cv.Description,
			Avatar:      cv.Avatar,
		},
		Enabled: cv.Enabled,
		Plugin:  &pluginKey,
		Owner:   &owner,
	}
	if cv.Scopes != nil {
		c.Scopes = &database.ConnectionScopeArray{
			ScopeArray: database.ScopeArray{
				Scopes: *cv.Scopes,
			},
		}
	}
	if cv.AccessToken == nil || !(*cv.AccessToken) {
		empty := ""
		c.AccessToken = &empty
	}
	if cv.SettingsSchema != nil {
		jo := database.JSONObject(*cv.SettingsSchema)
		c.SettingsSchema = &jo
	}
	if cv.Settings != nil {
		jo := database.JSONObject(*cv.Settings)
		c.Settings = &jo
	}
	if cv.Type != nil {
		c.Type = cv.Type
	}
	return c
}

func processSource(connection string, key string, as *assets.Source) *database.Source {
	s := &database.Source{
		Details: database.Details{
			Name:        &as.Name,
			Description: as.Description,
			Avatar:      as.Avatar,
		},
		Connection: &connection,
		Key:        &key,
		Type:       &as.Type,
	}
	if as.Meta != nil {
		jo := database.JSONObject(*as.Meta)
		s.Meta = &jo
	}
	if as.Scopes != nil {
		s.Scopes = &database.ScopeArray{
			Scopes: *as.Scopes,
		}
	}

	return s
}

func internalRequest(ir InternalRequester, method, path, plugin string, body interface{}) error {
	var bodybuffer io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodybuffer = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, path, bodybuffer)
	if err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	ir.ServeInternal(rec, req, plugin)
	if rec.Code != http.StatusOK {
		var er rest.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &er)
		if err != nil {
			return err
		}
		return &er
	}
	return nil
}

// BeforeStart is run before any of the plugin's executables are run.
// This function is used to check if we're to create connections/sources
// for the plugin
func (p *Plugin) BeforeStart(ir InternalRequester) error {
	psettings := p.Assets.Config.Plugins[p.Name]
	for cname, cv := range psettings.Connections {
		// For each connection
		// Check if the connection exists for all users
		var res []string

		pluginKey := p.Name + ":" + cname

		err := p.DB.DB.Select(&res, "SELECT username FROM users WHERE username NOT IN ('heedy', 'public', 'users') AND NOT EXISTS (SELECT 1 FROM connections WHERE owner=users.username AND connections.plugin=?);", pluginKey)
		if err != nil {
			return err
		}
		if len(res) > 0 {
			logrus.Debugf("%s: Creating '%s' connection for all users", p.Name, pluginKey)

			// aaand how exactly do I achieve this?

			for _, uname := range res {

				_, _, err = p.DB.CreateConnection(processConnection(pluginKey, uname, cv))
				if err != nil {
					return err
				}
			}
		}

		for skey, sv := range cv.Sources {
			if sv.Defer == nil || !*sv.Defer {
				res = []string{}
				err := p.DB.DB.Select(&res, "SELECT id FROM connections WHERE plugin=? AND NOT EXISTS (SELECT 1 FROM sources WHERE connection=connections.id AND key=?);", pluginKey, skey)
				if err != nil {
					return err
				}
				if len(res) > 0 {
					logrus.Debugf("%s: Creating '%s/%s' source for all users", p.Name, pluginKey, skey)

					for _, cid := range res {
						s := processSource(cid, skey, sv)
						err = internalRequest(ir, "POST", "/api/heedy/v1/sources", p.Name, s)
						if err != nil {
							return err
						}
					}
				}
			}

		}
	}
	return nil
}

// AfterStart is used for the same purpose as BeforeStart, but it creates deferred sources/connections.
// It also sets up all event callbacks
func (p *Plugin) AfterStart(ir InternalRequester) error {
	psettings := p.Assets.Config.Plugins[p.Name]
	for cname, cv := range psettings.Connections {
		// For each connection
		// Check if the connection exists for all users
		var res []string

		pluginKey := p.Name + ":" + cname

		for skey, sv := range cv.Sources {
			if sv.Defer != nil && *sv.Defer {
				err := p.DB.DB.Select(&res, "SELECT id FROM connections WHERE plugin=? AND NOT EXISTS (SELECT 1 FROM sources WHERE connection=connections.id AND key=?);", pluginKey, skey)
				if err != nil {
					return err
				}
				if len(res) > 0 {
					logrus.Debugf("%s: Creating '%s/%s' source for all users", p.Name, pluginKey, skey)

					for _, cid := range res {
						s := processSource(cid, skey, sv)
						err = internalRequest(ir, "POST", "/api/heedy/v1/sources", p.Name, s)
						if err != nil {
							return err
						}
					}
				}
			}

		}
	}

	// Finally, attach the event router to the event system
	events.AddHandler(p.EventRouter)
	return nil
}

// GetProcessByKey gets the process associated with a given API key
func (p *Plugin) GetProcessByKey(key string) (*Exec, error) {
	v, ok := p.Processes[key]
	if ok {
		return v, nil
	}
	return nil, errors.New("No such key")
}

// Interrupt signals all processes to stop
func (p *Plugin) Interrupt() error {
	p.cron.Stop()
	for _, e := range p.Processes {
		e.Interrupt()
	}
	return nil
}

func (p *Plugin) AnyRunning() bool {
	anyrunning := false
	for _, e := range p.Processes {
		if e.IsRunning() {
			anyrunning = true
		}
	}

	return anyrunning
}

func (p *Plugin) HasProcess() bool {
	return len(p.Processes) == 0
}

// Kill kills all processes
func (p *Plugin) Kill() {
	for _, e := range p.Processes {
		if e.IsRunning() {
			logrus.Warnf("%s: Killing %s", e.Exec)
			e.Kill()
		}
	}
}

func (p *Plugin) Stop() error {
	p.Interrupt()

	d := assets.Get().Config.GetExecTimeout()

	sleepDuration := 50 * time.Millisecond

	for i := time.Duration(0); i < d; i += sleepDuration {
		if !p.AnyRunning() {
			return nil
		}
		time.Sleep(sleepDuration)
	}

	p.Kill()
	return nil
}

func (p *Plugin) Close() error {
	events.RemoveHandler(p.EventRouter)
	if p.AnyRunning() {
		p.Stop()
	}
	return nil
}
