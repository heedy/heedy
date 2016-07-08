/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package dbutil

/**
postgresify file provides the ability to convert queries done with question
marks into named queries with the proper query placeholders for postgres.
**/

import "strconv"

var (
	postgresQueryConversions = make(map[string]string)
)

func QueryConvert(query, dbtype string) string {
	switch dbtype {
	case "postgres":
		return QueryToPostgres(query)
	}

	return query
}

// Converts all ? in a query to $n which is the postgres format
func QueryToPostgres(query string) string {

	// cacheing
	q := postgresQueryConversions[query]
	if q != "" {
		return q
	}

	output := ""
	position := 1

	for _, runeValue := range query {

		if runeValue == '?' {
			output += "$"
			output += strconv.Itoa(position)
			position += 1
			continue
		}

		output += string(runeValue)
	}

	return output
}
