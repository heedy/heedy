package main

import (
	"github.com/heedy/heedy/backend/cmd"
	"github.com/heedy/heedy/backend/events"
	"github.com/sirupsen/logrus"

	// Add the plugins, which will register their own routes
	// _ "github.com/heedy/heedy/plugins/registry/backend/registry"
	// _ "github.com/heedy/heedy/plugins/dashboard/backend/dashboard"
	_ "github.com/heedy/heedy/plugins/kv/backend/kv"
	_ "github.com/heedy/heedy/plugins/notifications/backend/notifications"
	_ "github.com/heedy/heedy/plugins/python/backend/python"
	_ "github.com/heedy/heedy/plugins/timeseries/backend/timeseries"
)

func main() {
	events.RegisterDatabaseHooks() // We're running the full server, so we want to trigger events on actions in database
	logrus.SetLevel(logrus.DebugLevel)
	cmd.Execute()
}
