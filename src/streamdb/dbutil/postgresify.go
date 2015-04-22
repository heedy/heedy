package dbutil

/**
postgresify file provides the ability to convert queries done with question
marks into named queries with the proper query placeholders for postgres.
**/

import (
    "strconv"
    "os/exec"
    "fmt"
    "strings"
)

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

// finds the postgres binary on the system, isn't very robust in checking though
// should work on ubuntu variants and when postgres is on $PATH
func FindPostgres() string {
    return findPostgresExecutable("postgres")
}


// finds the postgres init binary on the system, isn't very robust in checking though
// should work on ubuntu variants and when postgres is on $PATH
func FindPostgresInit() string {
    return findPostgresExecutable("initdb")
}

// finds the postgres psql binary on the system, isn't very robust in checking though
// should work on ubuntu variants and when postgres is on $PATH
func FindPostgresPsql() string {
    return findPostgresExecutable("psql")
}

func trimExecutablePath(exepath string) string {
    return strings.Trim(exepath, " \t\n\r")
}

func findPostgresExecutable(executableName string) string {
    // Start with which because we prefer a PATH version
    out := findPostgresExecutableWhich(executableName)

    if out != "" {
        return trimExecutablePath(out)
    }

    return trimExecutablePath(findPostgresExecutableGrep(executableName))
}

// Find a postgres utility e.g. initdb or postgres using the lame grep method, works on Ubuntu (for now)
func findPostgresExecutableGrep(executableName string) string {

    findCmd := fmt.Sprintf("find /usr/lib/postgresql/ | sort -r | grep -m 1 /bin/%v", executableName)

    cmd := exec.Command("bash", "-c", findCmd)
    out, err := cmd.CombinedOutput()

    if err != nil {
        return ""
    }

    return string(out)
}


// Finds a utility on $PATH
func findPostgresExecutableWhich(executableName string) string {
    cmd := exec.Command("which", executableName)
    out, err := cmd.CombinedOutput()

    if err != nil {
        return ""
    }

    return string(out)
}
