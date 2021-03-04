package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/sirupsen/logrus"
)

type Cmd struct {
	Cmd       *exec.Cmd
	done      bool
	iswaiting bool
	sync.Mutex
	waiter chan error
}

func (c *Cmd) Wait() error {
	c.Lock()
	waiting := c.iswaiting
	c.iswaiting = true
	c.Unlock()
	if waiting {
		err := <-c.waiter
		c.waiter <- err
		return err
	}

	err := c.Cmd.Wait()
	c.Lock()
	c.done = true
	c.Unlock()
	c.waiter <- err
	return err
}

func (c *Cmd) Done() bool {
	c.Lock()
	defer c.Unlock()
	return c.done
}

func NewCmd(c *exec.Cmd) *Cmd {
	return &Cmd{
		Cmd:    c,
		waiter: make(chan error, 1),
	}
}

type ExecHandler struct {
	sync.Mutex
	DB  *database.AdminDB
	Cmd map[string]*Cmd
}

func NewExecHandler(db *database.AdminDB) *ExecHandler {
	return &ExecHandler{
		DB:  db,
		Cmd: make(map[string]*Cmd),
	}
}

func (e *ExecHandler) Start(i *Info) (http.Handler, error) {
	// Check to make sure that the settings are set up correctly
	cmdv, ok := i.Run.Config["cmd"]
	if !ok {
		return nil, errors.New("exec requires command to execute")
	}
	cmda, ok := cmdv.([]interface{})
	if !ok {
		return nil, errors.New("cmd must be an array")
	}
	if len(cmda) == 0 {
		return nil, errors.New("empty exec comand")
	}

	cmds := make([]string, len(cmda))
	for i := range cmda {
		s, ok := cmda[i].(string)
		if !ok {
			return nil, fmt.Errorf("cmd element %d must be string", i)
		}
		cmds[i] = s
	}
	var h http.Handler
	var method, host string

	// Next check the API
	apiv, ok := i.Run.Config["api"]
	if ok {
		apis, ok := apiv.(string)
		if !ok {
			return nil, fmt.Errorf("exec api must be string")
		}

		hp, err := NewReverseProxy(e.DB.Assets().DataDir(), apis)
		if err != nil {
			return nil, err
		}
		h = hp
		method, host, err = GetEndpoint(e.DB.Assets().DataDir(), apis)
		if err != nil {
			return nil, err
		}
	}

	// Now set up the process
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = i.PluginDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	// Prepare the input
	infobytes, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}
	_, err = stdin.Write(infobytes)
	if err == nil {
		_, err = stdin.Write([]byte{'\n'})
	}
	if err != nil {
		// Kill the process if can't write to stdin
		cmd.Process.Kill()
		return nil, err
	}

	c := NewCmd(cmd)
	go c.Wait()
	e.Lock()
	e.Cmd[i.APIKey] = c
	e.Unlock()

	if h != nil {
		// There is a handler - wait until the given port is opened

		err = WaitForEndpoint(method, host, c)
		if err != nil {
			e.Lock()
			delete(e.Cmd, i.APIKey)
			e.Unlock()
			cmd.Process.Kill()
			return nil, err
		}
	}

	return h, nil
}

func (e *ExecHandler) Run(i *Info) error {
	_, err := e.Start(i)
	if err != nil {
		return err
	}
	e.Lock()
	cmd, ok := e.Cmd[i.APIKey]
	e.Unlock()
	if !ok {
		return errors.New("Exec failed to retrieve command")
	}
	return cmd.Wait()
}

func (e *ExecHandler) Stop(apikey string) error {
	e.Lock()
	cmd, ok := e.Cmd[apikey]
	e.Unlock()
	if !ok {
		return errors.New("Couldn't find the command")
	}
	cmd.Cmd.Process.Signal(os.Interrupt)

	d := assets.Get().Config.GetRunTimeout()

	sleepDuration := 50 * time.Millisecond
	for i := time.Duration(0); i < d; i += sleepDuration {
		if cmd.Done() {
			return nil
		}
		time.Sleep(sleepDuration)
	}
	logrus.Warn("Process not responding - killing")
	return e.Kill(apikey)
}

func (e *ExecHandler) Kill(apikey string) error {
	e.Lock()
	cmd, ok := e.Cmd[apikey]
	e.Unlock()
	if !ok {
		return errors.New("Couldn't find the command")
	}
	return cmd.Cmd.Process.Kill()
}
