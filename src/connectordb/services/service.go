package services

import (
	log "github.com/Sirupsen/logrus"
	//"os"
	"syscall"
)

type Status int

const (
	// The service hasn't had a call to init yet.
	StatusPreInit = Status(0)
	// The service has done an init
	StatusInit = Status(1)
	// The service is running
	StatusRunning = Status(2)
	// The service is not running
	StatusError   = Status(3)
	StatusCrashed = Status(4)
)

// A streamdb service is something that runs in the background, possibly in
// a goroutine, for example postgres or dbwriter.
type StreamdbService interface {
	// Sets up the service to be run, this may include:
	// - Creating configuration files
	// - Creating a directory, setting up databases, etc.
	Setup() error

	// Initializes the service, sort of a preflight check
	// Possible uses include:
	// - Checking that executables exist
	// - Loading the configuration files
	Init() error

	// Starts the service, calling start on a process that has not successfully
	// been Init()'d is undefined and may have unintended side effects.
	Start() error

	// Stops the service, closing all connections
	// calling stop on a process that is not in the "start" state, is undefined
	// and may cause unintended side effects
	Stop() error

	// Gets the current status of this process (stopped, running)
	Status() Status

	// Kills this process, doesn't have to do anything for non processes
	// calling kill on a process that is not in the "start" state, is undefined
	// and may cause unintended side effects
	Kill() error

	// Gets the name of this service
	Name()
}

// ServiceHelper allows a few of the functions of StreamdbService to be
// automagically done.
type ServiceHelper struct {
	Stat              Status
	ServiceName       string
	StreamdbDirectory string
}

// Initializes some values of servicehelper
func (sh *ServiceHelper) InitServiceHelper(streamdbDirectory, serviceName string) {
	sh.StreamdbDirectory = streamdbDirectory
	sh.ServiceName = serviceName
	sh.Stat = StatusPreInit
}

// Kills a process
func (sh *ServiceHelper) HelperKill() error {
	log.Printf("Killing %s server", sh.Name())
	sh.Stat = StatusInit

	if sh.Stat != StatusRunning {
		return nil
	}

	p, err := GetProcess(sh.StreamdbDirectory, sh.ServiceName, nil)
	if err != nil {
		return err
	}

	if err := p.Kill(); err != nil {
		sh.Stat = StatusError
		return err
	}

	return nil
}

// Returns the name of the service
func (sh *ServiceHelper) Name() string {
	return sh.ServiceName
}

func (sh *ServiceHelper) Status() Status {
	return sh.Stat
}

func (sh *ServiceHelper) HelperStop() error {
	log.Printf("Stopping %s server", sh.Name())

	p, err := GetProcess(sh.StreamdbDirectory, sh.ServiceName, nil)
	if err != nil {
		return err
	}

	log.Printf("%s running on %d", sh.Name(), p.Pid)
	//if err := p.Signal(os.Interrupt); err != nil {
	if err := p.Signal(syscall.SIGTERM); err != nil {
		sh.Stat = StatusError
		return err
	}

	return nil
}
