package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newTagCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Create, list, or delete git tags for skill versions",
	}
	cmd.AddCommand(newTagListCommand())
	cmd.AddCommand(newTagCreateCommand())
	cmd.AddCommand(newTagDeleteCommand())
	return cmd
}

func newTagListCommand() *cobra.Command {
	var repoPath string
	var pattern string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List git tags for skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			gitArgs := []string{"-C", abs, "tag", "--list"}
			if pattern != "" {
				gitArgs = append(gitArgs, pattern)
			}
			var out bytes.Buffer
			c := exec.Command("git", gitArgs...)
			c.Stdout = &out
			c.Stderr = &out
			if err := c.Run(); err != nil {
				return fmt.Errorf("git tag: %s", strings.TrimSpace(out.String()))
			}
			tags := strings.TrimSpace(out.String())
			if tags == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No tags found.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), tags)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&pattern, "pattern", "", "Shell glob pattern to filter tags (e.g. 'my-skill/*')")
	return cmd
}

func newTagCreateCommand() *cobra.Command {
	var repoPath string
	var tagPrefix string
	var bakeTarget string
	var skillFilter []string
	var dryRun bool
	var push bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create git tags for current skill versions",
		Long: `Reads each skill's version from its SKILL.md frontmatter and creates a git tag
in the form '<skill-name>/<prefix><version>' (e.g. 'my-skill/v1.2.0').

Pass --skill to tag specific skills only.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			// Collect skills.
			wanted := map[string]bool{}
			for _, s := range skillFilter {
				wanted[s] = true
			}

			var skills []string
			seen := map[string]struct{}{}
			t := bakeTarget
			if t == "" {
				t = "default"
			}
			tc := cfg.Targets[t]
			if tc == nil {
				return fmt.Errorf("target %q not found in bakefile", t)
			}
			for _, s := range tc.Skills {
				if _, dup := seen[s]; dup {
					continue
				}
				seen[s] = struct{}{}
				if len(wanted) == 0 || wanted[s] {
					skills = append(skills, s)
				}
			}

			created := 0
			for _, skillName := range skills {
				skillMD := filepath.Join(abs, sourceBase, skillName, "SKILL.md")
				raw, err := os.ReadFile(skillMD)
				data := string(raw)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  !  skip %s: %v\n", skillName, err)
					continue
				}
				fm, ok := parseFrontmatterDates(data)
				if !ok || fm.Version == "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  !  skip %s: no version in frontmatter\n", skillName)
					continue
				}
				tagName := skillName + "/" + tagPrefix + fm.Version
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "  ~  would create tag %s\n", tagName)
					continue
				}
				var errBuf bytes.Buffer
				c := exec.Command("git", "-C", abs, "tag", tagName)
				c.Stderr = &errBuf
				if err := c.Run(); err != nil {
					msg := strings.TrimSpace(errBuf.String())
					if strings.Contains(msg, "already exists") {
						fmt.Fprintf(cmd.OutOrStdout(), "  -  %s: tag already exists\n", tagName)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  !  %s: %s\n", tagName, msg)
					}
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓  created tag %s\n", tagName)
				if push {
					pushOut, pushErr := exec.Command("git", "-C", abs, "push", "origin", tagName).CombinedOutput()
					if pushErr != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "  !  push %s: %s\n", tagName, strings.TrimSpace(string(pushOut)))
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  ✓  pushed %s\n", tagName)
					}
				}
				created++
			}

			if !dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\nCreated %d tag(s)\n", created)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&tagPrefix, "prefix", "v", "Version prefix for tags")
	cmd.Flags().StringVar(&bakeTarget, "bake-target", "", "Bakefile target (default: default)")
	cmd.Flags().StringArrayVar(&skillFilter, "skill", nil, "Only tag these skills (repeatable)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without creating tags")
	cmd.Flags().BoolVar(&push, "push", false, "Push created tags to origin")
	return cmd
}

func newTagDeleteCommand() *cobra.Command {
	var repoPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "delete <tag...>",
		Short: "Delete one or more git tags",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			for _, tag := range args {
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "  ~  would delete tag %s\n", tag)
					continue
				}
				var errBuf bytes.Buffer
				c := exec.Command("git", "-C", abs, "tag", "-d", tag)
				c.Stderr = &errBuf
				if err := c.Run(); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  !  %s: %s\n", tag, strings.TrimSpace(errBuf.String()))
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓  deleted tag %s\n", tag)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without deleting")
	return cmd
}
