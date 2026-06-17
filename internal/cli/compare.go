package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newCompareCommand() *cobra.Command {
	var repoPath string
	var format string

	cmd := &cobra.Command{
		Use:   "compare <skill> <ref1> <ref2>",
		Short: "Show a diff of SKILL.md between two git refs",
		Long: `Compares the SKILL.md of a skill at two git refs (commits, tags, or branches).

Examples:
  skcr compare my-skill HEAD~1 HEAD
  skcr compare my-skill v1.0.0 v2.0.0
  skcr compare my-skill main feature/new-tool`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName, ref1, ref2 := args[0], args[1], args[2]

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

			content1, err := gitShow(abs, ref1, relPath)
			if err != nil {
				return fmt.Errorf("cannot read %s at %s: %w", relPath, ref1, err)
			}
			content2, err := gitShow(abs, ref2, relPath)
			if err != nil {
				return fmt.Errorf("cannot read %s at %s: %w", relPath, ref2, err)
			}

			if content1 == content2 {
				fmt.Fprintf(cmd.OutOrStdout(), "No differences in %s between %s and %s\n", skillName, ref1, ref2)
				return nil
			}

			switch strings.ToLower(format) {
			case "git":
				diff, diffErr := gitDiffStrings(abs, ref1, ref2, relPath)
				if diffErr == nil && diff != "" {
					fmt.Fprint(cmd.OutOrStdout(), diff)
					return nil
				}
				// Fall through to built-in diff
				fallthrough
			default:
				fmt.Fprint(cmd.OutOrStdout(), unifiedDiff(content1, content2, relPath, false))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&format, "format", "unified", "Diff format: unified or git")
	return cmd
}

// gitShow returns the content of a file at a given git ref.
func gitShow(repoPath, ref, relPath string) (string, error) {
	spec := ref + ":" + relPath
	var out bytes.Buffer
	var errBuf bytes.Buffer
	c := exec.Command("git", "-C", repoPath, "show", spec)
	c.Stdout = &out
	c.Stderr = &errBuf
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(errBuf.String()))
	}
	return out.String(), nil
}

// gitDiffStrings runs git diff between two refs for a specific file.
func gitDiffStrings(repoPath, ref1, ref2, relPath string) (string, error) {
	var out bytes.Buffer
	c := exec.Command("git", "-C", repoPath, "diff", ref1, ref2, "--", relPath)
	c.Stdout = &out
	if err := c.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}
