package rest

import (
	"encoding/json"
	"net/http"
)

//JSONWriter writes the given data as http
func JSONWriter(writer http.ResponseWriter, data interface{}, err error) error {
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return err
	}

	res, err := json.Marshal(data)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return err
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(res)

	return nil
}
