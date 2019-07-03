package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/robfig/cron"

	log "github.com/sirupsen/logrus"
)

// Exec represents a single executable that is being run.
// The struct is marshalled to json and sent to the executable
// when run
type Exec struct {
	Plugin  string                `json:"plugin"`
	Exec    string                `json:"exec"`
	Overlay int                   `json:"overlay"`
	APIKey  string                `json:"apikey"`
	Config  *assets.Configuration `json:"config"`

	MainDir   string `json:"main_dir"`
	DataDir   string `json:"data_dir"`
	PluginDir string `json:"plugin_dir"`

	cmd []string

	proc       *exec.Cmd
	isStopping bool
	keepAlive  bool

	// Allows locking and unlocking the process
	sync.Mutex
}

func (e *Exec) Run() {
	log.Debugf("Running cron job %s/%s", e.Plugin, e.Exec)
	err := e.Start()
	if err != nil {
		log.Error(err)
	}
}

func (e *Exec) IsRunning() bool {
	e.Lock()
	defer e.Unlock()
	return e.proc != nil
}

func (e *Exec) Start() error {
	e.Lock()
	defer e.Unlock()

	if e.proc != nil {
		return fmt.Errorf("%s/%s is already running - won't start another instance", e.Plugin, e.Exec)
	}

	e.proc = exec.Command(e.cmd[0], e.cmd[1:]...)
	e.proc.Stdout = os.Stdout
	e.proc.Stderr = os.Stderr
	e.proc.Dir = e.PluginDir
	stdin, err := e.proc.StdinPipe()
	if err != nil {
		e.proc = nil
		return err
	}

	// Prepare the input
	infobytes, err := json.Marshal(e)
	if err != nil {
		e.proc = nil
		return err
	}

	err = e.proc.Start()
	if err != nil {
		e.proc = nil
		return err
	}

	_, err = stdin.Write(infobytes)
	if err == nil {
		_, err = stdin.Write([]byte{'\n'})
	}
	if err != nil {
		// Kill the process if can't write to stdin
		e.proc.Process.Kill()
		e.proc = nil
		return err
	}

	// Wait until the process exits in a goroutine
	go func() {
		err := e.proc.Wait()

		e.Lock()
		if err != nil && !e.isStopping {
			log.Errorf("%s/%s finished with error %s", e.Plugin, e.Exec, err.Error())
		} else {
			log.Debugf("%s/%s finished", e.Plugin, e.Exec)
		}
		e.proc = nil
		e.isStopping = false
		e.Unlock()
	}()

	return nil
}

func (e *Exec) Interrupt() error {
	e.Lock()
	defer e.Unlock()
	if e.proc != nil {
		e.isStopping = true
		return e.proc.Process.Signal(os.Interrupt)
	}
	return nil
}

func (e *Exec) Kill() error {
	e.Lock()
	defer e.Unlock()

	if e.proc != nil {
		e.isStopping = true
		return e.proc.Process.Kill()
	}
	return nil // Process not running
}

// ExecManager manages all running external processes.
type ExecManager struct {
	// Processes holds all the exec programs being handled by heedy.
	// The map keys are their associated api keys
	Processes map[string]*Exec

	// The assets that are used by the server
	Assets *assets.Assets

	// The cron daemon to run in the background for cron processes
	cron *cron.Cron
}

func NewExecManager(a *assets.Assets) *ExecManager {
	return &ExecManager{
		Processes: make(map[string]*Exec),
		Assets:    a,
	}
}

func (em *ExecManager) Start() error {
	if len(em.Processes) > 0 {
		return errors.New("Must first stop running processes to restart ExecManager")
	}
	em.cron = cron.New()
	em.cron.Start()

	for pindex := range *(em.Assets.Config.ActivePlugins) {
		pname := (*(em.Assets.Config.ActivePlugins))[pindex]
		pv, ok := em.Assets.Config.Plugins[pname]
		if !ok {
			return fmt.Errorf("Could not find plugin %s", pname)
		}

		for ename, ev := range pv.Exec {
			if ev.Enabled == nil || ev.Enabled != nil && *ev.Enabled {
				keepAlive := false
				if ev.KeepAlive != nil {
					keepAlive = *ev.KeepAlive
				}
				if ev.Cmd == nil || len(*ev.Cmd) == 0 {
					em.Stop()
					return fmt.Errorf("%s/%s has empty command", pname, ename)
				}

				// Create an API key for the exec
				apikeybytes := make([]byte, 64)
				_, err := rand.Read(apikeybytes)

				e := &Exec{
					Plugin:    pname,
					Exec:      ename,
					Overlay:   pindex,
					APIKey:    base64.StdEncoding.EncodeToString(apikeybytes),
					Config:    em.Assets.Config,
					MainDir:   em.Assets.FolderPath,
					DataDir:   em.Assets.DataDir(),
					PluginDir: path.Join(em.Assets.PluginDir(), pname),
					keepAlive: keepAlive,
					cmd:       *ev.Cmd, // TODO: Handle Python
				}

				em.Processes[e.APIKey] = e

				if ev.Cron != nil && len(*ev.Cron) > 0 {
					log.Debugf("Enabling cron job %s/%s", pname, ename)
					err = em.cron.AddJob(*ev.Cron, e)
				} else {
					log.Debugf("Running %s/%s", pname, ename)
					err = e.Start()
				}
				if err != nil {
					em.Stop()
					return err
				}

			}
		}

	}
	return nil
}

func (em *ExecManager) Stop() error {
	em.cron.Stop()

	execTimeout := "5s"
	if em.Assets.Config.ExecTimeout != nil {
		execTimeout = *em.Assets.Config.ExecTimeout
	}
	d, err := time.ParseDuration(execTimeout)
	if err != nil {
		log.Error("Invalid exec timeout given in configuration. Using 5s.")
		d = 5 * time.Second
	}

	for _, e := range em.Processes {
		e.Interrupt()
	}

	sleepDuration := 50 * time.Millisecond

	for i := time.Duration(0); i < d; i += sleepDuration {
		anyrunning := false
		for _, e := range em.Processes {
			if e.IsRunning() {
				anyrunning = true
			}
		}
		if !anyrunning {
			return nil
		}
		time.Sleep(sleepDuration)
	}

	for _, e := range em.Processes {
		if e.IsRunning() {
			log.Warnf("Killing %s/%s", e.Plugin, e.Exec)
			e.Kill()
		}
	}

	return nil
}
