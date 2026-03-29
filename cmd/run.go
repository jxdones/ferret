package cmd

import (
	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
	"github.com/jxdones/ferret/internal/exec"
	"github.com/jxdones/ferret/internal/render"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <path>",
	Short: "Run a request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		argEnv, err := cmd.Flags().GetString("env")
		if err != nil {
			return err
		}
		argDir, err := cmd.Flags().GetString("dir")
		if err != nil {
			return err
		}

		environment, err := env.Load(argDir, argEnv)
		if err != nil {
			return err
		}

		request, err := collection.LoadRequest(path)
		if err != nil {
			return err
		}

		result, err := exec.Execute(request, environment)
		if err != nil {
			return err
		}

		raw, err := cmd.Flags().GetBool("raw")
		if err != nil {
			return err
		}
		if raw {
			return render.RawBody(result, cmd.OutOrStdout())
		}
		return render.Response(result, cmd.OutOrStdout())
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("env", "e", "", "Name of environments/<name>.yaml (required; no auto-pick like the TUI)")
	runCmd.Flags().StringP("dir", "d", ".", "Collection root directory")
	runCmd.Flags().BoolP("raw", "r", false, "Print raw HTTP response (status line, headers, and body)")
}
