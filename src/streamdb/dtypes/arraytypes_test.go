package dtypes

import (
    "testing"
    "bytes"
    "streamdb/timebatchdb"
    )


func TestFloatArray(t *testing.T) {
    tdp := FloatArray{BaseDatapoint{"","1.337"},[]float64{1.1,-2.2}}
    if len(tdp.Data())>8*2 {
        t.Errorf("Data size too large: %d",len(tdp.Data()))
        return
    }
    tim,_ := tdp.Timestamp()
    dp := timebatchdb.NewDatapoint(tim,tdp.Data())
    tdp.D=nil
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    if len(tdp.D)!=2 || tdp.D[0]!=1.1 || tdp.D[1]!=-2.2 {
        t.Errorf("Return Value (%v)",tdp.D)
        return
    }
    typ := FloatArrayType{}
    if !typ.IsValid(&tdp) || !typ.Len(2).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || typ.Len(1).IsValid(&tdp) || !typ.Len(-2).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}


func TestIntArray(t *testing.T) {
    tdp := IntArray{BaseDatapoint{"","1.337"},[]int64{1,-2}}
    if len(tdp.Data())>8*2 {
        t.Errorf("Data size too large: %d",len(tdp.Data()))
        return
    }
    tim,_ := tdp.Timestamp()
    dp := timebatchdb.NewDatapoint(tim,tdp.Data())
    tdp.D=nil
    err := tdp.Load(dp)
    if err!=nil || tdp.T.(float64)!=1.337 || 0!=bytes.Compare(dp.Data(),tdp.Data()) {
        t.Errorf("Load (%s,%v)",err,tdp.D)
        return
    }
    if len(tdp.D)!=2 || tdp.D[0]!=1 || tdp.D[1]!=-2 {
        t.Errorf("Return Value (%v)",tdp.D)
        return
    }
    typ := IntArrayType{}
    if !typ.IsValid(&tdp) || !typ.Len(2).IsValid(&tdp) || !typ.Len(-20).IsValid(&tdp) || typ.Len(1).IsValid(&tdp) || !typ.Len(-2).IsValid(&tdp) || !typ.Len(0).IsValid(&tdp) {
        t.Errorf("Length validity checks failed")
        return
    }
}
