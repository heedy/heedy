/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbutil

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"config"

	log "github.com/Sirupsen/logrus"
	//"path/filepath"

	//"github.com/kardianos/osext"
)

const (
	versionString    = "DBVersion"
	defaultDbversion = "00000000"
)

// TODO @josephlewis42, add daniel's upgrade for version number on timebatch stuff.
// TODO @josephlewis42, add read markers to datastream on read push to aux table, on update aux delete old

// the database meta type
type meta struct {
	key   string
	value string
}

// Gets the conversion script for the given params.
func getConversion(dbtype string, dbversion string, dropOld bool) (string, error) {
	templateParams := make(map[string]string)

	if dbversion == "" {
		dbversion = defaultDbversion
	}

	templateParams["IDScramblePrime"] = strconv.FormatInt(config.Get().IDScramblePrime, 10)

	templateParams["DBVersion"] = dbversion
	templateParams["DBType"] = dbtype
	if dropOld {
		templateParams["Reset"] = "true"
		templateParams["DBVersion"] = defaultDbversion
	} else {
		templateParams["Reset"] = "false"
	}

	if dbtype == "postgres" {
		templateParams["pkey_exp"] = "SERIAL PRIMARY KEY"
	} else {
		templateParams["pkey_exp"] = "INTEGER PRIMARY KEY AUTOINCREMENT"
	}

	conversion_template, err := template.New("modifier").Parse(dbconversion)

	if err != nil {
		return "", err
	}

	var doc bytes.Buffer
	conversion_template.Execute(&doc, templateParams)

	return doc.String(), nil
}

/** Upgrades the database with the given connection string, returns an error if anything goes wrong.
**/
func UpgradeDatabase(cxnstring string, reset bool) error {

	db, driver, err := OpenSqlDatabase(cxnstring)
	if err != nil {
		return err
	}

	// Check version of database
	version := GetDatabaseVersion(db, driver)
	log.Printf("Upgrading DB From Version: %v", version)

	conversionstr, err := getConversion(driver, version, reset)

	if err != nil {
		return err
	}

	defer db.Close()
	_, err = db.Exec(conversionstr)

	if err != nil {
		fmt.Println(conversionstr)
		return err
	}

	return nil
}
