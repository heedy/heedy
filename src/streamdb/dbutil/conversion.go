package dbutil

import (
	//"database/sql"
	"text/template"
	"os"
	"fmt"
    )

const (
	versionString = "DBVersion"
)

// TODO @josephlewis42, add daniel's upgrade for version number on timebatch stuff.
// TODO @josephlewis42, add read markers to timebatchdb on read push to aux table, on update aux delete old

// the database meta type
type meta struct {
	key string
	value string
}

// Gets the conversion script for the given params.
func GetConversion(dbtype, dbversion string, dropOld bool) string {
	templateParams := make(map[string] string)

	if dbversion == "" {
		dbversion = "00000000"
	}

	templateParams["DBVersion"] = dbversion
	templateParams["DBType"] = dbtype
	if dropOld {
		templateParams["DroppingTables"] = "true"
	} else {
		templateParams["DroppingTables"] = "false"
	}

	if dbversion == "postgres" {
		templateParams["pkey_exp"] = "SERIAL PRIMARY KEY"
	} else {
		templateParams["pkey_exp"] = "INTEGER PRIMARY KEY"
	}


	conversion_template := template.Must(template.ParseFiles("conversion.temp.sql"))

	err := conversion_template.Execute(os.Stdout, templateParams)
	if err != nil { panic(err) }
    return ""
}



func main() {
	GetConversion("sqlite3", "00000000", true)
	fmt.Printf("done")
}
