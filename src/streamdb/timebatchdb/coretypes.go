package timebatchdb

import (
    "reflect"
    "errors"
    "streamdb/timebatchdb/datastore"
    )


//The permitted datatypes
var CoreTypes = map[string]DataType{"binary": BinaryType{},
    "text": TextType{}}


func getDataField(r reflect.Value) reflect.Value {
    t := r.Type()
    sf,found := t.FieldByName("D")
    if !found {
        sf,found = t.FieldByName("Data")
        if !found {
            return reflect.Value{}  //Return the zero value
        }
    }
    return r.FieldByIndex(sf.Index)
}
func extractValue(i interface{}) reflect.Value {
    v := reflect.ValueOf(i)
    //BUG(daniel): This should probably be Interface or Ptr...
    for v.Kind()!=reflect.Struct {
        v = v.Elem()
    }
    return v
}

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




type IntDatapoint struct {
    K string `json:,omitempty`
    T string
    D int64
}
type FloatDatapoint struct {
    K string `json:,omitempty`
    T string
    D float64
}
