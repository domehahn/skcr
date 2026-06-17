package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/internal/skillversion"
	"github.com/spf13/cobra"
)

func newSkillVersionsCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "skill-versions <skill>",
		Short: "Print the version history from a skill's SKILL.md changelog",
		Long: `Reads the changelog: block from <skill>/SKILL.md frontmatter and prints
each release entry (version, date, change) in reverse chronological order.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			skillDir := filepath.Join(abs, sourceBase, skillName)

			entries, err := skillversion.Changelog(skillDir)
			if err != nil {
				return fmt.Errorf("read changelog: %w", err)
			}

			if len(entries) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No changelog entries found for %s.\n", skillName)
				return nil
			}

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(entries)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Version history: %s\n\n", skillName)
			for _, e := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s\n", e.Version, e.Date)
				if e.Change != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", e.Change)
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d release(s)\n", len(entries))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
