/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbutil

/**
postgresify file provides the ability to convert queries done with question
marks into named queries with the proper query placeholders for postgres.
**/

import (
	"testing"
)

// Converts all ? in a query to $n which is the postgres format
func TestQueryToPostgres(t *testing.T) {
	query1 := "SELECT * FROM Users;"
	query2 := "SELECT * FROM Users WHERE ? = 1;"
	answer2 := "SELECT * FROM Users WHERE $1 = 1;"
	query3 := "INSERT INTO Users VALUES (?,?,?,?,?)"
	answer3 := "INSERT INTO Users VALUES ($1,$2,$3,$4,$5)"

	a1 := QueryToPostgres(query1)
	if a1 != query1 {
		t.Errorf("Expected input: %v, output: %v, got: %v", query1, query1, a1)
	}

	a2 := QueryToPostgres(query2)
	if a2 != answer2 {
		t.Errorf("Expected the same: %v, %v", query2, answer2, a2)
	}

	a3 := QueryToPostgres(query3)
	if a3 != answer3 {
		t.Errorf("Expected the same: %v, %v", query3, answer3, a3)
	}
}

// Converts all ? in a query to $n which is the postgres format
func TestQueryConvert(t *testing.T) {
	query3 := "INSERT INTO Users VALUES (?,?,?,?,?)"
	answer3 := "INSERT INTO Users VALUES ($1,$2,$3,$4,$5)"

	a1 := QueryConvert(query3, POSTGRES)
	if a1 != answer3 {
		t.Errorf("Expected input: %v, output: %v, got: %v", query3, answer3, a1)
	}
}

func TestFindPostgres(t *testing.T) {
	ppath := FindPostgres()

	if ppath == "" {
		t.Errorf("could not find postgres")
	}
}
