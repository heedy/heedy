package streamdb

import (
    "testing"
    "os"
    )

func TestDatabaseOpen(t *testing.T) {
    os.Remove("testing.db") //Delete sqlite database if exists
    db,err := Open("sqlite://testing.db","localhost:6379","localhost:4222")
    if err!=nil {
        t.Errorf("Could not open streamdb: %s",err)
    }
    defer db.Close()
}
