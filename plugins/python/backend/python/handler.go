package python

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type PythonSettings struct {
	sync.Mutex                     // Defends Cmd
	Cmd        map[string]*run.Cmd `mapstructure:"-"`

	IsEnabled bool              `mapstructure:"-"`
	DB        *database.AdminDB `mapstructure:"-"`

	Path           string   `mapstructure:"path"`
	PipArgs        []string `mapstructure:"pip_args"`
	VenvArgs       []string `mapstructure:"venv_args"`
	PerPluginVenv  bool     `mapstructure:"per_plugin_venv"`
	ValidatePython bool     `mapstructure:"validate_python"`
}

var (
	l        = logrus.WithField("plugin", "python:backend")
	settings = PythonSettings{
		Cmd: make(map[string]*run.Cmd),
	}
)

// Start checks the currently set python path to make sure that it is valid
func Start(db *database.AdminDB, i *run.Info, h run.BuiltinHelper) error {
	settings.DB = db

	pyplugin, ok := db.Assets().Config.Plugins["python"]
	if !ok {
		return errors.New("Could not find python plugin configuration")
	}

	err := mapstructure.Decode(pyplugin.Config, &settings)
	if err != nil {
		return err
	}

	if settings.Path == "" {
		// Python is not set up, so don't do anything
		l.Debug("Python is not set up")
		return nil
	}
	if settings.ValidatePython {
		err = ValidatePython(settings.Path)
		settings.IsEnabled = err == nil
	} else {
		settings.IsEnabled = true
	}

	return err
}

func StartPythonProcess(w http.ResponseWriter, r *http.Request) {
	if !settings.IsEnabled {
		rest.WriteJSONError(w, r, http.StatusFailedDependency, errors.New("No valid python interpreter is set up. Please check your heedy configuration."))
		return
	}
	var i run.Info
	err := rest.UnmarshalRequest(r, &i)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	var rs struct {
		Path string   `mapstructure:"path"`
		Args []string `mapstructure:"args,omitempty"`
		API  string   `mapstructure:"api,omitempty"`
	}

	if err = mapstructure.Decode(i.Run.Config, &rs); err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	pypath := settings.Path
	if settings.PerPluginVenv {
		// The python path is now the venv's python
		pypath, err = EnsureVenv(pypath, path.Join(i.HeedyDir, "venv", i.Plugin))
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("Failed to create venv: %w", err))
			return
		}
	}

	fp := path.Join(i.PluginDir, rs.Path)
	_, err = os.Stat(fp)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	requirementsFile := path.Join(filepath.Dir(fp), "requirements.txt")
	_, err = os.Stat(requirementsFile)
	if err == nil {
		l.Debugf("Setting up requirements from %s", requirementsFile)
		err = RunCommand(pypath, append([]string{"-m", "pip", "install", "-r", requirementsFile}, settings.PipArgs...))

		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, fmt.Errorf("Failed to install plugin requirements: %w", err))
			return
		}
	}

	// Run the actual code now
	fullargs := append([]string{rs.Path}, rs.Args...)
	if settings.DB.Verbose {
		l.Debugf("%s %s", pypath, strings.Join(fullargs, " "))
	}
	cmd := exec.Command(pypath, fullargs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = i.PluginDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	// Prepare the input
	infobytes, err := json.Marshal(i)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	if err = cmd.Start(); err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	_, err = stdin.Write(infobytes)
	if err == nil {
		_, err = stdin.Write([]byte{'\n'})
	}
	if err != nil {
		// Kill the process if can't write to stdin
		cmd.Process.Kill()
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	c := run.NewCmd(cmd)
	go c.Wait()
	settings.Lock()
	settings.Cmd[i.APIKey] = c
	settings.Unlock()

	if rs.API != "" {
		method, host, err := run.GetEndpoint(settings.DB.Assets().DataDir(), rs.API)
		if err == nil {
			err = run.WaitForEndpoint(method, host, c)
		}
		if err != nil {
			settings.Lock()
			delete(settings.Cmd, i.APIKey)
			settings.Unlock()
			cmd.Process.Kill()
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	rest.WriteJSON(w, r, run.StartMessage{API: rs.API}, nil)
}

func RunPython(w http.ResponseWriter, r *http.Request) {
	// TODO: should modify to wait for the process to finish
	StartPythonProcess(w, r)
}

func StopPython(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("kill") == "true" {
		KillPython(w, r)
		return
	}
	apikey, err := rest.URLParam(r, "apikey", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	settings.Lock()
	cmd, ok := settings.Cmd[apikey]
	settings.Unlock()
	if !ok {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, errors.New("Can't stop python process: No such process"))
		return
	}
	cmd.Cmd.Process.Signal(os.Interrupt)

	d := assets.Get().Config.GetRunTimeout()

	sleepDuration := 50 * time.Millisecond
	for i := time.Duration(0); i < d; i += sleepDuration {
		if cmd.Done() {
			rest.WriteResult(w, r, nil)
			return
		}
		time.Sleep(sleepDuration)
	}
	l.Warn("Process not responding - killing")
	KillPython(w, r)
}

func KillPython(w http.ResponseWriter, r *http.Request) {
	apikey, err := rest.URLParam(r, "apikey", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	settings.Lock()
	cmd, ok := settings.Cmd[apikey]
	settings.Unlock()
	if !ok {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, errors.New("Couldn't find the command"))
		return
	}
	rest.WriteResult(w, r, cmd.Cmd.Process.Kill())
}

// Handler is the main API handler
var Handler = func() *chi.Mux {
	mux := chi.NewMux()
	mux.NotFound(rest.NotFoundHandler)
	mux.MethodNotAllowed(rest.NotFoundHandler)
	mux.Post("/runtypes/python", StartPythonProcess)
	mux.Delete("/runtypes/python/{apikey}", StopPython)
	return mux
}()
