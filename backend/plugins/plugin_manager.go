package plugins

import (
	"sync"
	"errors"
	"net/http"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/api/golang/rest"
	
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

	// The default handler to use
	Handler http.Handler
	// The admin database
	ADB *database.AdminDB

	// the plugin manager status
	status int

	Plugins map[string]*pluginElement

	start string
	order []string
	SourceManager *SourceManager
}

// NewPluginManager is given assets and the "backend" handler, and returns a plugin manager
func NewPluginManager(db *database.AdminDB,h http.Handler) (*PluginManager,error) {
	return &PluginManager{
		Plugins: make(map[string]*pluginElement),
		Handler: h,
		ADB: db,
		SourceManager: nil,
		status: statusClosed,
		start: "none", // Start without any plugins
		order: []string{},
	},nil
}

func (pm *PluginManager) Reload() error {
	pm.Close()
	pm.Lock()
	if pm.status!=statusClosed || len(pm.Plugins) > 0 {
		pm.Unlock()
		return errors.New("Plugins already being reloaded by another thread")
	}
	
	a :=  pm.ADB.Assets()
	sm, err :=  NewSourceManager(a,pm.Handler)
	if err!=nil {
		pm.Unlock()
		return err
	}
	pm.SourceManager = sm
	pm.order = a.Config.GetActivePlugins()
	order := pm.order
	pm.status = statusLoading
	pm.Unlock()

	for _,pname := range order {
		p,err := NewPlugin(pm.ADB,a,pname)
		if err!=nil {
			pm.Close()
			return err
		}
		err = p.BeforeStart()
		if err!=nil {
			pm.Close()
			return err
		}
		err = p.Start()
		if err!=nil {
			pm.Close()
			return err
		}
		
		// The plugin is now ready to go! We add it to the sequence
		pm.Lock()
		if pm.status!= statusLoading {
			pm.Unlock()
			p.Close()
			pm.Close()
			return errors.New("PluginManager was closed during loading")
		}

		pm.Plugins[pname] = &pluginElement{
			Plugin: p,
			Next: pm.start,
		}
		if p.Mux!=nil {
			// The plugin has a router component
			if pm.start=="none" {
				p.Mux.NotFound(pm.SourceManager.ServeHTTP)
			} else {
				p.Mux.NotFound(pm.Plugins[pm.start].Plugin.Mux.ServeHTTP)
			}
			pm.start = pname
		}
		pm.Unlock()

		// Now this plugin's API is active. Run the AfterStart handler
		err = p.AfterStart()
		if err!=nil {
			pm.Close()
			return err
		}

	}

	pm.Lock()
	pm.status = statusReady
	pm.Unlock()

	return nil
}

// Close shuts down the plugins that are currently
func (pm *PluginManager) Close() error {
	pm.Lock()
	if pm.status== statusClosing {
		pm.Unlock()
		return errors.New("Already closing")
	}
	pm.status = statusClosing
	order := pm.order
	pm.Unlock()

	// Close plugins sequentially, in the reverse order of creation, so that they still have access
	// to the API if they need to save state to the database. We achieve this stepwise:
	// first, we remove the plugin from the plugin map, and set the start point to
	// the next plugin in series. We continue doing this until no plugins are left
	for i:=len(order)-1; i>=0; i-- {
		pm.Lock()
		elem, ok := pm.Plugins[order[i]]
		if ok {
			delete(pm.Plugins,order[i])
			if pm.start==order[i] {
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
	pm.Unlock()
	return nil
}

func (p *PluginManager) GetProcessByKey(key string) (*Exec,error) {
	p.RLock()
	defer p.RUnlock()
	for _,v := range p.Plugins {
		e,err := v.Plugin.GetProcessByKey(key)
		if err==nil {
			return e,nil
		}
	}
	return nil, errors.New("The given plugin key was not found")
}

func (pm *PluginManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := rest.CTX(r)
	pm.RLock()

	// If it is not a plugin request, and manager is not ready, return an error
	if pm.status!=statusReady && len(ctx.Plugin) == 0 {
		pm.RUnlock()
		rest.WriteJSONError(w,r,http.StatusServiceUnavailable,errors.New("loading: heedy is currently loading plugins"))
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
		if overlay[0]=="none" {
			serveKey = "none"
		}
		if overlay[0]=="next" {
			// The overlay is next, so find which plugin we're coming from
			serveKey = pm.Plugins[ctx.Plugin].Next

		}
	}

	// The source manager gives the full API, with all sources well-defined
	if serveKey == "none" {
		var sm http.Handler = pm.SourceManager
		if sm==nil {
			sm = pm.Handler
		}
		pm.RUnlock()
		sm.ServeHTTP(w,r)
		return
	}
	// If serveKey is not none, serve the given plugin
	pmux := pm.Plugins[pm.start].Plugin.Mux
	pm.RUnlock()

	// Delete the overlay header if we're going to pass through plugins
	delete(r.Header, "X-Heedy-Overlay")
	pmux.ServeHTTP(w,r)

}