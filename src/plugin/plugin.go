package plugin

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"

	"github.com/connectordb/connectordb/assets/config"

	log "github.com/sirupsen/logrus"
)

var freePort = 10001

type pluginMux struct {
	Name string
	mux  *chi.Mux
}

type PluginManager struct {
	ph *ProcessHandler

	// The plugins call the same function, but sets the X-CDB-Chain: <name> <- name of plugin to forward to
	// This header is set automatically
	// The plugin also sets its plugin key: X-CDB-Key: <key> to authenticate the request without messing with the user's auth
	fullMux []pluginMux

	forwarders map[string]*httputil.ReverseProxy

	// Each plugin gets a "root" level API key. This is the associated key map
	PluginKeyMap map[string]string

	Middleware func(http.Handler) http.Handler
}

func NewPluginManager(assetPath string, c *config.Configuration) (*PluginManager, error) {
	pm := &PluginManager{
		ph:           NewProcessHandler(),
		forwarders:   make(map[string]*httputil.ReverseProxy),
		PluginKeyMap: make(map[string]string),
	}

	pluginRoutes := make([]pluginMux, 0)
	if c.ActivePlugins == nil {
		return pm, nil
	}
	// Loop through the plugins,
	for pindex := range *c.ActivePlugins {
		name := (*c.ActivePlugins)[pindex]
		settings, ok := c.Plugins[name]
		if !ok {
			return nil, fmt.Errorf("Plugin %s configuration not found", name)
		}
		log.Infof("Preparing plugin \"%s\"", name)

		// Set up ports for all the jobs
		for _, ejob := range settings.Exec {
			if ejob.Port == nil || *ejob.Port <= 0 {
				jobPort := freePort
				freePort++
				ejob.Port = &jobPort

			}
		}

		// If the plugin defines routes, set up the appropriate mux
		if settings.Routes != nil && len(*settings.Routes) > 0 {
			mux := chi.NewMux()
			for rname, redirect := range *settings.Routes {
				// First, check if the redirect is a URL or an exec name
				if !strings.HasPrefix(redirect, "http") {
					// It is assumed to be an exec name
					job, ok := settings.Exec[redirect]
					if !ok {
						return nil, fmt.Errorf("Plugin %s: Route '%s' does not have a valid target (%s)", name, rname, redirect)
					}
					if job.KeepAlive == nil {
						kalive := true
						job.KeepAlive = &kalive
					}

					// Now we assume that the plugin is at http (in the future might assume https)
					routeForwardingURL := "http://localhost:" + strconv.FormatInt(int64(*job.Port), 10) + "/"
					pluginForwarder, ok := pm.forwarders[routeForwardingURL]
					if !ok {
						u, err := url.Parse(routeForwardingURL)
						if err != nil {
							return nil, err
						}

						// The forwarder forwards requests
						pluginForwarder = httputil.NewSingleHostReverseProxy(u)

						pm.forwarders[routeForwardingURL] = pluginForwarder
					}
					log.Infof("Forwarding %s -> %s", rname, routeForwardingURL)

					(*settings.Routes)[rname] = routeForwardingURL
					mux.Handle(rname, pluginForwarder)

				}
			}
			pluginRoutes = append(pluginRoutes, pluginMux{
				Name: name,
				mux:  mux,
			})
		}

	}

	if len(pluginRoutes) > 0 {
		// Now set up the routes as a big middleware
		for i := len(pluginRoutes) - 1; i > 0; i-- {
			k := i - 1
			pluginRoutes[i].mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
				log.Info("Forwarding Request")

				// Delete the plugin key if it exists
				if _, ok := r.Header["X-CDB-Plugin"]; ok {
					delete(r.Header, "X-CDB-Plugin")
				}

				// Add the overlay header
				r.Header["X-Cdb-Overlay"] = []string{strconv.Itoa(k)}

				// And continue on our merry way
				pluginRoutes[k].mux.ServeHTTP(w, r)
			})
		}
		k := len(pluginRoutes) - 1
		pm.Middleware = func(next http.Handler) http.Handler {
			pluginRoutes[0].mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
				log.Info("Forwarding to CDB")
				next.ServeHTTP(w, r)
			})
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if overlay, ok := r.Header["X-Cdb-Overlay"]; ok {
					log.Info("Overlay detected")
					// Now let's check if the API key is correct
					// The request is asking for an overlay
					pluginKeys, ok := r.Header["X-Cdb-Plugin"]
					if !ok || len(pluginKeys) != 1 {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("400: X-Cdb-Overlay must also include X-Cdb-Plugin header with plugin key"))
						return
					}

					pName, ok := pm.PluginKeyMap[pluginKeys[0]]
					if !ok {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("400: X-Cdb-Plugin invalid"))
						return
					}

					if len(overlay) != 1 {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("400: X-Cdb-Overlay invalid"))
						return
					}
					oindex, err := strconv.Atoi(overlay[0])
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("400: X-Cdb-Overlay invalid"))
						return
					}
					if oindex > len(pluginRoutes) || oindex < 0 {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("400: X-Cdb-Overlay invalid"))
						return
					}
					if oindex == 0 {
						log.Infof("Forwarding Request from %s to CDB", pName)
						next.ServeHTTP(w, r)
					}
					if oindex < len(pluginRoutes) {
						log.Infof("Forwarding Request from %s to %s", pName, pluginRoutes[oindex-1].Name)
						// Delete the plugin key if it exists
						if _, ok := r.Header["X-CDB-Plugin"]; ok {
							delete(r.Header, "X-CDB-Plugin")
						}

						// Add the overlay header
						r.Header["X-Cdb-Overlay"] = []string{strconv.Itoa(k)}

						pluginRoutes[oindex-1].mux.ServeHTTP(w, r)
						return
					}

				}

				log.Info("Using raw data")
				// Delete the plugin key if it exists
				if _, ok := r.Header["X-CDB-Plugin"]; ok {
					delete(r.Header, "X-CDB-Plugin")
				}

				// Add the overlay header
				r.Header["X-Cdb-Overlay"] = []string{strconv.Itoa(k)}

				pluginRoutes[k].mux.ServeHTTP(w, r)
			})
		}
	}

	// Prepare the configuration bytes that wil be written to all plugin executables
	configBytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	for pindex := range *c.ActivePlugins {
		name := (*c.ActivePlugins)[pindex]
		settings, ok := c.Plugins[name]
		if !ok {
			return nil, fmt.Errorf("Plugin %s configuration not found", name)
		}
		// Start the executables
		execDir := path.Join(assetPath, "plugins", name)
		for ename, ejob := range settings.Exec {
			log.Infof("Preparing process \"%s\"", name+"/"+ename)
			if ejob.Cmd == nil {
				pm.ph.Stop(1)
				return nil, fmt.Errorf("")
			}

			// Prepare the plugin API key
			apikey := make([]byte, 64)
			_, err := rand.Read(apikey)
			if err != nil {
				pm.ph.Stop(1)
				return nil, err
			}

			apikeystring := base64.StdEncoding.EncodeToString(apikey)
			pm.PluginKeyMap[apikeystring] = name

			bytesToWrite := append([]byte(apikeystring+"\n"), configBytes...)
			bytesToWrite = append(bytesToWrite, '\n')

			if ejob.Cron != nil && *ejob.Cron != "" {
				// It is a cron job!
				err = pm.ph.StartCron(name+"/"+ename, *ejob.Cron, *ejob.Cmd, execDir, bytesToWrite)

			} else {
				err = pm.ph.StartProcess(name+"/"+ename, *ejob.Cmd, execDir, bytesToWrite)
			}
			if err != nil {
				pm.ph.Stop(1)
				return nil, err
			}
		}
	}

	return pm, nil
}

func (pm *PluginManager) Stop(d time.Duration) error {
	return pm.ph.Stop(d)
}
