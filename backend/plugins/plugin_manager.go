package plugins

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/sirupsen/logrus"
)

const (
	statusLoading = iota
	statusClosing
	statusClosed
	statusReady
)

type pluginElement struct {
	Plugin *Plugin

	// The next plugin to call in overlay
	Next string
}

type PluginManager struct {
	sync.RWMutex

	Plugins map[string]*pluginElement

	RunManager    *run.Manager
	ObjectManager *ObjectManager

	// The default handler to use
	Handler http.Handler
	// The admin database
	ADB *database.AdminDB

	// This is the plugin that is currently being set up
	initializingPlugin *Plugin

	// Which plugin the overlay starts with
	start string
	order []string

	// the plugin manager status
	status int
}

func NewPluginManager(db *database.AdminDB, h http.Handler) (*PluginManager, error) {
	m := run.NewManager(db)
	sm, err := NewObjectManager(db.Assets(), m, h)
	if err != nil {
		return nil, err
	}

	plugins := db.Assets().Config.GetActivePlugins()

	// First, perform a cleanup operation: find any apps that are owned by inactive plugins,
	// and remove them if they have empty objects
	pluginexclusion := ""
	neworder := []interface{}{}
	for _, pname := range plugins {
		pluginexclusion = pluginexclusion + " AND NOT plugin LIKE ?"
		neworder = append(neworder, pname+":%")
	}
	r, err := db.Exec(fmt.Sprintf("DELETE FROM objects WHERE last_modified IS NULL AND EXISTS (SELECT 1 FROM apps WHERE plugin IS NOT NULL %s AND apps.id=objects.app);", pluginexclusion), neworder...)
	if err != nil {
		return nil, err
	}
	r, err = db.Exec(fmt.Sprintf("DELETE FROM apps WHERE plugin IS NOT NULL %s AND NOT EXISTS (SELECT 1 FROM objects WHERE app=apps.id AND last_modified IS NOT NULL);", pluginexclusion), neworder...)
	if err != nil {
		return nil, err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows > 0 {
		logrus.Debug("Cleared database of apps from inactive plugins")
	}

	pm := &PluginManager{
		Plugins:       make(map[string]*pluginElement),
		Handler:       h,
		ADB:           db,
		RunManager:    m,
		ObjectManager: sm,
		start:         "none",
		order:         []string{},
		status:        statusLoading,
	}

	events.AddHandler(pm)

	return pm, nil
}

func (pm *PluginManager) Fire(e *events.Event) {
	if e.Event == "user_create" {
		// A user was created - run user creation code for each plugin
		go func() {
			pm.RLock()
			defer pm.RUnlock()
			for pname, p := range pm.Plugins {

				err := p.Plugin.OnUserCreate(e.User)
				if err != nil {
					logrus.Errorf("User creation failed %s (%s)", err.Error(), pname)
					// Delete the user on failure
					pm.ADB.DelUser(e.User)
					return
				}
			}
		}()

	}
}

func (pm *PluginManager) Start(heedyServer http.Handler) error {
	// First prepare all elements that don't require a plugin
	err := pm.ObjectManager.PreparePlugin("")
	if err != nil {
		pm.Close()
		return err
	}

	plugins := pm.ADB.Assets().Config.GetActivePlugins()

	for _, pname := range plugins {
		p, err := NewPlugin(pm.ADB, pm.RunManager, heedyServer, pname)
		if err != nil {
			pm.Close()
			return err
		}
		pm.Lock()
		if pm.status != statusLoading {
			pm.Unlock()
			pm.Close()
			return errors.New("Plugin manager closed")
		}
		pm.initializingPlugin = p
		pm.Unlock()

		err = p.Start()
		if err != nil {
			pm.Close()
			return err
		}

		// The plugin is now ready to go! We add it to the sequence
		pm.Lock()
		if pm.status != statusLoading {
			pm.Unlock()
			p.Close()
			pm.Close()
			return errors.New("PluginManager was closed during loading")
		}

		pm.Plugins[pname] = &pluginElement{
			Plugin: p,
			Next:   pm.start,
		}
		if p.Mux != nil {
			// The plugin has a router component
			if pm.start == "none" {
				p.Mux.NotFound(pm.ObjectManager.ServeHTTP)
				p.Mux.MethodNotAllowed(pm.ObjectManager.ServeHTTP)
			} else {
				p.Mux.NotFound(pm.Plugins[pm.start].Plugin.Mux.ServeHTTP)
				p.Mux.MethodNotAllowed(pm.Plugins[pm.start].Plugin.Mux.ServeHTTP)
			}
			pm.start = pname
		}
		pm.initializingPlugin = nil
		pm.order = append(pm.order, pname)
		pm.Unlock()

		// Now this plugin's API is active. Set up the object forwards and run the AfterStart handler
		err = pm.ObjectManager.PreparePlugin(pname)
		if err != nil {
			pm.Close()
			return err
		}

		err = p.AfterStart()
		if err != nil {
			pm.Close()
			return err
		}

	}

	pm.Lock()
	pm.status = statusReady
	pm.Unlock()

	return nil
}

func (pm *PluginManager) Close() error {
	pm.Lock()
	if pm.status == statusClosing {
		pm.Unlock()
		return errors.New("Already closing")
	}
	pm.status = statusClosing
	events.RemoveHandler(pm)
	pm.Unlock()

	order := pm.order

	// Close plugins sequentially, in the reverse order of creation, so that they still have access
	// to the API if they need to save state to the database. We achieve this stepwise:
	// first, we remove the plugin from the plugin map, and set the start point to
	// the next plugin in series. We continue doing this until no plugins are left
	for i := len(order) - 1; i >= 0; i-- {
		pm.Lock()
		elem, ok := pm.Plugins[order[i]]
		if ok {
			delete(pm.Plugins, order[i])
			if pm.start == order[i] {
				pm.start = elem.Next
			}
			pm.Unlock()
			elem.Plugin.Close()
		} else {
			pm.Unlock()
		}

	}

	pm.Lock()

	pm.order = []string{}
	pm.status = statusClosed
	pm.start = "none"
	if pm.initializingPlugin != nil {
		ip := pm.initializingPlugin
		pm.Unlock()
		ip.Close()
	} else {
		pm.Unlock()
	}

	return nil
}

func (pm *PluginManager) Kill() error {
	pm.Lock()
	defer pm.Unlock()

	pm.order = []string{}
	pm.status = statusClosed
	pm.start = "none"
	pm.initializingPlugin = nil

	return pm.RunManager.Kill()
}

func (pm *PluginManager) GetInfoByKey(apikey string) (*run.Info, error) {
	pm.RunManager.RLock()
	defer pm.RunManager.RUnlock()
	r, ok := pm.RunManager.Runners[apikey]
	if !ok {
		return nil, errors.New("API key not found")
	}
	return r.I, nil

}

func (pm *PluginManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := rest.CTX(r)
	pm.RLock()

	// If it is not a plugin request, and manager is not ready, return an error
	if pm.status != statusReady && len(ctx.Plugin) == 0 {
		pm.RUnlock()
		rest.WriteJSONError(w, r, http.StatusServiceUnavailable, errors.New("loading: heedy is currently loading plugins"))
		return
	}

	// Otherwise, answer the request as normal, making sure to unlock the moment we're
	// no longer using stuff from PluginManager
	serveKey := pm.start
	if overlay, ok := r.Header["X-Heedy-Overlay"]; ok {
		if len(overlay) != 1 {
			pm.RUnlock()
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin_error: invalid overlay"))
			return
		}
		if overlay[0] == "none" {
			serveKey = "none"
		}
		if overlay[0] == "next" {
			// The overlay is next, so find which plugin we're coming from
			serveKey = pm.Plugins[ctx.Plugin].Next

		}
	}

	// The object manager gives the full API, with all objects well-defined
	if serveKey == "none" {
		var sm http.Handler = pm.ObjectManager
		if sm == nil {
			sm = pm.Handler
		}
		pm.RUnlock()
		sm.ServeHTTP(w, r)
		return
	}
	// If serveKey is not none, serve the given plugin
	pmux := pm.Plugins[pm.start].Plugin.Mux
	pm.RUnlock()

	// Delete the overlay header if we're going to pass through plugins
	delete(r.Header, "X-Heedy-Overlay")
	pmux.ServeHTTP(w, r)

}
