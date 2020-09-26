package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/buildinfo"
	"github.com/heedy/heedy/backend/server"
	"github.com/heedy/heedy/backend/updater"
)

var (

	// ErrTooManyArgs is called when given too many args
	ErrTooManyArgs = errors.New("Too many arguments were specified.")
)

var verbose bool
var revert bool
var applyUpdates bool

var forceRun bool
var cpuprofile string
var cpuprofileFile *os.File
var memprofile string

// UserDataDir is identical to os.UserConfigDir, but returns the linux app data folder instead of config folder
func UserDataDir() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = os.Getenv("AppData")
		if dir == "" {
			return "", errors.New("%AppData% is not defined")
		}
	case "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return "", errors.New("$HOME is not defined")
		}
		dir += "/Library/Application Support"
	case "plan9":
		dir = os.Getenv("home")
		if dir == "" {
			return "", errors.New("$home is not defined")
		}
		dir += "/lib"
	default: // Unix
		dir = os.Getenv("XDG_DATA_HOME")
		if dir == "" {
			dir = os.Getenv("HOME")
			if dir == "" {
				return "", errors.New("neither $XDG_DATA_HOME nor $HOME are defined")
			}
			dir += "/.local/share"
		}
	}
	return dir, nil
}

// RootCmd is the root command under which all other commands are placed.
// It is used to initialize all variables that are global for the whole app
var RootCmd = &cobra.Command{
	Use:     "heedy",
	Short:   "Heedy is a personal data repository and analysis system",
	Long:    `Heedy is an aggregator and dashboard for storing and visualizing data gathered by various trackers. It is buit to be extensible and self-contained, with a powerful plugin system allowing for in-depth analysis and action.`,
	Version: buildinfo.Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := assets.NewConfiguration()
		c.Verbose = verbose

		// Check if a database exists in the root directory. If it doesn't runs the equivalent of "heedy create"
		directory, err := UserDataDir()
		if err != nil {
			return err
		}
		directory = path.Join(directory, "heedy")
		if _, err := os.Stat(path.Join(directory, "heedy.conf")); os.IsNotExist(err) {
			// A heedy database does not exist in the config directory

			return server.Setup(server.SetupContext{
				Config:    c,
				Directory: directory,
			}, ":1324")
		}

		directory, err = filepath.Abs(directory)
		if err != nil {
			return err
		}
		logrus.Infof("Using database at %s", directory)
		if err = writepid(directory); err != nil {
			return err
		}

		return updater.Run(updater.Options{
			ConfigDir:   directory,
			AddonConfig: c,
			Revert:      revert,
			Runner: func(a *assets.Assets) error {
				return server.Run(a, nil)
			},
			Update: applyUpdates,
		})
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cpuprofile != "" {
			logrus.Warnf("Creating CPU Profile at '%s'", cpuprofile)
			f, err := os.Create(cpuprofile)
			if err != nil {
				return err
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				return err
			}
			cpuprofileFile = f
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if cpuprofile != "" {
			pprof.StopCPUProfile()
			cpuprofileFile.Close()
		}
		if memprofile != "" {
			logrus.Warnf("Creating Memory Profile at '%s'", memprofile)
			f, err := os.Create(memprofile)
			if err != nil {
				return err
			}
			defer f.Close()
			runtime.GC()
			if err = pprof.WriteHeapProfile(f); err != nil {
				return err
			}
		}

		return nil
	},
}

// Execute runs the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func GetDirectory(args []string) (string, error) {
	if len(args) > 1 {
		return "", ErrTooManyArgs
	}
	var directory string
	if len(args) == 1 {
		directory = args[0]
	} else {
		f, err := UserDataDir()
		if err != nil {
			return "", err
		}
		directory = path.Join(f, "heedy")
	}
	var err error
	directory, err = filepath.Abs(directory)
	if err == nil {
		logrus.Infof("Using database at %s", directory)
	}

	return directory, err
}

func getpid(directory string) (*os.Process, error) {
	b, err := ioutil.ReadFile(path.Join(directory, "heedy.pid"))
	if err != nil {
		return nil, err
	}
	pid, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, err
	}
	return os.FindProcess(pid)
}

func writepid(cdir string) error {
	// First check if the pid exists and is running
	p, err := getpid(cdir)
	if err == nil {
		err = p.Signal(syscall.Signal(0))
		if err == nil && !forceRun {
			return fmt.Errorf("Heedy is already running at pid %d", p.Pid)
		}
	}

	// Create pid
	return ioutil.WriteFile(path.Join(cdir, "heedy.pid"), []byte(strconv.Itoa(os.Getpid())), os.ModePerm)
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Extremely verbose logging of server requests and responses. Only works in DEBUG log level.")
	RootCmd.PersistentFlags().BoolVar(&revert, "revert", false, "Reverts an update from backup if server fails to start")
	RootCmd.PersistentFlags().BoolVar(&applyUpdates, "update", false, "Applies any pending updates")
	RootCmd.PersistentFlags().BoolVar(&forceRun, "force", false, "Force the server to start even if it detects a heedy pid running")
	RootCmd.PersistentFlags().StringVar(&cpuprofile, "cpuprofile", "", "Saves a CPU profile to the given file")
	RootCmd.PersistentFlags().StringVar(&memprofile, "memprofile", "", "Saves a memory profile to the given file")
}
