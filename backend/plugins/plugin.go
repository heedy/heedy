package plugins

import (
	"fmt"
	"time"
	"errors"
	"path"
	"encoding/base64"
	"crypto/rand"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/assets"

	"github.com/robfig/cron"
	"github.com/go-chi/chi"

	"github.com/sirupsen/logrus"
)

type Plugin struct {
	// The mux that holds the overlay
	Mux *chi.Mux

	// The assets that are used by the server
	Assets *assets.Assets

	// Name of the plugin
	Name string

	// Processes holds all the exec programs being handled by heedy.
	// The map keys are their associated api keys
	Processes map[string]*Exec `json:"exec"`

	// The cron daemon to run in the background for cron processes
	cron *cron.Cron
}

func NewPlugin(db *database.AdminDB,a *assets.Assets, pname string) (*Plugin,error) {
	p := &Plugin{
		Processes: make(map[string]*Exec),
		Assets:    a,
		Name: pname,
	}
	logrus.Debugf("Loading plugin '%s'",pname)
	
	psettings := a.Config.Plugins[pname]

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

	// Initialize the plugin
	return p,nil
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

		}

	}
	return nil
}


func (p *Plugin) BeforeStart() error {
	return nil
}
func (p *Plugin) AfterStart() error {
	return nil
}


// HasKey checks whether the plugin has defined the given api key
func (p *Plugin) GetProcessByKey(key string) (*Exec,error) {
	v, ok := p.Processes[key]
	if ok {
		return v,nil
	}
	return nil,errors.New("No such key")
}

// Signals all processes to stop
func (p *Plugin) Interrupt() error {
	p.cron.Stop()
	for _,e := range p.Processes {
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
	return len(p.Processes)==0
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
	if p.AnyRunning() {
		p.Stop()
	}
	return nil
}