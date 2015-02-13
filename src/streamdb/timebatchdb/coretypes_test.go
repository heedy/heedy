package timebatchdb

import (
    "time"
    "testing"
    "streamdb/timebatchdb/datastore"
    )

func TestBinaryType(t *testing.T) {
    dtype := CoreTypes["binary"]

    dp := datastore.NewDatapoint(12345678,[]byte("Hello World!"))

    v := struct {Timestamp int64
        Data string
    }{}

    err := dtype.LoadInto(dp,&v)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }

    if (v.Timestamp!=12345678) {
        t.Errorf("Incorrect timestamp (%v)",v.Timestamp)
        return
    }
    if (v.Data!="Hello World!") {
        t.Errorf("Incorrect data (%v)",v.Data)
        return
    }

    times,data,err := dtype.Unload(v)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v2 := struct {Timestamp *time.Time
        Data []byte
    }{}

    err = dtype.LoadInto(dp,&v2)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }

    if (v2.Timestamp.UnixNano()!=12345678) {
        t.Errorf("Incorrect timestamp (%v)",v2.Timestamp)
        return
    }
    if (string(v2.Data)!="Hello World!") {
        t.Errorf("Incorrect data (%v)",v.Data)
        return
    }
    times,data,err = dtype.Unload(v2)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v3 := struct {T string
        D []byte
    }{}

    err = dtype.LoadInto(dp,&v3)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }
    var ts time.Time
    err = ts.UnmarshalText([]byte(v3.T))
    if (err!=nil || ts.UnixNano()!=12345678) {
        t.Errorf("Incorrect timestamp (%v %v)",v3.T, err)
        return
    }
    times,data,err = dtype.Unload(&v3)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v4,err := dtype.Load(dp)
    if err!= nil {
        t.Errorf("Load failed (%v)",err)
        return
    }
    times,data,err = dtype.Unload(v4)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

}


func TestTextType(t *testing.T) {
    dtype := CoreTypes["text"]

    dp := datastore.NewDatapoint(12345678,[]byte("Hello World!"))

    v := struct {Timestamp int64
        Data string
    }{}

    err := dtype.LoadInto(dp,&v)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }

    if (v.Timestamp!=12345678) {
        t.Errorf("Incorrect timestamp (%v)",v.Timestamp)
        return
    }
    if (v.Data!="Hello World!") {
        t.Errorf("Incorrect data (%v)",v.Data)
        return
    }

    times,data,err := dtype.Unload(v)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v2 := struct {Timestamp *time.Time
        Data []byte
    }{}

    err = dtype.LoadInto(dp,&v2)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }

    if (v2.Timestamp.UnixNano()!=12345678) {
        t.Errorf("Incorrect timestamp (%v)",v2.Timestamp)
        return
    }
    if (string(v2.Data)!="Hello World!") {
        t.Errorf("Incorrect data (%v)",v.Data)
        return
    }
    times,data,err = dtype.Unload(v2)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v3 := struct {T string
        D []byte
    }{}

    err = dtype.LoadInto(dp,&v3)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }
    var ts time.Time
    err = ts.UnmarshalText([]byte(v3.T))
    if (err!=nil || ts.UnixNano()!=12345678) {
        t.Errorf("Incorrect timestamp (%v %v)",v3.T, err)
        return
    }
    times,data,err = dtype.Unload(&v3)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v4,err := dtype.Load(dp)
    if err!= nil {
        t.Errorf("Load failed (%v)",err)
        return
    }
    times,data,err = dtype.Unload(v4)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

}
