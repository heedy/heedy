package server

import (
	"io"
	"io/ioutil"

	"encoding/json"

	"net/http"

	"github.com/connectordb/connectordb/assets"
)

//UnmarshalRequest unmarshals the input data to the given interface
func UnmarshalRequest(request *http.Request, unmarshalTo interface{}) error {
	defer request.Body.Close()

	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, unmarshalTo)
}
