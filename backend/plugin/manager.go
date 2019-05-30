package plugin

import (
	"net/http"

	"github.com/heedy/heedy/backend/assets"
)

// Manager holds all the backend machinery necessary for plugins
type Manager struct {
	Exec   *ExecManager
	Routes *RouteManager
}

// NewManager sets up the plugins' backend functionality, including all running executables,
// and all route overlays
func NewManager(a *assets.Assets, h http.Handler) (*Manager, error) {

	rm, err := NewRouteManager(a, h)
	if err != nil {
		return nil, err
	}
	em := NewExecManager(a)

	err = em.Start()
	if err != nil {
		return nil, err
	}
	return &Manager{
		Exec:   em,
		Routes: rm,
	}, nil
}

// Stop shuts down all background processes that the manager is managing
func (m *Manager) Stop() error {
	return m.Exec.Stop()
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Routes.ServeHTTP(w, r)
}
