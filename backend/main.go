package main

import (
	"mime"

	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/cmd"
)

func main() {
	mime.AddExtensionType(".mjs", "application/javascript")
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}
