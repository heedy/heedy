package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/heedy/heedy/backend/assets"

	"github.com/go-chi/chi"

	log "github.com/sirupsen/logrus"
)

// OverlayManager handles overloading the REST API from plugins
type OverlayManager struct {
	// Overlays specifies the MuxArray element of the ith plugin.
	// The value 0 is raw heedy, and all values >=1 refer to MuxArray[value-1]
	Overlays []int
	MuxArray []*chi.Mux

	apiHandler http.Handler
}

// NewOverlayManager creates the overlays for REST API
func NewOverlayManager(a *assets.Assets, h http.Handler) (*OverlayManager, error) {
	c := a.Config

	muxarray := make([]*chi.Mux, 0)
	overlays := make([]int, len(*c.ActivePlugins)+1)

	if c.ActivePlugins != nil {
		for pindex := range *c.ActivePlugins {
			pname := (*c.ActivePlugins)[pindex]

			psettings, ok := c.Plugins[pname]
			if !ok {
				return nil, fmt.Errorf("Plugin %s configuration not found", pname)
			}

			if psettings.Backend != nil && len(*psettings.Backend) > 0 {
				log.Debugf("Preparing routes for %s", pname)

				mux := chi.NewMux()

				for rname, redirect := range *psettings.Backend {
					revproxy, err := NewReverseProxy(a.DataDir(), redirect)
					if err != nil {
						return nil, err
					}
					log.Debugf("%s: Forwarding %s -> %s ", pname, rname, redirect)
					mux.Handle(rname, revproxy)
				}

				muxarray = append(muxarray, mux)
				overlays[pindex+1] = len(muxarray)

			}
		}

		if len(muxarray) > 0 {
			// Now chain the routes together
			for i := len(muxarray) - 1; i > 0; i-- {
				muxarray[i].NotFound(muxarray[i-1].ServeHTTP)
			}

			// Not found on the first plugin routes to the core heedy API
			muxarray[0].NotFound(h.ServeHTTP)

		}

	}

	return &OverlayManager{
		Overlays:   overlays,
		MuxArray:   muxarray,
		apiHandler: h,
	}, nil
}

func (m *OverlayManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	muxindex := len(m.MuxArray)

	if overlay, ok := r.Header["X-Heedy-Overlay"]; ok {
		if len(overlay) != 1 {
			WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin_error: invalid overlay number"))
			return
		}
		oindex, err := strconv.Atoi(overlay[0])
		if err != nil || oindex < -1 || oindex >= len(m.Overlays) {
			WriteJSONError(w, r, http.StatusBadRequest, errors.New("plugin_error: invalid overlay number"))
			return
		}

		// Remove the overlay header if the OverlayManager can handle it directly (ie: all values other than -1)
		if oindex >= 0 {
			delete(r.Header, "X-Heedy-Overlay")
		} else {
			oindex = 0
		}

		muxindex = m.Overlays[oindex]
	}

	if muxindex == 0 {
		// Go directly to heedy
		m.apiHandler.ServeHTTP(w, r)
		return
	}

	// Go through the overlays
	m.MuxArray[muxindex-1].ServeHTTP(w, r)

}
