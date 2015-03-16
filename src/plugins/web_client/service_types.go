package web_client

import (
	"net/http"
	"streamdb/dtypes"
	"streamdb/users"
)

type GenericResult struct {
	Status  int    // An HTTP status code
	Message string // Extra data needed to pass along
}

type ReadUserResult struct {
	Users []users.User
	GenericResult
}

type ReadDeviceResult struct {
	Devices []users.Device
	GenericResult
}

type ReadStreamResult struct {
	Streams []users.Stream
	GenericResult
}

type CreateSuccessResult struct {
	Id int64
	GenericResult
}

type DatapointResult struct {
	Data []dtypes.TypedDatapoint
	GenericResult
}

func NewCreateSuccessResult(id int64) CreateSuccessResult {
	var res CreateSuccessResult
	res.Id = id
	res.Message = "Success"
	res.Status = http.StatusOK

	return res
}
