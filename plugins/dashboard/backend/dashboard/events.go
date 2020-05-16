package dashboard

import (
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/sirupsen/logrus"
)

// The DashboardEventHandler listens to heedy events, and updates dashboards as necessary.
type DashboardEventHandler struct {
	db *database.AdminDB
}

func (eh DashboardEventHandler) Fire(e *events.Event) {
	// So... Let's check if a dashboard is listening to this event
	if e.Object == "" {
		return // Dashboards only listen to explicit object events
	}

	// Otherwise, get the API calls for matching events
	var s []struct {
		ID           string `db:"id"`
		Type         string `db:"type"`
		BackendQuery []byte `db:"backend_query"`
		Owner        string `db:"owner"`
		ObjectID     string `db:"objectID"`
	}
	err := eh.db.Select(&s, `SELECT dashboard_elements.id,dashboard_elements.type,dashboard_elements.backend_query,objects.owner,objects.id AS objectID FROM dashboard_events 
						JOIN dashboard_elements ON (dashboard_elements.id=dashboard_events.element_id)
						JOIN objects ON (dashboard_elements.object_id=objects.id) 
						WHERE dashboard_events.event=? AND dashboard_events.object_id=?;`, e.Event, e.Object)

	if err != nil {
		logrus.Errorf("Failed to get dashboard events for (%s,%s): %v", e.Event, e.Object, err)
		return
	}

	// For each of these, update the dashboard element!
	for i := range s {
		logrus.Debugf("Updating dashboard element %s/%s", s[i].ObjectID, s[i].ID)
	}
}
