package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	helpflag         = flag.Bool("help", false, "Prints this help message")
	subcommands_path = flag.String("subcommands_path", "./config/subcommands.json", "Specifies the path to the subcommands json config file.")
	subcommandlist   SubcommandList
	running_forever  bool

	running = true  // is the process running?
	pids    []int // pids of all running processes
	LOGDIR  = "logs"
)

type Subcommand struct {
	Name        string
	Command     []string
	Envflags    map[string]string
	Depends     []string
	Run_forever bool
}

type SubcommandList struct {
	Env         map[string]string // the default environment params
	Subcommands []Subcommand
}

func (sl SubcommandList) ParseEnv(env string) {
	strs := strings.Split(env, "=")

	if len(strs) != 2 {
		fmt.Printf("Cannot parse: '%v' as subcommand environment var\n")
	}

	sl.Env[strs[0]] = strs[1]
}

// Augment the environment vars with those from a command line
func (sl SubcommandList) Augment(params []string) {
	for _, arg := range params {
		if strings.Contains(arg, "|") {
			sl.ParseEnv(arg)
		}
	}
}

func (sl SubcommandList) HasAllKeys(s Subcommand) bool {

	for k, _ := range s.Envflags {
		_, ok := sl.Env[k]
		if !ok {
			return false
		}
	}

	return true
}

func (sl SubcommandList) RunCommand(cmd string) {

	log.Printf("RunCommand %v", cmd)
	var sc Subcommand

	found := false
	for _, s := range sl.Subcommands {
		if s.Name == cmd {
			sc = s
			found = true
			break
		}
	}

	if !found {
		log.Printf("Cannot find subcommand: '%v'\n", cmd)
		os.Exit(1)
	}

	log.Printf("Command Found: %v", sc)

	/**
	  if sl.HasAllKeys(sc) {
	      fmt.Printf("Already have all env vars for: %v, not running.\n", cmd)
	      return
	  }
	  **/

	// add all of our commands to the envflags, ignore those that already exist
	for k, v := range sc.Envflags {
		_, ok := sl.Env[k]
		if !ok {
			sl.Env[k] = v
		}
	}

	for _, cmd := range sc.Depends {
		sl.RunCommand(cmd)
	}

	// Replace all replacable arguments in the command.
	for pos, arg := range sc.Command {
		rep, ok := sl.Env[arg]

		if ok {
			sc.Command[pos] = rep
		}
	}

	log.Printf("Running: %v", sc)

	if len(sc.Command) == 0 {
		log.Printf("No command found.")
		return
	}

	if sc.Run_forever {
		running_forever = true
		go sc.RunForever()
		return
	} else {
		sc.RunOnce()
	}
}

// Runs the subcommand, returns the exit status of the process
func (s Subcommand) RunOnce() bool {
	cmd := exec.Command(s.Command[0], s.Command[1:]...)

	// this will overwrite whenever needed
	outfile, err := os.Create(LOGDIR + "/" + s.Name + ".log")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	cmd.Stderr = outfile

    cmd.Start()
    time.Sleep(time.Duration(5) * time.Second)

	for {
		switch {
		//case running == false:
		//	cmd.Process.Kill()
		//	return false // not success
		case cmd.ProcessState.Exited():
			success := cmd.ProcessState.Success()
			log.Printf("Process %v exited, success? %v", cmd.ProcessState.Pid(), success)
			return success
		default:
			// arbitrary sleep so we don't hog the CPU
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

func (s Subcommand) RunForever() {
	log.Printf("Running %s Forever", s.Name)
	for {
		if !running {
			return
		}

		s.RunOnce()
		log.Printf("%v crashed!", s.Name)
	}
}

func main() {
    running = true

	flag.Parse()

	if *helpflag {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	os.Mkdir(LOGDIR, 0777)

	content, err := ioutil.ReadFile(*subcommands_path)
	if err != nil {

		fmt.Printf("Cannot continue without the subcommands json file.\n")
		os.Exit(2)
	}

	err = json.Unmarshal(content, &subcommandlist)
	if err != nil {
		fmt.Printf("Cannot parse the json file. %v\n", err)
		os.Exit(3)
	}

	subcommandlist.Augment(flag.Args())

	for _, param := range flag.Args() {
		if !strings.Contains(param, "|") {
			subcommandlist.RunCommand(param)
		}
	}

	if running_forever {
		select {} // sleep forever.
	} else {
		fmt.Printf("Finished Executing")
	}
}
