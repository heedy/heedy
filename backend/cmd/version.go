package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/heedy/heedy/backend/buildinfo"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows detailed version information",
	Long:  "Shows heedy's compilation and version details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Heedy v%s\n", buildinfo.Version)

		fmt.Printf(`
Built on:    %s
Git Hash:    %s
Go Version:  %s
Arch:        %s/%s
`, buildinfo.BuildTimestamp, buildinfo.GitHash, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		if verbose {
			bi, ok := debug.ReadBuildInfo()
			if ok {
				fmt.Println("\nBuild Deps:")
				for _, d := range bi.Deps {
					fmt.Printf("%s %s (%s)\n", d.Path, d.Version, d.Sum)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}
