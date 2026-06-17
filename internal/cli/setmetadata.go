package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newSetMetadataCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "set-metadata <skill> <key> <value>",
		Short: "Set an arbitrary scalar field in a skill's SKILL.md frontmatter",
		Long: `Sets <key>: <value> in the YAML frontmatter of <skill>/SKILL.md.
Works for any scalar field: author, owner, status, stability, tags, etc.

Use 'skcr pin' for version pinning and 'skcr touch' for last_modified — this
command is the general form for every other frontmatter field.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, key, value := args[0], args[1], args[2]

			if key == "" || strings.ContainsAny(key, "\n\r:") || strings.TrimSpace(key) != key {
				return fmt.Errorf("invalid key %q: must not be empty, contain newlines, colons, or surrounding whitespace", key)
			}
			if err := validateSkillArg(skillName); err != nil {
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
			skillMD := filepath.Join(abs, sourceBase, skillName, "SKILL.md")

			data, err := cliReadFile(skillMD)
			if err != nil {
				return fmt.Errorf("read SKILL.md: %w", err)
			}

			if !strings.HasPrefix(string(data), "---\n") {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s has no frontmatter — skipping\n",
					filepath.Join(sourceBase, skillName, "SKILL.md"))
				return nil
			}
			newContent, changed := setFrontmatterField(string(data), key, value)
			if !changed {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s is already %q\n", skillName, key, value)
				return nil
			}

			rel := filepath.Join(sourceBase, skillName, "SKILL.md")
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would set %s: %s = %q in %s\n", skillName, key, value, rel)
				return nil
			}

			if err := cliWriteFile(skillMD, []byte(newContent), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", skillMD, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set %s: %s = %q\n", skillName, key, value)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}
