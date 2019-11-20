package run

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
)

type Info struct {
	Plugin    string                `json:"plugin"`
	Name      string                `json:"name"`
	APIKey    string                `json:"apikey"`
	Run       *assets.Run           `json:"run"`
	HeedyDir  string                `json:"heedy_dir"`
	DataDir   string                `json:"data_dir"`
	PluginDir string                `json:"plugin_dir"`
	Config    *assets.Configuration `json:"config"`
}

type TypeHandler interface {
	// Start performs initialization of the runner
	Start(*Info) (http.Handler, error)
	Run(*Info) error
	Stop(apikey string) error
	Kill(apikey string) error
}

type Runner struct {
	I       *Info
	Handler http.Handler

	m   *Manager
	cid cron.EntryID
}

func (r *Runner) Run() {
	logrus.Debugf("%s: Running cron job %s", r.I.Plugin, r.I.Name)

	rt := r.m.RunTypes[*r.I.Run.Type]
	err := rt.Run(r.I)
	if err != nil {
		logrus.Errorf("%s:%s %w", err)
		return
	}
}

type Manager struct {
	sync.RWMutex

	DB       *database.AdminDB
	RunTypes map[string]TypeHandler

	Runners map[string]*Runner

	// The APIKey that represents the "core" heedy server
	CoreKey string

	cron *cron.Cron
}

func NewManager(db *database.AdminDB) *Manager {
	runtypes := make(map[string]TypeHandler)

	// Add the run types built into the database
	runtypes["builtin"] = NewBuiltinHandler(db)
	runtypes["exec"] = NewExecHandler(db)

	c := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.PrintfLogger(logrus.StandardLogger()))))
	c.Start()

	runners := make(map[string]*Runner)

	// Create the heedy internal runner, which represents the heedy core server.
	// This allows the server to make requests as if it were a plugin. This is useful
	// for when the server is creating objects (which are implemented entirely by plugins)

	apikeybytes := make([]byte, 128)
	_, err := rand.Read(apikeybytes)
	if err != nil {
		panic(err)
	}
	a := db.Assets()

	apikey := base64.StdEncoding.EncodeToString(apikeybytes)
	runners[apikey] = &Runner{
		I: &Info{
			Plugin:    "heedy",
			Name:      "core",
			APIKey:    apikey,
			HeedyDir:  a.FolderPath,
			DataDir:   a.DataDir(),
			PluginDir: a.FolderPath,
			Config:    a.Config,
		},
	}

	return &Manager{
		RunTypes: runtypes,
		Runners:  runners,
		cron:     c,
		CoreKey:  apikey,
		DB:       db,
	}
}

func (m *Manager) Start(plugin, name string, run *assets.Run) error {
	if run.Enabled != nil && !*run.Enabled {
		return nil
	}
	if run.Type == nil {
		rtp := "exec"
		run.Type = &rtp
	}
	m.RLock()
	rt, ok := m.RunTypes[*run.Type]
	m.RUnlock()
	if !ok {
		return fmt.Errorf("runtype '%s' not recognized", run.Type)
	}

	a := m.DB.Assets()

	apikeybytes := make([]byte, 64)
	_, err := rand.Read(apikeybytes)
	if err != nil {
		return err
	}

	i := &Info{
		Plugin:    plugin,
		Name:      name,
		APIKey:    base64.StdEncoding.EncodeToString(apikeybytes),
		Run:       run,
		HeedyDir:  a.FolderPath,
		DataDir:   a.DataDir(),
		PluginDir: path.Join(a.PluginDir(), plugin),
		Config:    a.Config,
	}

	r := &Runner{
		I: i,
		m: m,
	}

	m.Lock()
	m.Runners[i.APIKey] = r

	if run.Cron == nil {
		m.Unlock()
		logrus.Debugf("Starting %s:%s", i.Plugin, i.Name)
		h, err := rt.Start(i)
		if err != nil {
			return err
		}
		r.Handler = h
		if a.Config.Verbose {
			r.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rest.CTX(r).Log.Debugf("Forwarding to %s:%s", i.Plugin, i.Name)
				h.ServeHTTP(w, r)
			})
		}
	} else {
		logrus.Debugf("Adding cron job %s:%s (%s)", i.Plugin, i.Name, *run.Cron)
		r.cid, err = m.cron.AddJob(*run.Cron, r)
		m.Unlock()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Find(plugin, name string) (*Runner, error) {
	m.RLock()
	defer m.RUnlock()
	for _, r := range m.Runners {
		if r.I.Plugin == plugin && r.I.Name == name {

			return r, nil
		}
	}
	return nil, fmt.Errorf("Cound not find active runner %s:%s", plugin, name)
}
func (m *Manager) Stop(plugin, name string) error {
	r, err := m.Find(plugin, name)
	if err != nil {
		return err
	}
	m.Lock()
	r, ok := m.Runners[r.I.APIKey]
	if !ok {
		m.Unlock()
		return fmt.Errorf("Cound not find active runner %s:%s", plugin, name)
	}
	delete(m.Runners, r.I.APIKey)
	m.Unlock()
	if r.I.Run.Cron != nil {
		m.cron.Remove(r.cid)
	}
	logrus.Debugf("Stopping %s:%s", r.I.Plugin, r.I.Name)
	return m.RunTypes[*r.I.Run.Type].Stop(r.I.APIKey)

}

func (m *Manager) Kill() error {
	m.Lock()
	defer m.Unlock()
	for apikey, r := range m.Runners {
		if r.I.Run != nil {
			err := m.RunTypes[*r.I.Run.Type].Kill(apikey)
			if err != nil {
				logrus.Error(err)
			}
		}

	}
	return nil
}

func (m *Manager) StopPlugin(plugin string) error {
	names := []string{}
	m.RLock()
	for _, r := range m.Runners {
		if r.I.Plugin == plugin {
			names = append(names, r.I.Name)
		}
	}
	m.RUnlock()
	for _, name := range names {
		err := m.Stop(plugin, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// https://golang.org/src/net/http/httputil/reverseproxy.go?s=3318:3379#L88
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func (m *Manager) GetHandler(plugin, uri string) (http.Handler, error) {
	plugin, pname, hpath := GetPlugin(plugin, uri)
	if len(plugin) == 0 {
		// If it is not a runner, use a standard reverse proxy
		return NewReverseProxy(m.DB.Assets().DataDir(), uri)
	}

	r, err := m.Find(plugin, pname)
	if err != nil {
		return nil, err
	}
	if r.Handler == nil {
		err = errors.New("No handler found")
	}
	if err != nil || hpath == "/" {
		return r.Handler, err
	}
	// We need to modify the path
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = singleJoiningSlash(hpath, req.URL.Path)
		r.Handler.ServeHTTP(w, req)
	}), nil
}
