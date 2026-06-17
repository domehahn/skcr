package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newCopySkillCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "copy-skill <source> <new-name>",
		Short: "Duplicate a skill directory under a new name",
		Long: `Copies <source> to a new directory named <new-name> inside
skill_sources.output_dir and updates the name: field in the copy's SKILL.md.

Useful when creating a variant of an existing skill without starting from scratch.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceName, newName := args[0], args[1]

			if err := validateSkillArg(sourceName); err != nil {
				return err
			}
			if err := validateSkillArg(newName); err != nil {
				return err
			}

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			sourceDir := filepath.Join(abs, sourceBase, sourceName)
			destDir := filepath.Join(abs, sourceBase, newName)

			if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
				return fmt.Errorf("skill %q not found in %s", sourceName, sourceBase)
			}
			if _, err := os.Stat(destDir); err == nil {
				return fmt.Errorf("skill %q already exists", newName)
			}

			rel := filepath.Join(sourceBase, newName)
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would copy %s → %s\n", sourceName, rel)
				return nil
			}

			if err := copyDir(sourceDir, destDir); err != nil {
				return fmt.Errorf("copy: %w", err)
			}

			// Update name: in the copy's SKILL.md. Roll back the copy on failure.
			skillMD := filepath.Join(destDir, "SKILL.md")
			data, err := cliReadFile(skillMD)
			if err != nil {
				_ = os.RemoveAll(destDir)
				return fmt.Errorf("read source SKILL.md: %w", err)
			}
			updated, _ := setFrontmatterField(string(data), "name", newName)
			if err := cliWriteFile(skillMD, []byte(updated), 0o644); err != nil {
				_ = os.RemoveAll(destDir)
				return fmt.Errorf("update SKILL.md: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Copied %s → %s\n", sourceName, rel)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
