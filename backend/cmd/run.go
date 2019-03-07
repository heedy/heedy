package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/server"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run [location of database]",
	Short: "Runs heedy",
	Long:  `Runs heedy using the passed database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return ErrTooManyArgs
		}
		if len(args) != 1 {
			return errors.New("Must specify a database location")
		}
		directory := args[0]
		a, err := assets.Open(directory, nil)
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
