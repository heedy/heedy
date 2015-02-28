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

    if (d3.String()!=d.String()) {
        t.Errorf("Datapoint string error: %s",d3)
        return
    }

}

func TestLargeDatapoint(t *testing.T) {
    datastring := "Hello World"

    for i :=0 ; i< 10000; i++ {
        datastring = datastring + "HelloWorld"
    }
    d := NewDatapoint(1337,[]byte(datastring))
    if (d.Len()!=len(d.Bytes()) || d.Timestamp()!=1337 || string(d.Data())!=datastring || d.DataLen()!=len(datastring)) {
        t.Errorf("Datapoint read error: %s",d)
        return
    }
}


func BenchmarkDatapointCreate(b *testing.B) {
    //Create the datapoint and get the Bytes from it
    for n := 0; n < b.N; n++ {
        d := NewDatapoint(1337,[]byte("Hello World!"))
        d.Bytes()
    }
}

func BenchmarkDatapointRead(b *testing.B) {
    //Read stuff from the datapoint
    d := NewDatapoint(1337,[]byte("Hello World!"))
    for n := 0; n < b.N; n++ {
        d.Timestamp()
        d.Data()
    }
}
