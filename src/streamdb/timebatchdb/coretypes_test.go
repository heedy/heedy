package timebatchdb

import (
    "time"
    "testing"
    "streamdb/timebatchdb/datastore"
    "bytes"
    //"math"
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
        t.Errorf("Incorrect data (%v)",v2.Data)
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
    if err!= nil || v4.Key()!="" {
        t.Errorf("Load failed (%v)",err)
        return
    }
    times,data,err = dtype.Unload(v4)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

    v5 := struct {Timestamp time.Time
        Data []byte
    }{}

    err = dtype.LoadInto(dp,&v5)
    if (err!=nil) {
       t.Errorf("Couldn't load datapoint (%s)",err)
       return
    }

    if (v5.Timestamp.UnixNano()!=12345678) {
        t.Errorf("Incorrect timestamp (%v)",v5.Timestamp)
        return
    }
    if (string(v5.Data)!="Hello World!") {
        t.Errorf("Incorrect data (%v)",v5.Data)
        return
    }
    times,data,err = dtype.Unload(v5)
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
    if err!= nil || v4.Key()!="" {
        t.Errorf("Load failed (%v)",err)
        return
    }
    times,data,err = dtype.Unload(v4)
    if (err!=nil || times!=12345678 || string(data)!="Hello World!") {
        t.Errorf("Incorrect unload (%v, %v, %v)",times,data,err)
        return
    }

}

func TestFloatType(t *testing.T) {
    dtype := CoreTypes["float"]

    v := struct {T int64; D float64}{1,3.1415}
    ts,d,err := dtype.Unload(v)
    if err!=nil || ts!=1 || len(d)!=8 {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

    dp := datastore.NewDatapoint(1337,d)

    v.D=5.0

    err = dtype.LoadInto(dp,&v)
    if err!=nil || v.D!=3.1415 || v.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v.D,v.T,err)
        return
    }

    v2 := struct {T int64; D string}{}
    err = dtype.LoadInto(dp,&v2)
    if err!=nil || v2.D!="3.1415e+00" || v2.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v2.D,v2.T,err)
        return
    }
    v2.D="3.1415"
    ts,d,err = dtype.Unload(v2)
    if err!=nil || ts!=1337 || 0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

    v3 := struct {T int64; D int}{}
    err = dtype.LoadInto(dp,&v3)
    if err!=nil || v3.D!=3 || v3.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v3.D,v3.T,err)
        return
    }
    ts,d,err = dtype.Unload(v3)
    if err!=nil || ts!=1337 || len(d)!=8 {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

    v4 := struct {T int64; D []byte}{}
    err = dtype.LoadInto(dp,&v4)
    if err!=nil || 0!=bytes.Compare(v4.D,dp.Data()) || v4.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v4.D,v4.T,err)
        return
    }
    ts,d,err = dtype.Unload(v4)
    if err!=nil || ts!=1337 || 0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }


    v5,err := dtype.Load(dp)
    if err!=nil || v5.Key()!="" {
        t.Errorf("Load failed (%v)",err)
        return
    }
    ts,d,err = dtype.Unload(v5)
    if err!=nil || ts!=1337 || 0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

}



func TestIntType(t *testing.T) {
    dtype := CoreTypes["int"]

    v := struct {T int64; D float64}{1,-3.0}
    ts,d,err := dtype.Unload(v)
    if err!=nil || ts!=1 || len(d)!=8 {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

    dp := datastore.NewDatapoint(1337,d)

    v.D=5.0

    err = dtype.LoadInto(dp,&v)
    if err!=nil || v.D!=-3.0 || v.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v.D,v.T,err)
        return
    }

    v2 := struct {T int64; D string}{}
    err = dtype.LoadInto(dp,&v2)
    if err!=nil || v2.D!="-3" || v2.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v2.D,v2.T,err)
        return
    }
    ts,d,err = dtype.Unload(v2)
    if err!=nil || ts!=1337 || 0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

    v3 := struct {T int64; D int}{}
    err = dtype.LoadInto(dp,&v3)
    if err!=nil || v3.D!=-3 || v3.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v3.D,v3.T,err)
        return
    }
    ts,d,err = dtype.Unload(v3)
    if err!=nil || ts!=1337 ||  0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }

    v4 := struct {T int64; D []byte}{}
    err = dtype.LoadInto(dp,&v4)
    if err!=nil || 0!=bytes.Compare(v4.D,dp.Data()) || v4.T != 1337 {
        t.Errorf("Incorrect load (%v, %v, %v)",v4.D,v4.T,err)
        return
    }
    ts,d,err = dtype.Unload(v4)
    if err!=nil || ts!=1337 || 0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }


    v5,err := dtype.Load(dp)
    if err!=nil || v5.Key()!="" {
        t.Errorf("Load failed (%v)",err)
        return
    }
    ts,d,err = dtype.Unload(v5)
    if err!=nil || ts!=1337 || 0!=bytes.Compare(d,dp.Data()) {
        t.Errorf("Unload failed (%v,%v,%v)",err,ts,d)
        return
    }



}
