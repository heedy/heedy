package plugin

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/robfig/cron"

	log "github.com/sirupsen/logrus"
)

type Job struct {
	sync.Mutex

	Name         string
	Command      []string
	Dir          string
	Proc         *exec.Cmd
	Keepalive    bool
	Stdin        io.WriteCloser
	WriteOnStart []byte
	isStopping   bool
}

func (j *Job) jobHandler() {
	j.Lock()
	p := j.Proc
	j.Unlock()

	// Write the configuration to the process stdin
	j.Stdin.Write(j.WriteOnStart)

	// p is assumed to be running
	err := p.Wait()
	j.Lock()
	if err != nil && !j.isStopping {
		log.Errorf("Process %s finished with error %s", j.Name, err.Error())
	} else {
		log.Infof("Process %s finished", j.Name)
	}
	j.Proc = nil
	j.Unlock()

}

func (j *Job) Run() {
	log.Infof("Running %s", j.Name)
	err := j.Start()
	if err != nil {
		log.Warn(err)
	}
}

func (j *Job) IsRunning() bool {
	j.Lock()
	defer j.Unlock()
	return j.Proc != nil
}

func (j *Job) Start() (err error) {
	j.Lock()
	defer j.Unlock()

	if j.Proc != nil {
		return fmt.Errorf("Plugin executable %s is already running - won't start another instance", j.Name)
	}
	// Runs the job if it is not running
	j.Proc = exec.Command(j.Command[0], j.Command[1:]...)
	j.Proc.Stdout = os.Stdout
	j.Proc.Stderr = os.Stderr
	j.Proc.Dir = j.Dir
	j.Stdin, err = j.Proc.StdinPipe()
	if err != nil {
		return err
	}

	err = j.Proc.Start()
	if err != nil {
		j.Proc = nil
		return err
	}

	// Run the job handler
	go j.jobHandler()

	return nil
}

func (j *Job) Interrupt() error {
	j.Lock()
	defer j.Unlock()
	j.isStopping = true
	if j.Proc != nil {
		return j.Proc.Process.Signal(os.Interrupt)
	}
	return nil
}

func (j *Job) Kill() error {
	j.Lock()
	defer j.Unlock()
	j.isStopping = true
	if j.Proc != nil {
		return j.Proc.Process.Kill()
	}
	return nil // Process not running
}

func NewJob(name string, cmd []string, dir string, writeOnStart []byte) *Job {
	return &Job{
		Name:         name,
		Command:      cmd,
		Dir:          dir,
		WriteOnStart: writeOnStart,
	}
}

type ProcessHandler struct {
	sync.Mutex

	// The cron daemon to run in the background for cron processes
	cron *cron.Cron

	// Currently running commands by the file that is running.
	// This ensures that heedy doesn't stack long-running scripts
	processes map[string]*Job
}

func NewProcessHandler() *ProcessHandler {
	c := cron.New()
	c.Start()
	return &ProcessHandler{
		cron:      c,
		processes: make(map[string]*Job),
	}
}

func (ph *ProcessHandler) StartProcess(name string, cmd []string, dir string, writeOnStart []byte) error {
	ph.Lock()
	defer ph.Unlock()
	j, ok := ph.processes[name]
	if ok {
		return fmt.Errorf("Process '%s' already running", name)
	}
	log.Infof("Starting process %s", name)
	j = NewJob(name, cmd, dir, writeOnStart)

	err := j.Start()
	if err != nil {
		return err
	}

	ph.processes[name] = j

	return nil

}

func (ph *ProcessHandler) StartCron(name string, spec string, cmd []string, dir string, writeOnStart []byte) error {
	ph.Lock()
	defer ph.Unlock()
	j, ok := ph.processes[name]
	if ok {
		return fmt.Errorf("Process '%s' already scheduled", name)
	}

	log.Infof("Adding %s cron job", name)
	j = NewJob(name, cmd, dir, writeOnStart)

	err := ph.cron.AddJob(spec, j)
	if err != nil {
		return err
	}

	ph.processes[name] = j

	return nil
}

func (ph *ProcessHandler) Stop(timeout time.Duration) error {
	ph.Lock()
	defer ph.Unlock()
	ph.cron.Stop()

	sleepduration, _ := time.ParseDuration("50ms")

	for _, j := range ph.processes {
		j.Interrupt()
	}
	for i := time.Duration(0); i < timeout; i += sleepduration {
		anyrunning := false
		for _, j := range ph.processes {
			if j.IsRunning() {
				anyrunning = true
			}
		}
		if !anyrunning {
			return nil
		}
		time.Sleep(sleepduration)
	}
	log.Warn("Killing plugin processes")
	for _, j := range ph.processes {
		j.Kill()
	}
	return nil
}
