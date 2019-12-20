package kv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
)

func GetObjectMeta(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "objid")
	ctx := rest.CTX(r)
	var jo []database.JSONObject

	err := ctx.DB.AdminDB().Select(&jo, "SELECT meta FROM objects WHERE id=?", aid)
	rest.WriteJSON(w, r, jo[0], err)
}

func SetObjectMeta(w http.ResponseWriter, r *http.Request) {
	var jo database.JSONObject
	aid := chi.URLParam(r, "objid")
	ctx := rest.CTX(r)
	if err := rest.UnmarshalRequest(r, &jo); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}

	rest.WriteResult(w, r, ctx.DB.AdminDB().UpdateObject(&database.Object{
		Details: database.Details{
			ID: aid,
		},
		Meta: &jo,
	}))
}

func UpdateObjectMeta(w http.ResponseWriter, r *http.Request) {
	var jo database.JSONObject
	aid := chi.URLParam(r, "objid")
	ctx := rest.CTX(r)
	if err := rest.UnmarshalRequest(r, &jo); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	if len(jo) == 0 {
		rest.WriteResult(w, r, nil)
		return
	}
	sqlStatement := "UPDATE objects SET meta=json_set(meta"
	args := []interface{}{}
	for k, v := range jo {
		kb, err := json.Marshal(k)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
		vb, err := json.Marshal(v)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
			return
		}
		sqlStatement += ",?,json(?)"
		args = append(args, fmt.Sprintf(`$.%s`, string(kb)), string(vb))
	}
	sqlStatement += `) WHERE id=?;`
	args = append(args, aid)

	res, err := ctx.DB.AdminDB().Exec(sqlStatement, args...)
	rest.WriteResult(w, r, database.GetExecError(res, err))

}

func GetObjectMetaKey(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "objid")
	key := chi.URLParam(r, "key")
	ctx := rest.CTX(r)
	var value []string
	kb, err := json.Marshal(key)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	err = ctx.DB.AdminDB().Select(&value, "SELECT json_extract(meta,?) FROM objects WHERE id=?", fmt.Sprintf(`$.%s`, string(kb)), aid)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	bval := []byte(value[0])
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(bval)))
	w.WriteHeader(http.StatusOK)
	w.Write(bval)
}
func SetObjectMetaKey(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "objid")
	key := chi.URLParam(r, "key")
	ctx := rest.CTX(r)
	var value interface{}
	kb, err := json.Marshal(key)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	if err := rest.UnmarshalRequest(r, &value); err != nil {
		rest.WriteJSONError(w, r, 400, err)
		return
	}
	vb, err := json.Marshal(value)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	res, err := ctx.DB.AdminDB().Exec("UPDATE objects SET meta=json_set(meta,?,json(?)) WHERE id=?;", fmt.Sprintf(`$.%s`, string(kb)), string(vb), aid)
	rest.WriteResult(w, r, database.GetExecError(res, err))

}
func DeleteObjectMetaKey(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "objid")
	key := chi.URLParam(r, "key")
	ctx := rest.CTX(r)
	kb, err := json.Marshal(key)
	if err != nil {
		rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}
	res, err := ctx.DB.AdminDB().Exec("UPDATE objects SET meta=json_remove(meta,?) WHERE id=?;", fmt.Sprintf(`$.%s`, string(kb)), aid)
	rest.WriteResult(w, r, database.GetExecError(res, err))
}
