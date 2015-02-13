package timebatchdb

import (
    "testing"
    "streamdb/timebatchdb/datastore"
    )

func TestTypes(t *testing.T) {
    _,ok := GetType("text")
    if (!ok) {
        t.Errorf("Type text not found")
        return
    }
    _,ok = GetType("thisisnotatype")
    if (ok) {
        t.Errorf("Bad type gives True")
        return
    }

}

func TestTypedRange(t *testing.T) {
    timestamps := []int64{1,2}
    datab := [][]byte{[]byte("test0"),[]byte("test1")}

    da := datastore.CreateDatapointArray(timestamps,datab)

    dtype,ok := GetType("text")
    if (!ok) {
        t.Errorf("Type text not found")
        return
    }
    da.Init()
    tr := TypedRange{da,dtype}
    v := struct {T int64; D string}{}

    ok = tr.UnmarshalNext(&v)
    if !ok || v.T!=1 || v.D!="test0" {
        t.Errorf("Values Incorrect!")
        return
    }
    result :=tr.Next()
    if result==nil {
        t.Errorf("Null valued next!")
        return
    }
    times,data,err := dtype.Unload(result)
    if (err!=nil || times!=2 || string(data)!="test1") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    ok = tr.UnmarshalNext(&v)
    if ok {
        t.Errorf("Null value unmarshal Failed!")
        return
    }
    if tr.Next()!=nil {
        t.Errorf("Null value Failed!")
        return
    }
    tr.Close()
}

func TestKey(t *testing.T) {
    v := struct {K string
    }{"key1"}

    if ExtractKey(v)!="key1" {
        t.Errorf("Key extraction failed")
        return
    }
    v2 := struct {Key string
    }{"key2"}

    if ExtractKey(v2)!="key2" {
        t.Errorf("Key extraction failed")
        return
    }
    v3 := struct {S string
    }{"key3"}

    if ExtractKey(v3)!="key3" {
        t.Errorf("Key extraction failed")
        return
    }
    v4 := struct {Stream string
    }{"key4"}

    if ExtractKey(v4)!="key4" {
        t.Errorf("Key extraction failed")
        return
    }
    v5 := struct {K int
    }{1337}

    if ExtractKey(v5)!="" {
        t.Errorf("Key extraction failed")
        return
    }

    v6 := struct {Kblah string
    }{"key1"}

    if ExtractKey(v6)!="" {
        t.Errorf("Key extraction failed")
        return
    }
}
