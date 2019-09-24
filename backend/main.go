package main

import (
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/cmd"
	"github.com/heedy/heedy/backend/events"

	// Add the plugins, which will register their own routes
	_ "github.com/heedy/heedy/plugins/notifications/backend/api"
	_ "github.com/heedy/heedy/plugins/streams/backend/api"
)

func main() {
	events.RegisterHooks()
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}
