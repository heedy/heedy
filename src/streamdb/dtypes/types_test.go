package dtypes

import (
    "testing"
    "reflect"
    "time"
    "bytes"
    "streamdb/timebatchdb"
    )

func TestGetType(t *testing.T) {
    dpt := &TextDatapoint{BaseDatapoint{"",1.234},"Hello World!"}
    typ,ok := GetType("text[12]/html/someother[43]")
    if ok!=true || !typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[-18]/html")
    if ok!=true || !typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[0]/html")
    if ok!=true || !typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[]/html")
    if ok!=true || !typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[-12]/html")
    if ok!=true || !typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[3]/html/someot")
    if ok!=true || typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[-11]/html/someot")
    if ok!=true || typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
    typ,ok = GetType("text[-12]/html/someot")
    if ok!=true || !typ.IsValid(dpt) {
        t.Errorf("Validity check failed")
        return
    }
}

func TestParseTime(t *testing.T) {

    i,err := ParseTime(1.337)
    if err != nil || i!= 1337000000 {
        t.Errorf("ParseTime (%s,%v)",err,i)
        return
    }
    i,err = ParseTime(int64(1234))
    if err != nil || i!= 1234 {
        t.Errorf("ParseTime (%s,%v)",err,i)
        return
    }
    i,err = ParseTime(time.Unix(0,4567))
    if err != nil || i!= 4567 {
        t.Errorf("ParseTime (%s,%v)",err,i)
        return
    }
    tm := time.Unix(0,45678)
    i,err = ParseTime(&tm)
    if err != nil || i!= 45678 {
        t.Errorf("ParseTime (%s,%v)",err,i)
        return
    }
    i,err = ParseTime("1.337")
    if err != nil || i!= 1337000000 {
        t.Errorf("ParseTime (%s,%v)",err,i)
        return
    }
    str, _ := time.Unix(0, 10123).MarshalText()
    i,err = ParseTime(string(str))
    if err != nil || i!= 10123 {
        t.Errorf("ParseTime (%s,%v,%s)",err,i,str)
        return
    }
    _,err = ParseTime("notatime")
    if err == nil {
        t.Errorf("ParseTime (%s)",err)
        return
    }
    _,err = ParseTime(true)
    if err == nil {
        t.Errorf("ParseTime (%s)",err)
        return
    }
}

func TestBaseDatapoint(t *testing.T) {
    bdp := BaseDatapoint{"hello",int64(1337)}
    i,err := bdp.Timestamp()
    if err != nil || i!= 1337 {
        t.Errorf("Timestamp (%s,%v)",err,i)
        return
    }
    if bdp.Key()!= "hello" {
        t.Errorf("Key (%v)",bdp.Key())
        return
    }
    bdp.LoadTime(1337000000)
    if reflect.ValueOf(bdp.T).Kind()!=reflect.Float64 || bdp.T.(float64)!=1.337 {
        t.Errorf("TimeLoad (%v)",bdp.T)
        return
    }
}

func TestBinaryDatapoint(t *testing.T) {
    dp := timebatchdb.NewDatapoint(1337000000,[]byte("Hello World!"))
    tdp := BinaryDatapoint{}
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    typ := BinaryType{}
    if !typ.IsValid(&tdp) || typ.Len(3).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || !typ.Len(12).IsValid(&tdp) || typ.Len(-3).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}

func TestTextDatapoint(t *testing.T) {
    dp := timebatchdb.NewDatapoint(1337000000,[]byte("Hello World!"))
    tdp := TextDatapoint{}
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    if tdp.D!="Hello World!" {
        t.Errorf("Return Value (%v)",tdp.D)
        return
    }
    typ := TextType{}
    if !typ.IsValid(&tdp) || typ.Len(3).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || !typ.Len(12).IsValid(&tdp) || typ.Len(-3).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}

func TestIntDatapoint(t *testing.T) {
    tdp := IntDatapoint{BaseDatapoint{"","1.337"},-42}
    if len(tdp.Data())>8 {
        t.Errorf("Data size too large: %d",len(tdp.Data()))
        return
    }
    tim,_ := tdp.Timestamp()
    dp := timebatchdb.NewDatapoint(tim,tdp.Data())
    tdp.D=101
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    if tdp.D!=-42 {
        t.Errorf("Return Value (%v)",tdp.D)
        return
    }
    typ := IntType{}
    if !typ.IsValid(&tdp) || !typ.Len(3).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || !typ.Len(12).IsValid(&tdp) || !typ.Len(-3).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}

func TestFloatDatapoint(t *testing.T) {
    tdp := FloatDatapoint{BaseDatapoint{"","1.337"},-42.1456}
    if len(tdp.Data())>8 {
        t.Errorf("Data size too large: %d",len(tdp.Data()))
        return
    }
    tim,_ := tdp.Timestamp()
    dp := timebatchdb.NewDatapoint(tim,tdp.Data())
    tdp.D=101
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    if tdp.D!=-42.1456 {
        t.Errorf("Return Value (%v)",tdp.D)
        return
    }
    typ := FloatType{}
    if !typ.IsValid(&tdp) || !typ.Len(3).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || !typ.Len(12).IsValid(&tdp) || !typ.Len(-3).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}


func TestBoolDatapoint(t *testing.T) {
    tdp := BoolDatapoint{BaseDatapoint{"","1.337"},true}
    if len(tdp.Data())>1 {
        t.Errorf("Data size too large: %d",len(tdp.Data()))
        return
    }
    tim,_ := tdp.Timestamp()
    dp := timebatchdb.NewDatapoint(tim,tdp.Data())
    tdp.D=false
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    if tdp.D!=true {
        t.Errorf("Return Value (%v)",tdp.D)
        return
    }
    typ := BoolType{}
    if !typ.IsValid(&tdp) || !typ.Len(3).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || !typ.Len(12).IsValid(&tdp) || !typ.Len(-3).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}
