package users

import (
    "net/http"
    "streamdb/dtypes"
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
