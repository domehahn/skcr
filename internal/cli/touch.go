package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func newTouchCommand() *cobra.Command {
	var target string
	var date string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "touch <skill>",
		Short: "Update last_modified: in a skill's SKILL.md to today",
		Long: `Sets the last_modified: field in the SKILL.md frontmatter of <skill> to
today's date (YYYY-MM-DD). Useful after manual edits that don't warrant a
full version bump.`,
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
			skillMD := filepath.Join(abs, sourceBase, skillName, "SKILL.md")

			raw, err := cliReadFile(skillMD)
			if err != nil {
				return fmt.Errorf("read SKILL.md: %w", err)
			}

			today := date
			if today == "" {
				today = time.Now().UTC().Format("2006-01-02")
			}

			updated, changed := setFrontmatterField(string(raw), "last_modified", today)
			if !changed {
				fmt.Fprintf(cmd.OutOrStdout(), "%s already has last_modified: %s\n", skillName, today)
				return nil
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would set %s last_modified: %s\n", skillName, today)
				return nil
			}

			if err := cliWriteFile(skillMD, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("write SKILL.md: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Touched %s (last_modified: %s)\n", skillName, today)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().StringVar(&date, "date", "", "Override date (YYYY-MM-DD); defaults to today")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
