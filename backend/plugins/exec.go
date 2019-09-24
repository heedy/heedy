package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/heedy/heedy/backend/assets"

	log "github.com/sirupsen/logrus"
)

// Exec represents a single executable that is being run.
// The struct is marshalled to json and sent to the executable
// when run
type Exec struct {
	Plugin string                `json:"plugin"`
	Exec   string                `json:"exec"`
	APIKey string                `json:"apikey"`
	Config *assets.Configuration `json:"config"`

	RootDir   string `json:"root_dir"`
	DataDir   string `json:"data_dir"`
	PluginDir string `json:"plugin_dir"`

	cmd []string

	proc       *exec.Cmd
	isStopping bool
	keepAlive  bool

	err error

	// Allows locking and unlocking the process
	sync.Mutex
}

func (e *Exec) Run() {
	log.Debugf("%s: Running cron job %s", e.Plugin, e.Exec)
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

func (e *Exec) HadError() error {
	e.Lock()
	defer e.Unlock()
	return e.err
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
			log.Errorf("%s: %s finished with error %s", e.Plugin, e.Exec, err.Error())
			e.err = err
		} else {
			log.Debugf("%s: %s closed", e.Plugin, e.Exec)
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
