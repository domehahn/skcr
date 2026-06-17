package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newPinCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "pin <skill> <version>",
		Short: "Pin a skill's version in its SKILL.md frontmatter",
		Long: `Updates the version: field in <skill>/SKILL.md to exactly <version>.
This records a canonical version for the skill and is read by 'skcr changelog',
'skcr show', and 'skcr publish'.

Run 'skcr bake' after pinning to regenerate rendered files.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, version := args[0], args[1]

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

			content, err := cliReadFile(skillMD)
			if err != nil {
				return fmt.Errorf("read %s: %w", skillMD, err)
			}

			updated, changed := setFrontmatterField(string(content), "version", version)
			if !changed {
				fmt.Fprintf(cmd.OutOrStdout(), "%s already has version: %s\n", skillName, version)
				return nil
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would set %s version: %s\n", skillName, version)
				return nil
			}

			if err := cliWriteFile(skillMD, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", skillMD, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pinned %s to version %s\n", skillName, version)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
