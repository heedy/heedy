package dbsetup

import (
	"syscall"
	"util"

	log "github.com/Sirupsen/logrus"
)

//Status represents the status of the service
type Status int

const (
	// The service hasn't had a call to init yet.
	StatusNone    = iota
	StatusStopped = iota
	// The service is running
	StatusRunning = iota
	// The service is not running
	StatusError   = iota
	StatusCrashed = iota
)

type Service interface {
	//Creates the service. This may include:
	// - Creating configuration files
	// - Creating directories, setting up databases
	Create() error

	//Starts the service from existing config
	// - Make sure all necessary files exist, load configs
	// - starts the service
	Start() error

	//Stops the service, closing all connections
	//Running stop on a nonexisting process has undefined behavior
	Stop() error

	//Immediately kills the process
	Kill() error

	//Name returns the name of the service
	Name() string

	//Status of the service
	Status() Status
}

//BaseService allows a few functions of a Service to be automatically implemented
type BaseService struct {
	ServiceDirectory string
	ServiceName      string
	Stat             Status
}

//Name returns the name of the service
func (bs BaseService) Name() string {
	return bs.ServiceName
}

//Status returns the status of the service
func (bs BaseService) Status() Status {
	return bs.Stat
}

//Stop shuts down a service
func (bs BaseService) Stop() error {
	log.Infof("Stopping %s server", bs.Name())

	p, err := util.GetProcess(bs.ServiceDirectory, bs.ServiceName, nil)
	if err != nil {
		return err
	}

	log.Debugf("%s has pid %d", bs.Name(), p.Pid)
	//if err := p.Signal(os.Interrupt); err != nil {
	if err := p.Signal(syscall.SIGTERM); err != nil {
		bs.Stat = StatusError
		return err
	}

	return nil
}

// Kill murders a process in cold blood
func (bs *BaseService) Kill() error {
	log.Warnf("Killing %s server", bs.Name())

	if bs.Stat != StatusRunning {
		return nil
	}
	bs.Stat = StatusStopped

	p, err := util.GetProcess(bs.ServiceDirectory, bs.ServiceName, nil)
	if err != nil {
		return err
	}

	if err := p.Kill(); err != nil {
		bs.Stat = StatusError
		return err
	}

	return nil
}
