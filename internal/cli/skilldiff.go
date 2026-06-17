package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newSkillDiffCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "skill-diff <skill> <ref>",
		Short: "Diff a skill's current SKILL.md against a git ref",
		Long: `Fetches <skill>/SKILL.md at <ref> using git and diffs it against the
working-tree version. Useful for reviewing what changed in a skill since a
specific commit, tag, or branch.

Example:
  skcr skill-diff my-skill HEAD~1
  skcr skill-diff my-skill v1.2.0`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, ref := args[0], args[1]

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
			relPath := filepath.Join(sourceBase, skillName, "SKILL.md")

			historical, err := gitShow(abs, ref, relPath)
			if err != nil {
				return fmt.Errorf("git show %s:%s: %w", ref, relPath, err)
			}

			current, err := cliReadFile(filepath.Join(abs, relPath))
			if err != nil {
				return fmt.Errorf("read current SKILL.md: %w", err)
			}

			if historical == string(current) {
				fmt.Fprintf(cmd.OutOrStdout(), "No differences in %s since %s\n", skillName, ref)
				return nil
			}

			diff := unifiedDiff(historical, string(current), relPath, false)
			// Replace the generic a/b prefixes with meaningful labels.
			diff = strings.Replace(diff, "a/"+relPath, skillName+"@"+ref+"/SKILL.md", 1)
			diff = strings.Replace(diff, "b/"+relPath, skillName+"/SKILL.md (working tree)", 1)
			fmt.Fprint(cmd.OutOrStdout(), diff)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	return cmd
}
