package python

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
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
	"github.com/sirupsen/logrus"
)

type PythonSettings struct {
	sync.Mutex
	IsEnabled bool
	Path      string
	PipArgs   []string
	DB        *database.AdminDB
	Cmd       map[string]*run.Cmd
}

var (
	l        = logrus.WithField("plugin", "python:server")
	settings = PythonSettings{
		Cmd: make(map[string]*run.Cmd),
	}
)

// Start checks the currently set python path to make sure that it is valid
func Start(db *database.AdminDB, i *run.Info, h run.BuiltinHelper) error {

	pyplugin, ok := db.Assets().Config.Plugins["python"]
	if !ok {
		return errors.New("Could not find python plugin configuration")
	}
	p, ok := pyplugin.Settings["path"]
	if !ok || p == nil {
		// Python is not set up
		l.Debug("Python is not set up")
		return nil
	}
	ps, ok := p.(string)
	if !ok {
		return errors.New("Python path must be a string")
	}
	if ps == "" {
		l.Debug("Python is not set up")
		return nil
	}
	pipargs := make([]string, 0)
	ipipargs, ok := pyplugin.Settings["pip_args"]
	if ok {
		apipargs, ok := ipipargs.([]interface{})
		if !ok {
			return errors.New("pip_args must be an array of strings")
		}
		for _, pai := range apipargs {
			pais, ok := pai.(string)
			if !ok {
				return errors.New("pip_args must be an array of strings")
			}
			pipargs = append(pipargs, pais)
		}
	}

	err := TestPython(ps)
	if err == nil {
		settings.IsEnabled = true
		settings.Path = ps
		settings.PipArgs = pipargs
		settings.DB = db
	}
	return err
}

func StartPython(w http.ResponseWriter, r *http.Request) {
	if !settings.IsEnabled {
		rest.WriteJSONError(w, r, http.StatusFailedDependency, errors.New("No python interpreter is set up. Please check your heedy configuration."))
		return
	}
	var i run.Info
	err := rest.UnmarshalRequest(r, &i)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}

	// Prepare the API message
	sm := run.StartMessage{}
	apii, ok := i.Run.Settings["api"]
	if ok {
		apis, ok := apii.(string)
		if !ok {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin 'api' must be string"))
			return
		}
		sm.API = apis
	}

	// Get the filename to run
	filenamei, ok := i.Run.Settings["path"]
	if !ok {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("The path to a python file must be specified"))
		return
	}
	filename, ok := filenamei.(string)
	if !ok {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("The 'path' argument must be a string"))
		return
	}

	// Extract args
	var args []string
	argsi, ok := i.Run.Settings["args"]
	if ok {
		argsa, ok := argsi.([]interface{})
		if !ok {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("The 'args' argument must be an array of strings"))
			return
		}
		args = make([]string, 0, len(argsa))
		for i := range argsa {
			s, ok := argsa[i].(string)
			if !ok {
				rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("The 'args' argument must be an array of strings"))
				return
			}
			args = append(args, s)
		}
	}

	fp := path.Join(i.PluginDir, filename)
	_, err = os.Stat(fp)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	requirementsFile := path.Join(filepath.Dir(fp), "requirements.txt")
	_, err = os.Stat(requirementsFile)
	if err == nil {
		l.Debugf("Setting up requirements from %s", requirementsFile)
		fullargs := append([]string{"-m", "pip", "install", "-r", requirementsFile}, settings.PipArgs...)
		if settings.DB.Verbose {
			l.Debugf("%s %s", settings.Path, strings.Join(fullargs, " "))
		}
		cmd := exec.Command(settings.Path, fullargs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}
	}
	fullargs := append([]string{fp}, args...)
	if settings.DB.Verbose {
		l.Debugf("%s %s", settings.Path, strings.Join(fullargs, " "))
	}
	cmd := exec.Command(settings.Path, fullargs...)
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

	if sm.API != "" {
		method, host, err := run.GetEndpoint(settings.DB.Assets().DataDir(), sm.API)
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

	rest.WriteJSON(w, r, &sm, nil)
}

func RunPython(w http.ResponseWriter, r *http.Request) {
	// TODO: should modify to wait for the process to finish
	StartPython(w, r)
}

func StopPython(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("kill") == "true" {
		KillPython(w, r)
		return
	}
	apikey, err := url.PathUnescape(chi.URLParam(r, "apikey"))
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
	apikey, err := url.PathUnescape(chi.URLParam(r, "apikey"))
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
	mux.Post("/runtypes/python", StartPython)
	mux.Delete("/runtypes/python/{apikey}", StopPython)
	return mux
}()
