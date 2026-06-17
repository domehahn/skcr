package cli

import (
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/internal/models"
	"github.com/spf13/cobra"
)

func newRenameTargetToplevelCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "rename-target <old> <new>",
		Short: "Rename a bakefile target and update inherits: references",
		Long: `Renames a target key in agentic.bake.yaml and rewrites any other target's
inherits: list that references the old name.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName, newName := args[0], args[1]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			if _, ok := cfg.Targets[oldName]; !ok {
				return fmt.Errorf("target %q does not exist in bakefile", oldName)
			}
			if _, exists := cfg.Targets[newName]; exists {
				return fmt.Errorf("target %q already exists in bakefile", newName)
			}

			renamed := renameTargetInConfig(cfg, oldName, newName)
			if err := cliDumpBakeFile(cfg, filepath.Join(abs, "agentic.bake.yaml")); err != nil {
				return fmt.Errorf("write bakefile: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Renamed target %q → %q", oldName, newName)
			if renamed > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), " (%d inherits: reference(s) updated)", renamed)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	return cmd
}

// renameTargetInConfig moves cfg.Targets[oldName] → cfg.Targets[newName] and
// rewrites every inherits: list that references oldName. Returns the count of
// inherits: entries rewritten.
func renameTargetInConfig(cfg *models.BakeConfig, oldName, newName string) int {
	tc := cfg.Targets[oldName]
	delete(cfg.Targets, oldName)
	cfg.Targets[newName] = tc

	rewritten := 0
	for _, t := range cfg.Targets {
		for i, inh := range t.Inherits {
			if inh == oldName {
				t.Inherits[i] = newName
				rewritten++
			}
		}
	}
	return rewritten
}
