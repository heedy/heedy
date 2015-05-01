package dbutil

import (
	"testing"
	"os"
)

func TestDriverStrString(t *testing.T) {

	if SQLITE3.String() != "sqlite3" {
        t.Errorf("could not compare driver string")
	}
}

func TestUriIsSqlite(t *testing.T) {
    if ! UriIsSqlite("testing.db") {
        t.Errorf("URI should be sqlite")
    }
    if ! UriIsSqlite("testing.sqlite") {
        t.Errorf("URI should be sqlite")
    }

    if ! UriIsSqlite("testing.sqlite3") {
        t.Errorf("URI should be sqlite")
    }

    if ! UriIsSqlite("sqlite://testing/foo/bar/baz") {
        t.Errorf("URI should be sqlite")
    }

    if UriIsSqlite("/testing/foo/bar/baz") {
        t.Errorf("URI should not be sqlite")
    }

    if UriIsSqlite("postgres:///testing/foo/bar/baz") {
        t.Errorf("URI should not be sqlite")
    }
}


func TestSqliteURIToPath(t *testing.T){

    if SqliteURIToPath("/foo/bar/baz/sqlite.db") != "/foo/bar/baz/sqlite.db" {
        t.Errorf("URI tampered with")
    }

    if SqliteURIToPath("sqlite:///foo/bar/baz/sqlite.db") != "/foo/bar/baz/sqlite.db" {
        t.Errorf("URI not tampered with")
    }
}


// From a connection string, gets the cleaned connection path and database type
func TestProcessConnectionString(t *testing.T){

    _, driver := ProcessConnectionString("test.sqlite3")

    if driver != SQLITE3 {
        t.Errorf("URI should be sqlite")
    }

    _, driver = ProcessConnectionString("postgres://foobar.com")
    if driver != POSTGRES {
        t.Errorf("URI should be postgres")
    }
}

func TestSqlOpen(t *testing.T) {
	filepath := "testing_testsqlopen.sqlite3"
	defer os.Remove(filepath)

	db, ds, err := OpenSqlDatabase(filepath)
	if err != nil {
        t.Errorf(err.Error())
		return
	}

	defer db.Close()

	version := GetDatabaseVersion(db, ds)

	if version != "00000000" {
        t.Errorf("Wrong version gotten for empty database")
	}
}
