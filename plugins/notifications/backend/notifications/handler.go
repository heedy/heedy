package notifications

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
)

func readNotifications(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	var o NotificationsQuery
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	n, err := ReadNotifications(c.DB, &o)
	rest.WriteJSON(w, r, n, err)
}

func writeNotification(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	var n Notification
	err := rest.UnmarshalRequest(r, &n)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	rest.WriteResult(w, r, WriteNotification(c.DB, &n))
}

func deleteNotification(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	var o NotificationsQuery
	err := rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	rest.WriteResult(w, r, DeleteNotification(c.DB, &o))
}

func updateNotification(w http.ResponseWriter, r *http.Request) {
	c := rest.CTX(r)
	var n Notification
	var o NotificationsQuery
	err := rest.UnmarshalRequest(r, &n)
	if err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	err = rest.QueryDecoder.Decode(&o, r.URL.Query())
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusBadRequest, err)
		return
	}
	rest.WriteResult(w, r, UpdateNotification(c.DB, &n, &o))
}

// Handler is the main API handler
var Handler = func() *chi.Mux {
	v1mux := chi.NewMux()
	v1mux.Get("/notifications", readNotifications)
	v1mux.Post("/notifications", writeNotification)
	v1mux.Patch("/notifications", updateNotification)
	v1mux.Delete("/notifications", deleteNotification)

	apiMux := chi.NewMux()
	apiMux.NotFound(rest.NotFoundHandler)
	apiMux.MethodNotAllowed(rest.NotFoundHandler)
	apiMux.Mount("/api", v1mux)

	return apiMux
}()
