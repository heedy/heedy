package main

import (
    "flag"
    "fmt"
    "os"
    "io/ioutil"
    "os/exec"
    "strings"
    "encoding/json"
    )

var (
    helpflag = flag.Bool("help", false, "Prints this help message")
    subcommands_path = flag.String("subcommands_path", "./config/subcommands.json", "Specifies the path to the subcommands json config file.")
    subcommandlist SubcommandList
)


type Subcommand struct{
    Name string
    Command []string
    Envflags map[string]string
    Depends []string
    Run_forever bool
}


type SubcommandList struct {
    Env map[string]string  // the default environment params
    Subcommands []Subcommand
    running_forever bool
}

func (sl SubcommandList) ParseEnv(env string) {
    strs := strings.Split(env, "=")

    if len(strs) != 2{
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
        if ! ok {
            return false
        }
    }

    return true
}


func (sl SubcommandList) RunCommand(cmd string) {

    var sc Subcommand

    found := false
    for _, s := range sl.Subcommands {
        if s.Name == cmd {
            sc = s
            found = true
            break
        }
    }

    if ! found {
        fmt.Printf("Cannot find subcommand: '%v'\n", cmd)
        os.Exit(1)
    }

    if sl.HasAllKeys(sc) {
        fmt.Printf("Already have all env vars for: %v, not running.\n", cmd)
        return
    }

    for _, cmd := range sc.Depends {
        sl.RunCommand(cmd)
    }

    // add all of our commands to the envflags, ignore those that already exist
    for k, v := range sc.Envflags {
        _, ok := sl.Env[k]
        if ! ok {
            sl.Env[k] = v
        }
    }

    // Replace all replacable arguments in the command.
    for pos, arg := range sc.Command {
        rep, ok := sl.Env[arg]

        if ok {
            sc.Command[pos] = rep
        }
    }

    fmt.Printf("Running: %v\n", sc.Command)

    if sc.Run_forever {
        go sc.RunForever()
        sl.running_forever = true
        return
    } else {
        sc.RunOnce()
    }
}

func (s Subcommand) RunOnce() string {
    cmd := exec.Command(s.Command[0], s.Command[1:]...)
    out, err := cmd.Output()
    if err != nil {
        return ""
    }
    return string(out)
}

func (s Subcommand) RunForever() {
    for {
        s.RunOnce()
        fmt.Printf("%v crashed!\n", s.Name)
    }
}

func main() {

    flag.Parse()

    if *helpflag {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        return
    }

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

    for _, param := range(flag.Args()){
        if ! strings.Contains(param, "|") {
            subcommandlist.RunCommand(param)
        }
    }

    if subcommandlist.running_forever{
        select{} // sleep forever.
    } else {
        fmt.Printf("Finished Executing")
    }
}
