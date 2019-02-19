package main

import (
	"mime"

	"github.com/connectordb/connectordb/cmd"
)

func main() {
	mime.AddExtensionType(".jsm", "application/javascript")
	cmd.Execute()
}
