package main

import (
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/cmd"

	// Add the plugins, which will register their own routes
	_ "github.com/heedy/heedy/plugins/streams/backend/api"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}
