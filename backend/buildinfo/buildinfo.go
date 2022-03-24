package buildinfo

import "time"

// These variables are set during linking

var Version = "0.0.0"
var BuildTimestamp string
var GitHash string

// Allow the server to be put into dev mode for building plugins.
var DevMode = false

var StartTime time.Time

func init() {
	StartTime = time.Now()
}
