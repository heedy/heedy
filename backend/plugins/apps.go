package plugins

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/heedy/heedy/api/golang/rest"
)

// InternalRequester allows to serve internal requests
type InternalRequester interface {
	ServeInternal(w http.ResponseWriter, r *http.Request, plugin string)
}

func InternalRequest(ir InternalRequester, method, path, plugin string, body interface{}) error {
	var bodybuffer io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodybuffer = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, path, bodybuffer)
	if err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	ir.ServeInternal(rec, req, plugin)
	if rec.Code != http.StatusOK {
		var er rest.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &er)
		if err != nil {
			return err
		}
		return &er
	}
	return nil
}
