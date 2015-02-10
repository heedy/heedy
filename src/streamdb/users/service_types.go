package users

import (
    "net/http"
    )

type GenericResult struct {
    Status int  // An HTTP status code
    Message string  // Extra data needed to pass along
}


type ReadUserResult struct {
    Users []User
    GenericResult
}

type ReadDeviceResult struct {
    Devices []CleanDevice
    Unsanitized []Device
    GenericResult
}

type ReadStreamResult struct {
    Streams []CleanStream
    Unsanitized []Stream
    GenericResult
}

type CreateSuccessResult struct {
    Id int64
    GenericResult
}

type Datapoint struct {
    Timestamp string // rfc3339 formatted timestamp
    Data string
}

type DatapointResult struct {
    Data []Datapoint
    GenericResult
}

func NewCreateSuccessResult(id int64) CreateSuccessResult {
    var res CreateSuccessResult
    res.Id = id
    res.Message = "Success"
    res.Status = http.StatusOK

    return res
}
