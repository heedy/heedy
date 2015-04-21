package dbutil

import (
    "testing"
    )


func TestGetConversion(t *testing.T) {
    // note that v0 should include all subsequent upgrades, so we should be fine here.

    _, err := getConversion(POSTGRES, "", false)
    if err != nil {
        t.Errorf("could not compile postgres conversion 0, no drop %v", err.Error())
    }

    _, err = getConversion(POSTGRES, "", true)
    if err != nil {
        t.Errorf("could not compile postgres conversion 0, drop %v", err.Error())
    }

    _, err = getConversion(SQLITE3, "", false)
    if err != nil {
        t.Errorf("could not compile sqlite3 conversion 0, no drop %v", err.Error())
    }

    _, err = getConversion(SQLITE3, "", true)
    if err != nil {
        t.Errorf("could not compile sqlite3 conversion 0, drop %v", err.Error())
    }
}

func TestUpgradeDatabase(t *testing.T) {
    err := UpgradeDatabase("testing.sqlite3", true)

    if err != nil {
        t.Errorf("could not upgrade sqlite3 database: %v", err.Error())
    }
}
