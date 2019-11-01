package plugins

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"

	"github.com/sirupsen/logrus"
)

const (
	statusLoading = iota
	statusClosing
	statusClosed
	statusReady
)

type pluginElement struct {
	// The plugin object
	Plugin *Plugin

	// The next plugin to call in overlay
	Next string
}

//PluginManager handles all aspects of plugin backends
type PluginManager struct {
	sync.RWMutex

	// The internal request handler that handles
	// heedy's server context and whatnot
	IR InternalRequester

	// The default handler to use
	Handler http.Handler
	// The admin database
	ADB *database.AdminDB

	// the plugin manager status
	status int

	Plugins map[string]*pluginElement

	start         string
	order         []string
	SourceManager *SourceManager

	// This is the plugin that is currently being set up
	initializingPlugin *Plugin
}

// NewPluginManager is given assets and the "backend" handler, and returns a plugin manager
func NewPluginManager(db *database.AdminDB, h http.Handler) (*PluginManager, error) {
	return &PluginManager{
		Plugins:       make(map[string]*pluginElement),
		Handler:       h,
		ADB:           db,
		SourceManager: nil,
		status:        statusClosed,
		start:         "none", // Start without any plugins
		order:         []string{},
	}, nil
}

func (pm *PluginManager) Reload() error {
	pm.Close()
	pm.Lock()
	if pm.status != statusClosed || len(pm.Plugins) > 0 {
		pm.Unlock()
		return errors.New("Plugins already being reloaded by another thread")
	}

	a := pm.ADB.Assets()
	sm, err := NewSourceManager(a, pm.Handler)
	if err != nil {
		pm.Unlock()
		return err
	}
	pm.SourceManager = sm
	pm.order = a.Config.GetActivePlugins()
	order := pm.order
	pm.status = statusLoading

	events.AddHandler(pm)
	pm.Unlock()

	// First, perform a cleanup operation: find any apps that are owned by inactive plugins,
	// and remove them if they have empty sources
	pluginexclusion := ""
	neworder := []interface{}{}
	for _, pname := range order {
		pluginexclusion = pluginexclusion + " AND NOT plugin LIKE ?"
		neworder = append(neworder, pname+":%")
	}
	r, err := pm.ADB.Exec(fmt.Sprintf("DELETE FROM sources WHERE last_modified IS NULL AND EXISTS (SELECT 1 FROM apps WHERE plugin IS NOT NULL %s AND apps.id=sources.app);", pluginexclusion), neworder...)
	if err != nil {
		pm.Close()
		return err
	}
	r, err = pm.ADB.Exec(fmt.Sprintf("DELETE FROM apps WHERE plugin IS NOT NULL %s AND NOT EXISTS (SELECT 1 FROM sources WHERE app=apps.id AND last_modified IS NOT NULL);", pluginexclusion), neworder...)
	if err != nil {
		pm.Close()
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		pm.Close()
		return err
	}
	if rows > 0 {
		logrus.Debug("Cleared database of apps from inactive plugins")
	}
	// Now actually initialize the plugins
	for _, pname := range order {
		p, err := NewPlugin(pm.ADB, a, pname)
		if err != nil {
			pm.Close()
			return err
		}
		err = p.BeforeStart(pm.IR)
		if err != nil {
			pm.Close()
			return err
		}
		pm.Lock()
		pm.initializingPlugin = p
		if pm.status != statusLoading {
			pm.Unlock()
			pm.Close()
			return errors.New("Plugins manager closed")
		}
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
				p.Mux.NotFound(pm.SourceManager.ServeHTTP)
			} else {
				p.Mux.NotFound(pm.Plugins[pm.start].Plugin.Mux.ServeHTTP)
			}
			pm.start = pname
		}
		pm.initializingPlugin = nil
		pm.Unlock()

		// Now this plugin's API is active. Run the AfterStart handler
		err = p.AfterStart(pm.IR)
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

func (pm *PluginManager) Fire(e *events.Event) {
	if e.Event == "user_create" {
		// A user was created - run user creation code for each plugin
		go func() {
			pm.RLock()
			defer pm.RUnlock()
			for pname, p := range pm.Plugins {
				err := p.Plugin.OnUserCreate(e.User, pm.IR)
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

// Close shuts down the plugins that are currently
func (pm *PluginManager) Close() error {
	pm.Lock()
	if pm.status == statusClosing {
		pm.Unlock()
		return errors.New("Already closing")
	}
	events.RemoveHandler(pm)
	pm.status = statusClosing
	order := pm.order
	pm.Unlock()

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
	for _, p := range pm.Plugins {
		p.Plugin.Kill()
	}
	return nil
}

func (p *PluginManager) GetProcessByKey(key string) (*Exec, error) {
	p.RLock()
	defer p.RUnlock()
	for _, v := range p.Plugins {
		e, err := v.Plugin.GetProcessByKey(key)
		if err == nil {
			return e, nil
		}
	}
	if p.initializingPlugin != nil {
		e, err := p.initializingPlugin.GetProcessByKey(key)
		if err == nil {
			return e, nil
		}
	}
	return nil, errors.New("The given plugin key was not found")
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

	// The source manager gives the full API, with all sources well-defined
	if serveKey == "none" {
		var sm http.Handler = pm.SourceManager
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
