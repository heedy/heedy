package timebatchdb
import (
    "time"
    "strconv"
    )

type Datapoint interface {
    GetData() ([]byte, error)
    SetData([]byte) error
    Type() string
}

//Converts a string time to an integer timestamp
func strTimeInt(st string) (int64,error) {
    ts, err := strconv.ParseInt(st,10,64)
    if err==nil {
        return ts,nil
    }
    var t time.Time
    err = t.UnmarshalText([]byte(st))
    return t.UnixNano(), err
}
//Converts an integer timestamp to a string time
func intTimeStr(ts int64) string {
    str, err := time.Unix(0, int64(ts)).MarshalText()
    if err != nil {
        return "0000-00-00T00:00:00"
    } else {
        return string(str)
    }
}

//The "core" data types have their own special structs which they encode/decode.

type StringDatapoint struct {
    Timestamp string    `json:"t"`
    Data string         `json:"d"`
}
func (d StringDatapoint) Type() string {
    return "text"
}
func (d StringDatapoint) GetData() ([]byte,error) {
    return []byte(d.Data),nil
}
func (d StringDatapoint) SetData(dta []byte) error {
    d.Data = string(dta)
    return nil
}


//The core types
var coretypemap = map[string]Datapoint{"string": StringDatapoint{}}


func NewDatapoint(dtype string) Datapoint {
    return coretypemap[dtype]
}
