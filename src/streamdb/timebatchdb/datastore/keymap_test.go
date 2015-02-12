package datastore

import (
    "testing"
    "os"
    )

func TestKeyMap(t *testing.T) {
    os.RemoveAll("testdatabase/keys")
    m1,err := OpenKeyMap("testdatabase")
    if (err!=nil) {
        t.Errorf("Error opening file: %s",err)
        return
    }
    defer m1.Close()

    if (m1.Len()!=0) {
        t.Errorf("Incorrect length: %d",m1.Len())
        return
    }

    //Now test for existence of key
    if (m1.Get("hello/world")!=0) {
        t.Errorf("key exists where it shouldn't: %d",m1.Len())
        return
    }

    if (1!=m1.Map("hello/world")) {
        t.Errorf("key mapping failed: %d",m1.Get("hello/world"))
        return
    }
    if (2!=m1.Map("hello/world2")) {
        t.Errorf("key mapping failed: %d",m1.Get("hello/world2"))
        return
    }
    if (1!=m1.Map("hello/world")) {
        t.Errorf("key mapping didn't keep: %d",m1.Get("hello/world"))
        return
    }
    val,err := m1.Create("hello/world3")
    if (3!=val || err!=nil) {
        t.Errorf("key mapping failed: %d (%s)",m1.Get("hello/world3"),err)
        return
    }

    m2,err := OpenKeyMap("testdatabase")
    if (err!=nil) {
        t.Errorf("Error opening file: %s",err)
        return
    }
    defer m2.Close()
    if (2!=m2.Map("hello/world2")) {
        t.Errorf("key mapping failed: %d",m2.Get("hello/world2"))
        return
    }
    if (3!=m2.Get("hello/world3")) {
        t.Errorf("key mapping failed: %d",m2.Get("hello/world3"))
        return
    }

    if (4!=m2.Map("woo")) {
        t.Errorf("key mapping failed: %d",m2.Get("woo"))
        return
    }

    err = m1.Reload()
    if err!=nil {
        t.Errorf("reload failed: %d",m2.Get("woo"))
    }
    if (4!=m2.Get("woo")) {
        t.Errorf("key mapping failed: %d",m2.Get("woo"))
        return
    }

}
