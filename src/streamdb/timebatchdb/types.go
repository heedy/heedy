package timebatchdb

import (
    "time"
    "strconv"
    "reflect"
    "errors"
    "streamdb/timebatchdb/datastore"
    )

//The core datatype structs - they are good methods for use in unmarshalling data
//T represents the timestamp
//D represents the data payload
type TypedDatapoint interface {
    Key() string    //Gives the datapoint's key (if any)
}

type DataType interface {
    New() TypedDatapoint                             //Returns the empty TypeDatapoint associated with the datatype
    Load(dp datastore.Datapoint) (TypedDatapoint,error)    //Loads Datapoint into TypeDatapoint
    LoadInto(dp datastore.Datapoint, i interface{}) error //Attempts loading into given struct
    Unload(dp interface{}) (int64,[]byte,error) //Attempts to load the struct's data to the raw stuff
}

type NilType struct {}
func (n NilType) New() TypedDatapoint {
    return nil
}
func (n NilType) Load(dp datastore.Datapoint) (TypedDatapoint,error) {
    return nil,errors.New("This is a NilType")
}
func (n NilType) LoadInto(dp datastore.Datapoint, i interface{}) error {
    return errors.New("This is a NilType")
}
func (n NilType) Unload(dp interface{}) (int64,[]byte,error) {
    return 0,nil,errors.New("This is a NilType")
}

//Checks whether TimeBatchDB supports the given data type, and returns it if yes
func GetType(dtype string) (DataType,bool) {
    v,ok := CoreTypes[dtype]
    return v,ok
}

//Reads the key from an interface
func ExtractKey(i interface{}) string {
    v := extractValue(&i)
    t := v.Type()
    sf,found := t.FieldByName("K")
    if !found {
        sf,found = t.FieldByName("Key")
        if !found {
            sf,found = t.FieldByName("S")
            if !found {
                sf,found = t.FieldByName("Stream")
                if !found {
                    return ""
                }
            }
        }
    }
    v = v.FieldByIndex(sf.Index)
    if v.Kind()==reflect.String {
        return v.String()
    }
    return ""
}


//A simple wrapper for DataRange which returns marshalled data
type TypedRange struct {
    dr datastore.DataRange
    dtype DataType
}

func (tr TypedRange) Close() {
    tr.dr.Close()
}
func (tr TypedRange) Next() interface{} {
    d := tr.dr.Next()
    if d==nil {
        return nil
    }
    v,err := tr.dtype.Load(*d)
    if err != nil {
        return nil
    }
    return v
}
func (tr TypedRange) UnmarshalNext(i interface{}) bool {
    d := tr.dr.Next()
    if d==nil {
        return false
    }

    err := tr.dtype.LoadInto(*d,&i)
    if err!=nil {
        return false
    }
    return true
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

func getTimeField(r reflect.Value) reflect.Value {
    t := r.Type()
    sf,found := t.FieldByName("T")
    if !found {
        sf,found = t.FieldByName("Timestamp")
        if !found {
            return reflect.Value{}  //Return the zero value
        }
    }
    return r.FieldByIndex(sf.Index)
}

//Timestamp supports 3 types: time.Time, int64, and string
func loadTimestamp(dp datastore.Datapoint, r reflect.Value) error {
    tv := getTimeField(r)
    if !tv.IsValid() {
        return errors.New("Could not find Timestamp field in struct")
    }
    if !tv.CanSet() {
        return errors.New("Can't write to Timestamp field - pass by pointer plz!")
    }
    k := tv.Type()

    if (k.Kind()==reflect.String) {
        //Convert the datapoint timestamp to a string
        tv.SetString(intTimeStr(dp.Timestamp()))
    } else if (k.Kind()==reflect.Int64) {
        tv.SetInt(dp.Timestamp())
    } else if (k.Name()=="Time" && k.PkgPath()=="time") {
        ts := time.Unix(0, int64(dp.Timestamp()))
        tv.Set(reflect.ValueOf(ts))
    } else if (k.Kind()==reflect.Ptr) {
        //BUG(daniel): I don't know how to check if this is the correct type of pointer if nil
        ts := time.Unix(0, int64(dp.Timestamp()))
        tv.Set(reflect.ValueOf(&ts))
    } else {
        return errors.New("Timestamp not string/int64/time type")
    }
    return nil
}
func unloadTimestamp(r reflect.Value) (int64,error) {
    tv := getTimeField(r)
    if !tv.IsValid() {
        return 0,errors.New("Could not find Timestamp field in struct")
    }
    k := tv.Type()
    if (k.Kind()==reflect.String) {
        s := tv.String()
        return strTimeInt(s)
    } else if (k.Kind()==reflect.Int64) {
        return tv.Int(),nil
    } else if (k.Name()=="Time" && k.PkgPath()=="time") {
        return tv.Interface().(time.Time).UnixNano(),nil
    } else if (k.Kind()==reflect.Ptr) {
        tv = tv.Elem()
        if !tv.IsValid() {
            return 0,errors.New("Invalid pointer")
        }
        k = tv.Type()
        if (k.Name()=="Time" && k.PkgPath()=="time") {
            return tv.Interface().(time.Time).UnixNano(),nil
        }
    }
    return 0,errors.New("Timestamp not string/int/time type")
}
