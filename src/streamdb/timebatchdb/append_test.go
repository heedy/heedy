package timebatchdb

import (
    "testing"
    "os"
    )

func TestAppendReadWrite(t *testing.T) {
    os.RemoveAll("testdatabase/append.0")

    w,err := NewAppendWriter("testdatabase",0)
    if (err!=nil) {
        t.Errorf("Error opening file: %s",err)
        return
    }
    defer w.Close()

    r,err := NewAppendReader("testdatabase",0)
    if (err!=nil) {
        t.Errorf("Error opening file: %s",err)
        return
    }
    defer r.Close()

    //Start off by making sure that there is nothing to read
    if (r.Size()!=0 || w.Size()!=0) {
        t.Errorf("Failure to length")
        return
    }

    if (w.Len()!=0) {
        t.Errorf("length of empty writer dun goofed")
        return
    }

    w.WriteBuffer("hello",1000,[]byte("Hello World!"))

    if (w.Len()<=0 || w.Size()>0 || r.Size() >0) {
        t.Errorf("writing commits when it shouldnt")
        return
    }

    w.WriteBuffer("hello2",2000,[]byte("Hello World2!"))
    if (w.Len()<=0 || w.Size()>0 || r.Size() >0) {
        t.Errorf("writing commits when it shouldnt")
        return
    }

    _,_,_,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when empty")
        return
    }


    err = w.FlipWrite(0)
    if (err!=nil) {
        t.Errorf("FlipWrite failed: %s",err)
        return
    }

    if (w.Len()>0 || w.Size()!=r.Size()) {
        t.Errorf("incorrect sizes returned")
        return
    }
    w.WriteBuffer("hello3",3000,[]byte("Hello World3!"))

    key,time,data,err := r.Next()
    if (err!=nil || key!="hello" || string(data)!="Hello World!" || time!=1000) {
        t.Errorf("incorrect read: k=%s t=%d d=%s e=%s",key,time,string(data),err)
        return
    }
    key,time,data,err = r.Next()
    if (err!=nil || key!="hello2" || string(data)!="Hello World2!" || time!=2000) {
        t.Errorf("reader does not give correct result")
        return
    }
    _,_,_,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when at end of file")
        return
    }
    err = w.FlipWrite(0)
    if (err!=nil) {
        t.Errorf("FlipWrite failed: %s",err)
        return
    }
    key,time,data,err = r.Next()
    if (err!=nil || key!="hello3" || string(data)!="Hello World3!" || time!=3000) {
        t.Errorf("reader does not give correct result")
        return
    }
    _,_,_,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when at end of file")
        return
    }
    r.Reset()
    key,time,data,err = r.Next()
    if (err!=nil || key!="hello" || string(data)!="Hello World!" || time!=1000) {
        t.Errorf("reader does not give correct result")
        return
    }

    r.ToEnd()
    _,_,_,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when at end of file")
        return
    }

}
