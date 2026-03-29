package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ferret",
	Run: func(cmd *cobra.Command, args []string) {
		PrintVersion()
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func PrintVersion() {
	fmt.Println("ferret v" + version)
}
