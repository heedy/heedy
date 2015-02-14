package timebatchdb

import (
    "reflect"
    "errors"
    "strconv"
    "streamdb/timebatchdb/datastore"
    "encoding/binary"
    "math"
    )


//The permitted datatypes
var CoreTypes = map[string]DataType{"binary": BinaryType{},
    "text": TextType{},
    "float": FloatType{},
    "int": IntType{}}


type BinaryDatapoint struct {
    K string `json:,omitempty`  //The K value represents the datapoint key
    T string
    D []byte
}
func (td BinaryDatapoint) Key() string {
    return td.K
}
type BinaryType struct {}
func (dt BinaryType) New() TypedDatapoint {
    return new(BinaryDatapoint)
}
func (dt BinaryType) Load(dp datastore.Datapoint) (TypedDatapoint,error) {
    dtype := dt.New()
    err := dt.LoadInto(dp,&dtype)
    return dtype,err
}
func (dt BinaryType) LoadInto(dp datastore.Datapoint, i interface{}) error {
    v := extractValue(&i)
    err := loadTimestamp(dp,v)
    if err!=nil {
        return err
    }
    v = getDataField(v)
    if !v.IsValid() {
        return errors.New("Could not find Data field in struct")
    }
    if !v.CanSet() {
        return errors.New("Can't set data field. Pass by struct ptr.")
    }
    k := v.Type()
    var btype byte
    if (k.Kind()==reflect.String) {
        v.SetString(string(dp.Data()))
    } else if (k==reflect.SliceOf(reflect.TypeOf(btype))) {
        v.SetBytes(dp.Data())
    } else {
        return errors.New("Data field type not byte[] or string")
    }

    return nil
}
func (dt BinaryType) Unload(i interface{}) (t int64, d []byte,e error) {
    v := extractValue(&i)
    t, err := unloadTimestamp(v)
    if err != nil {
        return t,nil,err
    }
    v = getDataField(v)
    if !v.IsValid() {
        return t,nil,errors.New("Could not find Data field in struct")
    }
    k := v.Type()
    var btype byte
    if (k.Kind()==reflect.String) {
        d = []byte(v.String())
    } else if (k==reflect.SliceOf(reflect.TypeOf(btype))) {
        d = v.Bytes()
    } else {
        return t,nil,errors.New("Data field type not byte[] or string")
    }
    return t,d,nil
}


type TextDatapoint struct {
    K string `json:,omitempty`
    T string
    D string
}
func (td TextDatapoint) Key() string {
    return td.K
}
type TextType struct {}
func (dt TextType) New() TypedDatapoint {
    return new(TextDatapoint)
}
func (dt TextType) Load(dp datastore.Datapoint) (TypedDatapoint,error) {
    dtype := dt.New()
    err := dt.LoadInto(dp,&dtype)
    return dtype,err
}
func (dt TextType) LoadInto(dp datastore.Datapoint, i interface{}) error {
    //The binary encoding is exactly the same as text encoding
    return BinaryType{}.LoadInto(dp,&i)
}
func (dt TextType) Unload(i interface{}) (t int64, d []byte,e error) {
    //Can use binary unloader!
    return BinaryType{}.Unload(i)
}


type FloatDatapoint struct {
    K string `json:,omitempty`  //The K value represents the datapoint key
    T string
    D float64
}
func (td FloatDatapoint) Key() string {
    return td.K
}
type FloatType struct {}
func (dt FloatType) New() TypedDatapoint {
    return new(FloatDatapoint)
}
func (dt FloatType) Load(dp datastore.Datapoint) (TypedDatapoint,error) {
    dtype := dt.New()
    err := dt.LoadInto(dp,&dtype)
    return dtype,err
}
func (dt FloatType) LoadInto(dp datastore.Datapoint, i interface{}) error {
    v := extractValue(&i)
    err := loadTimestamp(dp,v)
    if err!=nil {
        return err
    }
    v = getDataField(v)
    if !v.IsValid() {
        return errors.New("Could not find Data field in struct")
    }
    if !v.CanSet() {
        return errors.New("Can't set data field. Pass by struct ptr.")
    }
    k := v.Type()
    var btype byte
    d := math.Float64frombits(binary.LittleEndian.Uint64(dp.Data()))
    if (k.Kind()==reflect.Float32 || k.Kind()==reflect.Float64) {
        v.SetFloat(d)
    } else if (k.Kind()==reflect.String) {
        v.SetString(strconv.FormatFloat(d,'e',-1,64))
    } else if (k.Kind()==reflect.Int || k.Kind()==reflect.Int64) {
        v.SetInt(int64(d))
    } else if (k==reflect.SliceOf(reflect.TypeOf(btype))) {
        v.SetBytes(dp.Data())
    } else {
        return errors.New("Data field type not float or string")
    }

    return nil
}
func (dt FloatType) Unload(i interface{}) (t int64, d []byte,e error) {
    v := extractValue(&i)
    t, err := unloadTimestamp(v)
    if err != nil {
        return t,nil,err
    }
    v = getDataField(v)
    if !v.IsValid() {
        return t,nil,errors.New("Could not find Data field in struct")
    }

    k := v.Type()
    var btype byte
    d = make([]byte,8)
    if (k.Kind()==reflect.Float32 || k.Kind()==reflect.Float64) {
        f:= v.Float()
        binary.LittleEndian.PutUint64(d,math.Float64bits(f))
    } else if (k.Kind()==reflect.String) {
        f,err := strconv.ParseFloat(v.String(),64)
        if err!=nil {
            return t,nil,errors.New("Can't parse string as float")
        }
        binary.LittleEndian.PutUint64(d,math.Float64bits(f))
    } else if (k.Kind()==reflect.Int || k.Kind()==reflect.Int64) {
        f:= float64(v.Int())
        binary.LittleEndian.PutUint64(d,math.Float64bits(f))
    } else if (k==reflect.SliceOf(reflect.TypeOf(btype)) && v.Len()==8) {
        d = v.Bytes()
    } else {
        return t,nil, errors.New("Data field type not float or string")
    }
    return t,d,nil
}



type IntDatapoint struct {
    K string `json:,omitempty`  //The K value represents the datapoint key
    T string
    D int64
}
func (td IntDatapoint) Key() string {
    return td.K
}
type IntType struct {}
func (dt IntType) New() TypedDatapoint {
    return new(IntDatapoint)
}
func (dt IntType) Load(dp datastore.Datapoint) (TypedDatapoint,error) {
    dtype := dt.New()
    err := dt.LoadInto(dp,&dtype)
    return dtype,err
}
func (dt IntType) LoadInto(dp datastore.Datapoint, i interface{}) error {
    v := extractValue(&i)
    err := loadTimestamp(dp,v)
    if err!=nil {
        return err
    }
    v = getDataField(v)
    if !v.IsValid() {
        return errors.New("Could not find Data field in struct")
    }
    if !v.CanSet() {
        return errors.New("Can't set data field. Pass by struct ptr.")
    }
    k := v.Type()
    var btype byte
    d := int64(binary.LittleEndian.Uint64(dp.Data()))
    if (k.Kind()==reflect.Int || k.Kind()==reflect.Int64) {
        v.SetInt(d)
    } else if (k.Kind()==reflect.Float32 || k.Kind()==reflect.Float64) {
        v.SetFloat(float64(d))
    } else if (k.Kind()==reflect.String) {
        v.SetString(strconv.FormatInt(d,10))
    } else if (k==reflect.SliceOf(reflect.TypeOf(btype))) {
        v.SetBytes(dp.Data())
    } else {
        return errors.New("Data field type not int or string")
    }

    return nil
}
func (dt IntType) Unload(i interface{}) (t int64, d []byte,e error) {
    v := extractValue(&i)
    t, err := unloadTimestamp(v)
    if err != nil {
        return t,nil,err
    }
    v = getDataField(v)
    if !v.IsValid() {
        return t,nil,errors.New("Could not find Data field in struct")
    }

    k := v.Type()
    var btype byte
    d = make([]byte,8)
    if (k.Kind()==reflect.Int || k.Kind()==reflect.Int64) {
        f:= v.Int()
        binary.LittleEndian.PutUint64(d,uint64(f))
    } else if (k.Kind()==reflect.Float32 || k.Kind()==reflect.Float64) {
        f:= int64(v.Float())
        binary.LittleEndian.PutUint64(d,uint64(f))
    } else if (k.Kind()==reflect.String) {
        f,err := strconv.ParseFloat(v.String(),64) //Float parsing is more general
        if err!=nil {
            return t,nil,errors.New("Can't parse string as int")
        }
        binary.LittleEndian.PutUint64(d,uint64(int64(f)))
    } else if (k==reflect.SliceOf(reflect.TypeOf(btype)) && v.Len()==8) {
        d = v.Bytes()
    } else {
        return t,nil, errors.New("Data field type not int or string")
    }
    return t,d,nil
}
