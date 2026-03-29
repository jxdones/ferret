package cmd

import (
	"strings"

	"github.com/jxdones/ferret/internal/config"
	tui "github.com/jxdones/ferret/internal/tui/model"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "ferret",
	Short: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetBool("version")
		if version {
			PrintVersion()
			return nil
		}

		envName, _ := cmd.Flags().GetString("env")
		dir, _ := cmd.Flags().GetString("dir")
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		implicitDir := !cmd.Flags().Changed("dir")
		hasWorkspaces := len(cfg.Workspaces) > 0
		workspaceName := ""

		if implicitDir && hasWorkspaces {
			w := cfg.Workspaces[0]
			workspaceName = strings.TrimSpace(w.Name)
			if p := strings.TrimSpace(w.Path); p != "" {
				expanded, err := config.ExpandPath(p)
				if err != nil {
					return err
				}
				if expanded != "" {
					dir = expanded
				}
			}
		}

		return tui.Start(tui.StartOptions{
			Dir:                 dir,
			EnvName:             envName,
			ImplicitDirectory:   implicitDir,
			ConfigHasWorkspaces: hasWorkspaces,
			WorkspaceName:       workspaceName,
		})
	},
}

func init() {
	RootCmd.Flags().StringP("env", "e", "", `Load environments/<name>.yaml; if omitted, use first env file alphabetically when present, else shell-only`)
	RootCmd.Flags().StringP("dir", "d", ".", "Collection root directory")
	RootCmd.Flags().BoolP("version", "v", false, "Print the version number of ferret")
}
