package kv

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
)

func GenerateHandler(authenticator func(ctx *rest.Context, id string, namespace string) (KV, error)) *chi.Mux {
	kvmux := chi.NewMux()

	getauth := func(r *http.Request) (KV, error) {
		ctx := rest.CTX(r)
		id := chi.URLParam(r, "id")
		namespace := chi.URLParam(r, "namespace")
		if id == "" || namespace == "" {
			return nil, errors.New("bad_request: Incomplete information for request")
		}
		return authenticator(ctx, id, namespace)
	}

	kvmux.Get("/{id}/{namespace}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := getauth(r)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusForbidden, err)
			return
		}
		m, err := v.Get()
		rest.WriteJSON(w, r, m, err)
	}))

	kvmux.Post("/{id}/{namespace}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := getauth(r)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusForbidden, err)
			return
		}
		var m map[string]interface{}
		err = rest.UnmarshalRequest(r, &m)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}

		rest.WriteResult(w, r, v.Set(m))
	}))

	kvmux.Patch("/{id}/{namespace}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := getauth(r)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusForbidden, err)
			return
		}
		var m map[string]interface{}
		err = rest.UnmarshalRequest(r, &m)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}

		rest.WriteResult(w, r, v.Update(m))
	}))

	kvmux.Get("/{id}/{namespace}/{key}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := getauth(r)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusForbidden, err)
			return
		}
		m, err := v.GetKey(chi.URLParam(r, "key"))
		rest.WriteJSON(w, r, m, err)
	}))

	kvmux.Post("/{id}/{namespace}/{key}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := getauth(r)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusForbidden, err)
			return
		}
		var m interface{}
		err = rest.UnmarshalRequest(r, &m)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusBadRequest, err)
			return
		}

		rest.WriteResult(w, r, v.SetKey(chi.URLParam(r, "key"), m))
	}))

	kvmux.Delete("/{id}/{namespace}/{key}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := getauth(r)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusForbidden, err)
			return
		}
		rest.WriteResult(w, r, v.DelKey(chi.URLParam(r, "key")))
	}))

	return kvmux
}

var Handler = func() *chi.Mux {

	apiMux := chi.NewMux()
	apiMux.NotFound(rest.NotFoundHandler)
	apiMux.MethodNotAllowed(rest.NotFoundHandler)

	apiMux.Mount("/api/kv/users", GenerateHandler(UserAuth))
	apiMux.Mount("/api/kv/apps", GenerateHandler(AppAuth))
	apiMux.Mount("/api/kv/objects", GenerateHandler(ObjectAuth))

	return apiMux
}()
