package timebatchdb

import (
    "testing"
    "bytes"
    )

func TestDatapoint(t *testing.T) {
    d := NewDatapoint(1337,[]byte("Hello World!"))

    if (d.Len()!=len(d.Bytes()) || d.Timestamp()!=1337 || string(d.Data())!="Hello World!" || d.DataLen()!=12) {
        t.Errorf("Datapoint read error: %s",d)
        return
    }

    //Now check that the bytes can be reread from a file
    buf := new(bytes.Buffer)
    buf.Write(d.Bytes())
    buf.Write([]byte("END"))

    d2,err := ReadDatapoint(buf)
    if err!=nil {
        t.Errorf("Datapoint error %s",err)
        return
    }

    if (string(buf.Next(3))!="END") {
        t.Errorf("Datapoint reading went out of bounds")
        return
    }
    if (d2.Len()!=d.Len() || d2.Timestamp()!=1337 || string(d2.Data())!="Hello World!" || d2.DataLen()!=12) {
        t.Errorf("Datapoint read error: %s",d2)
        return
    }

    //Lastly, check if the datapoint can be created from a byte array
    buf = new(bytes.Buffer)
    buf.Write(d.Bytes())
    buf.Write([]byte("END"))
    d3,n := DatapointFromBytes(buf.Bytes())
    if (int(n)!=d3.Len()) {
        t.Errorf("Datapoint read bytenum error: %s",d3)
        return
    }

    if (d3.Len()!=d.Len() || d3.Timestamp()!=1337 || string(d3.Data())!="Hello World!" || d3.DataLen()!=12) {
        t.Errorf("Datapoint read error: %s",d3)
        return
    }


}

func TestKeyedDatapoint(t *testing.T) {
    d := NewKeyedDatapoint("testing/hi",1337,[]byte("Hello World!"))

    if (d.Key()!="testing/hi" || d.Len()!=len(d.Bytes()) || d.Timestamp()!=1337 || string(d.Data())!="Hello World!" || d.DataLen()!=12) {
        t.Errorf("KeyedDatapoint read error: %s",d)
        return
    }

    //Now make sure that the internal datapoint is correct
    dp := d.Datapoint()
    if (dp.Len()!=len(dp.Bytes()) || dp.Timestamp()!=1337 || string(dp.Data())!="Hello World!" || dp.DataLen()!=12) {
        t.Errorf("Datapoint read error: %s",dp)
        return
    }

    //Now check that the bytes can be reread from a file
    buf := new(bytes.Buffer)
    buf.Write(d.Bytes())
    buf.Write([]byte("END"))

    d2,err := ReadKeyedDatapoint(buf)
    if err!=nil {
        t.Errorf("KeyedDatapoint error %s",err)
        return
    }

    if (string(buf.Next(3))!="END") {
        t.Errorf("KeyedDatapoint reading went out of bounds")
        return
    }

    if (d2.Key()!="testing/hi" || d2.Len()!=d.Len() || d2.Timestamp()!=1337 || string(d2.Data())!="Hello World!" || d2.DataLen()!=12) {
        t.Errorf("KeyedDatapoint read error: %s",d2)
        return
    }

}

func TestFailure(t *testing.T) {
    buf := new(bytes.Buffer)
    _,err := ReadKeyedDatapoint(buf)
    if (err==nil) {
        t.Errorf("KeyedDatapoint fails to fail")
        return
    }
}
