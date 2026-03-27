package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "ferret",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.Flags().StringP("env", "e", "", "Environment to use (e.g. dev, prod)")
	RootCmd.Flags().StringP("dir", "d", ".", "Collection root directory")
}
