package main

import (
	"github.com/sirupsen/logrus"

	"github.com/heedy/heedy/backend/cmd"
	"github.com/heedy/heedy/backend/events"

	// Add the plugins, which will register their own routes
	_ "github.com/heedy/heedy/plugins/notifications/backend/notifications"
	_ "github.com/heedy/heedy/plugins/python/backend/python"
	_ "github.com/heedy/heedy/plugins/registry/backend/registry"
	_ "github.com/heedy/heedy/plugins/streams/backend/streams"
)

func main() {
	events.RegisterDatabaseHooks() // We're running the full server, so we want to trigger events on actions in database
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}
