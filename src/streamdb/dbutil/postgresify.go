package dbutil

/**
postgresify file provides the ability to convert queries done with question
marks into named queries with the proper query placeholders for postgres.
**/

import (
    "strconv"
)


// Converts all ? in a query to $n which is the postgres format
func QueryToPostgres(query string) string {
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
