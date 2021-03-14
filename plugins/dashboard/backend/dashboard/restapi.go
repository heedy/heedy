package dashboard

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/plugin"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
)

func validateRequest(w http.ResponseWriter, r *http.Request, scope string) (*plugin.ObjectInfo, bool) {
	oi, err := plugin.GetObjectInfo(r)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return nil, false
	}
	if !oi.Access.HasScope(scope) {
		rest.WriteJSONError(w, r, http.StatusForbidden, database.ErrAccessDenied("Insufficient permissions"))
		return nil, false
	}
	return oi, true
}

// ReadDashboard reads the entire dashboard
func ReadHandler(w http.ResponseWriter, r *http.Request) {
	oi, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	c := rest.CTX(r)
	da, err := ReadDashboard(c.DB.AdminDB(), oi.AsObject(), oi.ID, oi.Access.HasScope("write"))

	rest.WriteGzipJSON(w, r, da, err)
}

func WriteHandler(w http.ResponseWriter, r *http.Request) {
	oi, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	c := rest.CTX(r)
	var elements []DashboardElement
	err := rest.UnmarshalRequest(r, &elements)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	err = WriteDashboard(c.DB.AdminDB(), oi.AsObject(), oi.ID, elements)
	rest.WriteResult(w, r, err)
}

func ReadElementHandler(w http.ResponseWriter, r *http.Request) {
	oi, ok := validateRequest(w, r, "read")
	if !ok {
		return
	}
	c := rest.CTX(r)
	eid, err := rest.URLParam(r, "element_id", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	de, err := ReadDashboardElement(c.DB.AdminDB(), oi.AsObject(), oi.ID, eid, oi.Access.HasScope("write"))
	rest.WriteGzipJSON(w, r, de, err)

}

func DeleteElementHandler(w http.ResponseWriter, r *http.Request) {
	oi, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	c := rest.CTX(r)
	eid, err := rest.URLParam(r, "element_id", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	err = DeleteDashboardElement(c.DB.AdminDB(), oi.ID, eid)
	rest.WriteResult(w, r, err)
}

func WriteElementHandler(w http.ResponseWriter, r *http.Request) {
	oi, ok := validateRequest(w, r, "write")
	if !ok {
		return
	}
	c := rest.CTX(r)
	eid, err := rest.URLParam(r, "element_id", nil)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	var element DashboardElement
	err = rest.UnmarshalRequest(r, &element)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	if element.ID != "" && element.ID != eid {
		rest.WriteJSONError(w, r, http.StatusBadRequest, errors.New("bad_request: element ID doesn't match URL"))
		return
	}
	element.ID = eid
	err = WriteDashboard(c.DB.AdminDB(), oi.AsObject(), oi.ID, []DashboardElement{element})
	rest.WriteResult(w, r, err)
}

// Handler is the global router for the timeseries API
var Handler = func() *chi.Mux {
	m := chi.NewMux()

	m.Get("/object/dashboard", ReadHandler)
	m.Post("/object/dashboard", WriteHandler)
	m.Get("/object/dashboard/{element_id}", ReadElementHandler)
	m.Patch("/object/dashboard/{element_id}", WriteElementHandler)
	m.Delete("/object/dashboard/{element_id}", DeleteElementHandler)

	m.NotFound(rest.NotFoundHandler)
	m.MethodNotAllowed(rest.NotFoundHandler)

	return m
}()
