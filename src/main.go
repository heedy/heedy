package main

import (
	"mime"

	"github.com/connectordb/connectordb/src/cmd"
)

func main() {
	mime.AddExtensionType(".mjs", "application/javascript")
	cmd.Execute()
}
