package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/afero"

	"github.com/connectordb/connectordb/assets"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func test() {
	a, err := assets.NewAssets("./test", nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	b, err := json.MarshalIndent(a.Config, "", " ")
	if err != nil {
		log.Error(err.Error())
		return
	}
	fmt.Println(string(b))
	/*
		f, err := a.AssetFS.Open("/setup/app.css")
		if err != nil {
			log.Error(err.Error())
			return
		}

		finfo, err := f.Stat()
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Println(finfo.Name())

		buf := new(bytes.Buffer)

		buf.ReadFrom(f)
		fmt.Println(buf.String())
	*/

	http.Handle("/", http.FileServer(afero.NewHttpFs(a.AssetFS)))
	http.ListenAndServe(":3000", nil)
}

var rootCmd = &cobra.Command{
	Use:   "connectordb",
	Short: "ConnectorDB is a repository for your quantified-self and IoT data",
	Long:  `ConnectorDB is a database built for interacting with your IoT devices and for storing your quantified-self data.`,
	Run: func(cmd *cobra.Command, args []string) {
		test()
		cmd.HelpFunc()(cmd, args)
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
