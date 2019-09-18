package plugins

import (
	"net/http"
	"time"
	"errors"
	"fmt"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/assets"

	"github.com/heedy/heedy/api/golang/rest"
)
type pluginElement struct {
	Plugin *Plugin
	Next string
}
// PluginManager handles all aspects of plugins.
type PluginManager struct {
	Plugins map[string]*pluginElement

	start string

	sm *SourceManager 
}

// NewPluginManager is given assets and the "backend" handler, and returns a plugin manager
func NewPluginManager(db *database.AdminDB, h http.Handler) (*PluginManager,error) {
	a := db.Assets()
	c := a.Config
	// First set up all the source types
	sm, err := NewSourceManager(a,h)
	if err!=nil {
		return nil,err
	}
	

	// Next, initialize the plugins, one by one
	pm := make(map[string]*pluginElement)

	ap := c.GetActivePlugins()

	if c.ActivePlugins != nil {
		for pindex := range *c.ActivePlugins {
			pname := (*c.ActivePlugins)[pindex]

			_, ok := c.Plugins[pname]
			if !ok {
				return nil, fmt.Errorf("Plugin %s configuration not found", pname)
			}

			p,err := NewPlugin(db,pname)
			if err!=nil {
				return nil, fmt.Errorf("Plugin %s: %w",pname, err)
			}
			pm[pname] = &pluginElement{
				Plugin: p,
			}
		}
	}
	

	// Next, chain the routes together, starting forward from the raw server
	prevChain := sm.ServeHTTP
	next := "none"
	for i := range ap {
		if pm[ap[i]].Plugin.Mux!=nil {
			pm[ap[i]].Next = next
			next = ap[i]
			pm[ap[i]].Plugin.Mux.NotFound(prevChain)
			prevChain = pm[ap[i]].Plugin.Mux.ServeHTTP
		}
	}

	pmo := &PluginManager{
		sm: sm,
		Plugins: pm,
		start: next,
	}

	// All the plugin forwards have been set up. Now, start the plugins one by one
	
	for i := range ap {
		err := pm[ap[i]].Plugin.Start()
		if err!=nil {
			pmo.Close()
			return nil,err
		}
	}

	return pmo,nil
}


func (p *PluginManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serveKey := p.start
	if overlay, ok := r.Header["X-Heedy-Overlay"]; ok {
		if len(overlay) != 1 {
			rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin_error: invalid overlay"))
			return
		}
		if overlay[0]=="none" {
			serveKey = "none"
		}
		if overlay[0]=="next" {
			// The overlay is next, so find which plugin we're coming from
			serveKey = p.Plugins[rest.CTX(r).Plugin].Next

		}
	}

	// The source manager gives the full API, with all sources well-defined
	if serveKey == "none" {
		p.sm.ServeHTTP(w,r)
		return
	}
	// If serveKey is not none, serve the given plugin

	// Delete the overlay header if we're going to pass through plugins
	delete(r.Header, "X-Heedy-Overlay")
	p.Plugins[p.start].Plugin.Mux.ServeHTTP(w,r)
}

func (p *PluginManager) GetProcessByKey(key string) (*Exec,error) {
	for _,v := range p.Plugins {
		e,err := v.Plugin.GetProcessByKey(key)
		if err==nil {
			return e,nil
		}
	}
	return nil, errors.New("The given plugin key was not found")
}

func (p *PluginManager) Close() error {
	for _,v := range p.Plugins {
		v.Plugin.Interrupt()
	}
	d := assets.Get().Config.GetExecTimeout()

	sleepDuration := 50 * time.Millisecond

	for i := time.Duration(0); i < d; i += sleepDuration {
		anyrunning := false
		for _,v := range p.Plugins {
			if v.Plugin.AnyRunning() {
				anyrunning = true
			}
		}
		if !anyrunning {
			return nil
		}
		time.Sleep(sleepDuration)
	}

	// Shut down all processes
	for _,v := range p.Plugins {
		v.Plugin.Kill()
	}
	return nil
}