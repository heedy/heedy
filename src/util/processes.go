/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrTimeout is thrown when a port does not open on time
	ErrTimeout = errors.New("Timeout on operation reached - it looks like something crashed.")
	//ErrProcessNotFound is thrown when the process is not found
	ErrProcessNotFound = errors.New("The process was not found")

	//PortTimeoutLoops go in time of 100 milliseconds
	PortTimeoutLoops = 100
)

// PortOpen checks if it can connect to the port using TCP
func PortOpen(host string, port int) bool {
	hostPort := fmt.Sprintf("%s:%d", host, port)
	_, err := net.Dial("tcp", hostPort)
	return err == nil
}

//WaitPort waits for a port to open
func WaitPort(host string, port int, err error) error {
	if err != nil {
		return err
	}

	log.Debugf("Waiting for %s:%d to open...", host, port)

	i := 0
	for ; !PortOpen(host, port) && i < PortTimeoutLoops; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	if i >= PortTimeoutLoops {
		return ErrTimeout
	}

	log.Debugf("...%s:%d is now open.", host, port)
	return nil
}

func cmd2Str(command string, args ...string) string {
	return fmt.Sprintf("> %v %v", command, strings.Join(args, " "))
}

//RunCommand runs the given command in foreground
func RunCommand(err error, command string, args ...string) error {
	if err != nil {
		return err
	}
	log.Debugf(cmd2Str(command, args...))

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

//RunDaemon runs the given command as a daemon (in the background)
func RunDaemon(err error, command string, args ...string) (int, error) {
	if err != nil {
		return 0, err
	}
	log.Debugf(cmd2Str(command, args...))

	cmd := exec.Command(command, args...)

	//No need for redirecting stuff, since log/pid files are configured in .conf files
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//I am not convinced at the moment that restarting postgres/other stuff will be a good idea
	//especially since that is what happens when we want to kill them from another process.
	//So, for the moment, just start the process
	err = cmd.Start()
	if err != nil {
		return 0, err
	}
	return cmd.Process.Pid, nil
}

//GetProcess gets the given process using its process name
func GetProcess(streamdbDirectory, procname string, err error) (*os.Process, error) {
	if err != nil {
		return nil, err
	}

	pidfile := filepath.Join(streamdbDirectory, procname+".pid")
	if !PathExists(pidfile) {
		log.Debugf("Pid Not Found For: %s", procname)
		return nil, ErrProcessNotFound
	}

	pidbytes, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return nil, err
	}

	pids := strings.Fields(string(pidbytes))

	if len(pids) < 1 {
		log.Errorf("Numpids = 0 for: %s", procname)
		return nil, ErrProcessNotFound
	}

	pid, err := strconv.Atoi(pids[0])
	if err != nil {
		return nil, err
	}

	p, err := os.FindProcess(pid)

	if err != nil || runtime.GOOS == "windows" {
		return p, err
	}

	// In unix systems, this always succeeds, so we need to explicitly check
	// if the process exists
	// http://stackoverflow.com/questions/15204162/check-if-a-process-exists-in-go-way
	return p, p.Signal(syscall.Signal(0))
}
