package main

import (
	"commands"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		if !strings.HasPrefix(err.Error(), "unknown command") {
			log.Error(err)
		}
		os.Exit(-1)
	}
}
