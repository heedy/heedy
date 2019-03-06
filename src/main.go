package main

import (
	"mime"

	"github.com/sirupsen/logrus"

	"github.com/connectordb/connectordb/src/cmd"
)

func main() {
	mime.AddExtensionType(".mjs", "application/javascript")
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}
