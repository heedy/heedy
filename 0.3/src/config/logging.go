package config

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

var prevLogFile *os.File
var prevLogFileName string

// SetLogging sets the given log level and log file for the entire application.
func SetLogging(loglevel string, logfile string) error {

	// First, set up the log level
	switch loglevel {
	default:
		return fmt.Errorf("Unrecognized log level %s. Must be one of debug,info,warn,error", loglevel)
	case "INFO", "info", "":
		log.SetLevel(log.InfoLevel)
	case "WARN", "warn":
		log.SetLevel(log.WarnLevel)
	case "DEBUG", "debug":
		log.SetLevel(log.DebugLevel)
		log.Debug("Setting DEBUG log level")
	case "ERROR", "error":
		log.SetLevel(log.ErrorLevel)
	}

	//  Next set up the log file if it is different than the current one
	if logfile != "" && logfile != prevLogFileName {
		log.Infof("Writing logs to %s", logfile)
		logf, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("Could not open file %s: %s", logfile, err.Error())
		}
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(logf)

		// Now we close previous file if it exists
		if prevLogFile != nil {
			prevLogFile.Close()
		}
		prevLogFile = logf
		prevLogFileName = logfile
	} else if logfile == "" && prevLogFile != nil {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stdout)
		prevLogFile = nil
		prevLogFileName = ""
	}
	return nil
}

// SetLoggingFromConfig is a simple wrapper for SetLogging
func SetLoggingFromConfig(c *Configuration) error {
	return SetLogging(c.LogLevel, c.LogFile)
}
