package dbutil

import (
	"database/sql"
	"text/template"
	"bytes"
	"log"
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
func GetConversion(dbtype DRIVERSTR, dbversion string, dropOld bool) string {
	templateParams := make(map[string] string)

	if dbversion == "" {
		dbversion = "00000000"
	}

	templateParams["DBVersion"] = dbversion
	templateParams["DBType"] = dbtype.String()
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


	conversion_template, err := template.New("modifier").Parse(dbconversion)

	if err != nil {
		panic(err.Error())
	}

	var doc bytes.Buffer
	conversion_template.Execute(&doc, templateParams)
    return doc.String()
}

func DoConversion(db *sql.DB, dbtype DRIVERSTR, deleteold bool) error {
	version := "00000000"

	var mixin SqlxMixin
	mixin.InitSqlxMixin(db, dbtype.String())

	err := mixin.Get(&version, "SELECT Value FROM StreamdbMeta WHERE Key = 'DBVersion'")

	if err != nil {
		version = "00000000"
	}

	conversionstr := GetConversion(dbtype, version, deleteold)

	log.Printf("Conversion string\n\n%v", conversionstr)

	_, err = mixin.Exec(conversionstr)
	return err
}
