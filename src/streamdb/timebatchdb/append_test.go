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

    w.WriteBuffer(NewKeyedDatapoint("hello",1000,[]byte("Hello World!")))

    if (w.Len()<=0 || w.Size()>0 || r.Size() >0) {
        t.Errorf("writing commits when it shouldnt")
        return
    }

    w.WriteBuffer(NewKeyedDatapoint("hello2",2000,[]byte("Hello World2!")))
    if (w.Len()<=0 || w.Size()>0 || r.Size() >0) {
        t.Errorf("writing commits when it shouldnt")
        return
    }

    _,err = r.Next()
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
    w.WriteBuffer(NewKeyedDatapoint("hello3",3000,[]byte("Hello World3!")))

    d,err := r.Next()
    if (err!=nil || d.Key()!="hello" || string(d.Data())!="Hello World!" || d.Timestamp()!=1000) {
        t.Errorf("incorrect read: %s %s",d,err)
        return
    }
    d,err = r.Next()
    if (err!=nil || d.Key()!="hello2" || string(d.Data())!="Hello World2!" || d.Timestamp()!=2000) {
        t.Errorf("reader does not give correct result")
        return
    }
    _,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when at end of file")
        return
    }
    err = w.FlipWrite(0)
    if (err!=nil) {
        t.Errorf("FlipWrite failed: %s",err)
        return
    }
    d,err = r.Next()
    if (err!=nil || d.Key()!="hello3" || string(d.Data())!="Hello World3!" || d.Timestamp()!=3000) {
        t.Errorf("reader does not give correct result")
        return
    }
    _,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when at end of file")
        return
    }
    r.Reset()
    d,err = r.Next()
    if (err!=nil || d.Key()!="hello" || string(d.Data())!="Hello World!" || d.Timestamp()!=1000) {
        t.Errorf("reader does not give correct result")
        return
    }

    r.ToEnd()
    _,err = r.Next()
    if (err==nil) {
        t.Errorf("reader does not give error when at end of file")
        return
    }

}
