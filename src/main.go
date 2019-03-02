package main

import (
	"mime"

	"github.com/connectordb/connectordb/src/cmd"
)

func main() {
	mime.AddExtensionType(".jsm", "application/javascript")
	cmd.Execute()
}
