package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newRevertCommand() *cobra.Command {
	var repoPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "revert <skill> <ref>",
		Short: "Revert a SKILL.md to its state at a given git ref",
		Long: `Fetches the SKILL.md content of <skill> at <ref> (a commit, tag, or branch)
using 'git show' and overwrites the current file.

Examples:
  skcr revert my-skill HEAD~1
  skcr revert my-skill v1.2.0
  skcr revert my-skill abc1234`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, ref := args[0], args[1]

			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			relPath := filepath.Join(sourceBase, skillName, "SKILL.md")
			destPath := filepath.Join(abs, relPath)

			content, err := gitShow(abs, ref, relPath)
			if err != nil {
				return fmt.Errorf("cannot read %s at %s: %w", relPath, ref, err)
			}

			// Read current content to check if it differs.
			current, readErr := os.ReadFile(destPath)
			if readErr == nil && string(current) == content {
				fmt.Fprintf(cmd.OutOrStdout(), "  -  %s: already matches %s\n", skillName, ref)
				return nil
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would revert %s to %s\n", skillName, ref)
				return nil
			}

			if err := os.WriteFile(destPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write SKILL.md: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Reverted %s to %s\n", skillName, ref)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
