package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/connectordb/connectordb/assets"
	"github.com/connectordb/connectordb/server"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [location of database]",
	Short: "Runs connectorDB",
	Long:  `Runs ConnectorDB using the passed database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return ErrTooManyArgs
		}
		if len(args) != 1 {
			return errors.New("Must specify a database location")
		}
		directory := args[0]
		a, err := assets.Load(directory, nil)
		if err != nil {
			return err
		}
		assets.SetGlobal(a)

		b, err := json.MarshalIndent(a.Config, "", " ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))

		return server.Run(a)
	},
}

func init() {
	RootCmd.AddCommand(RunCmd)
}
