package main

import (
	"mime"

	"github.com/connectordb/connectordb/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting")
	mime.AddExtensionType(".jsm", "application/javascript")
	cmd.Execute()
}
