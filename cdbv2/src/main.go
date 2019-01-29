package main

import (
	"github.com/connectordb/connectordb/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting")
	cmd.Execute()
}
