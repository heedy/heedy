package dtypes

import (
    "time"
    "math"
    "strings"
    "strconv"
    "errors"
    "encoding/binary"
    "streamdb/timebatchdb"
    )

var (
    ERROR_TIME_STRING_PARSE = errors.New("Parsing string time failed")
    ERROR_TIME_UNKNOWN = errors.New("Failed to parse unknown type")
    ERROR_CANTLOADNIL = errors.New("Can't load to a nil-typed datapoint")

    CoreTypes = map[string]DataType{"x": &BinaryType{},
                        "s": &TextType{},
                        "i": &IntType{},
                        "f": &FloatType{},
                        "b": &BoolType{}}
    )



//Timestamp supports several types: time.Time, int64 (unixNano), float (unixSeconds) and string.
//This function accepts all of these types and attempts to convert to the int64 UnixNano used by TimebatchDB
func ParseTime(i interface{}) (int64,error) {
    switch t := i.(type) {
        default:
            return 0,ERROR_TIME_UNKNOWN
        case string:
            //The timestamp might be a string of float seconds
            tsf,err := strconv.ParseFloat(t,64)
            if err == nil { //Totally was
                return int64(tsf*1000000000),nil
            }
            var tm time.Time
            err = tm.UnmarshalText([]byte(t))
            if err == nil {
                return tm.UnixNano(),nil
            }
            return 0,ERROR_TIME_STRING_PARSE
        case float64:
            return int64(t*1000000000),nil
        case int64:
            return t,nil
        case time.Time:
            return t.UnixNano(),nil
        case *time.Time:
            return t.UnixNano(),nil
    }
}

//The base datapoint type
type TypedDatapoint interface {
    Key() string
    Timestamp() (int64,error)
    Data() []byte
    Load(dp timebatchdb.Datapoint) error
}

type DataType interface {
    New() TypedDatapoint
    IsValid(dp TypedDatapoint) bool //Sets the permitted length of the data within the datapoint
    Len(i int) DataType //Sets the permitted data length. negative is up to, and positive is "must equal"
}

func GetType(t string) (DataType,bool) {
    //First, we split the string to only worry about the underlying datatype
    i := strings.Index(t,"/")
    if i!=-1 {
        t = t[:i]
    }
    if len(t)==0 {
        return nil,false
    }

    //Now we check if it is an array type

    if t[len(t)-1]==']' {
        i := strings.Index(t,"[")
        if i==-1 {
            return nil,false
        }
        //Okay, we check if the type exists
        dp,ok := ArrayTypes[t[:i]]
        if !ok {
            return nil,false
        }
        //Great - it exists! now let's give it its limit
        lim := int64(0)
        if i+2!=len(t) {    //There is a number
            l2,err := strconv.ParseInt(t[i+1:len(t)-1],10,32)
            if err!=nil {
                return nil,false
            }
            lim = l2
        }
        return dp.Len(int(lim)),true
    }

    //It is not an array type
    dp,ok := CoreTypes[t]
    if !ok {
        return nil,false
    }
    return dp,true
}


//The basic datapoint stuff implementation
type BaseDatapoint struct {
    K string `json:",omitempty"`
    T interface{}
}
func (td BaseDatapoint) Key() string {
    return td.K
}
func (td BaseDatapoint) Timestamp() (int64,error) {
    return ParseTime(td.T)
}
//Loads the given timestamp into the struct as a float unix time
func (td *BaseDatapoint) LoadTime(t int64) {
    td.T = float64(t)/1000000000.0
}


type NilDatapoint struct {
    BaseDatapoint
}
func (td NilDatapoint) Data() []byte {
    return []byte{}
}
func (td *NilDatapoint) Load(dp timebatchdb.Datapoint) error {
    return ERROR_CANTLOADNIL
}
type NilType struct {}
func (t NilType) New() TypedDatapoint {return new(NilDatapoint)}
func (t NilType) Len(i int) DataType { return NilType{} }
func (td NilType) IsValid(tda TypedDatapoint) bool {return false}

type BinaryDatapoint struct {
    BaseDatapoint
    D []byte
}
func (td BinaryDatapoint) Data() []byte {
    return td.D
}
func (td *BinaryDatapoint) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    td.D = dp.Data()
    return nil
}
type BinaryType struct {
    lenlimit int
}
func (t *BinaryType) New() TypedDatapoint {return new(BinaryDatapoint)}
func (t *BinaryType) Len(i int) DataType {
    return &BinaryType{i}
}
func (td *BinaryType) IsValid(tda TypedDatapoint) bool {
    t := tda.(*BinaryDatapoint)
    if td.lenlimit < 0 {
        return -td.lenlimit >= len(t.D)
    } else if td.lenlimit > 0 {
        return td.lenlimit == len(t.D)
    }
    return true
}


type TextDatapoint struct {
    BaseDatapoint
    D string
}
func (td TextDatapoint) Data() []byte {
    return []byte(td.D)
}
func (td *TextDatapoint) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    td.D = string(dp.Data())
    return nil
}
type TextType struct {
    lenlimit int
}
func (t *TextType) New() TypedDatapoint {return new(TextDatapoint)}
func (t *TextType) Len(i int) DataType {
    return &TextType{i}
}
func (td *TextType) IsValid(tda TypedDatapoint) bool {
    t := tda.(*TextDatapoint)
    if td.lenlimit < 0 {
        return -td.lenlimit >= len(t.D)
    } else if td.lenlimit > 0 {
        return td.lenlimit == len(t.D)
    }
    return true
}


type IntDatapoint struct {
    BaseDatapoint
    D int64
}
func (td IntDatapoint) Data() []byte {
    d := make([]byte,8)
    binary.LittleEndian.PutUint64(d,uint64(td.D))
    return d
}
func (td *IntDatapoint) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    td.D = int64(binary.LittleEndian.Uint64(dp.Data()))
    return nil
}
type IntType struct {}
func (t *IntType) New() TypedDatapoint {return new(IntDatapoint)}
func (t *IntType) Len(i int) DataType { return &IntType{} }
func (td *IntType) IsValid(tda TypedDatapoint) bool {return true}

type FloatDatapoint struct {
    BaseDatapoint
    D float64
}
func (td FloatDatapoint) Data() []byte {
    d := make([]byte,8)
    binary.LittleEndian.PutUint64(d,math.Float64bits(td.D))
    return d
}
func (td *FloatDatapoint) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    td.D = math.Float64frombits(binary.LittleEndian.Uint64(dp.Data()))
    return nil
}
type FloatType struct {}
func (t *FloatType) New() TypedDatapoint {return new(FloatDatapoint)}
func (t *FloatType) Len(i int) DataType { return &FloatType{} }
func (td *FloatType) IsValid(tda TypedDatapoint) bool {return true}


type BoolDatapoint struct {
    BaseDatapoint
    D bool
}
func (td BoolDatapoint) Data() []byte {
    d := make([]byte,1)
    d[0] = 0
    if td.D {
        d[0] = 1
    }
    return d
}
func (td *BoolDatapoint) Load(dp timebatchdb.Datapoint) error {
    td.LoadTime(dp.Timestamp())
    td.D = dp.Data()[0]!=0
    return nil
}
type BoolType struct {}
func (t *BoolType) New() TypedDatapoint {return new(BoolDatapoint)}
func (t *BoolType) Len(i int) DataType { return &BoolType{} }
func (td *BoolType) IsValid(tda TypedDatapoint) bool {return true}
