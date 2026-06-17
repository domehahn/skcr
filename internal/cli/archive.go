package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newArchiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive or restore skills in the bakefile",
	}
	cmd.AddCommand(newArchiveSkillCommand())
	cmd.AddCommand(newRestoreSkillCommand())
	return cmd
}

func newArchiveSkillCommand() *cobra.Command {
	var repoPath string
	var change string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Mark a skill deprecated and remove it from all bakefile targets",
		Long: `Sets stability: deprecated in SKILL.md frontmatter, removes the skill from
every bakefile target, and appends a CHANGELOG entry.

The skill directory is NOT deleted — use 'skcr remove skill --delete-dirs' for that.
Restore with 'skcr archive restore <name>'.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			skillMDPath := filepath.Join(abs, sourceBase, name, "SKILL.md")

			data, readErr := os.ReadFile(skillMDPath)
			if readErr != nil {
				return fmt.Errorf("SKILL.md not found for %q: %w", name, readErr)
			}

			updated, changed := setFrontmatterField(string(data), "stability", "deprecated")
			if !changed {
				fmt.Fprintf(cmd.OutOrStdout(), "  -  %s: already deprecated\n", name)
			}

			// Remove from all targets.
			removed := 0
			for tName, t := range cfg.Targets {
				before := len(t.Skills)
				t.Skills = filterSlice(t.Skills, name)
				if len(t.Skills) < before {
					cfg.Targets[tName] = t
					removed++
				}
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would archive %q (removed from %d target(s))\n", name, removed)
				return nil
			}

			if changed {
				if err := os.WriteFile(skillMDPath, []byte(updated), 0o644); err != nil {
					return fmt.Errorf("write SKILL.md: %w", err)
				}
			}

			if err := cliDumpBakeFile(cfg, filepath.Join(abs, "agentic.bake.yaml")); err != nil {
				return fmt.Errorf("update bakefile: %w", err)
			}

			if err := appendChangelogEntry(filepath.Join(abs, sourceBase, name), change); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  !  could not update CHANGELOG.md: %v\n", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Archived %q — removed from %d target(s), stability set to deprecated\n", name, removed)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&change, "change", "Archived skill", "CHANGELOG entry message")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without modifying files")
	return cmd
}

func newRestoreSkillCommand() *cobra.Command {
	var repoPath string
	var addToTarget string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "restore <name>",
		Short: "Un-archive a skill: remove deprecated status and add back to a target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			abs, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			skillMDPath := filepath.Join(abs, sourceBase, name, "SKILL.md")

			data, readErr := os.ReadFile(skillMDPath)
			if readErr != nil {
				return fmt.Errorf("SKILL.md not found for %q: %w", name, readErr)
			}

			updated, _ := setFrontmatterField(string(data), "stability", "stable")

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would restore %q (stability → stable)\n", name)
				return nil
			}

			if err := os.WriteFile(skillMDPath, []byte(updated), 0o644); err != nil {
				return fmt.Errorf("write SKILL.md: %w", err)
			}

			if addToTarget != "" {
				t, ok := cfg.Targets[addToTarget]
				if !ok {
					return fmt.Errorf("target %q not found in bakefile", addToTarget)
				}
				for _, s := range t.Skills {
					if s == name {
						goto skipAdd
					}
				}
				t.Skills = append(t.Skills, name)
				cfg.Targets[addToTarget] = t
				if err := cliDumpBakeFile(cfg, filepath.Join(abs, "agentic.bake.yaml")); err != nil {
					return fmt.Errorf("update bakefile: %w", err)
				}
			skipAdd:
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Restored %q — stability set to stable\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&addToTarget, "add-to-target", "", "Bakefile target to re-add the skill to")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without modifying files")
	return cmd
}

// setFrontmatterField sets a single top-level key in SKILL.md frontmatter.
// Returns the updated content and whether a change was made.
func setFrontmatterField(content, key, value string) (string, bool) {
	if !strings.HasPrefix(content, "---\n") {
		return content, false
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return content, false
	}
	fmRaw := content[4 : 4+end]
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return content, false
	}
	if fmt.Sprintf("%v", fm[key]) == value {
		return content, false
	}
	fm[key] = value
	newFM, err := yaml.Marshal(fm)
	if err != nil {
		return content, false
	}
	body := content[4+end+4:]
	return "---\n" + string(newFM) + "---" + body, true
}

// filterSlice returns a copy of s with all occurrences of exclude removed.
func filterSlice(s []string, exclude string) []string {
	out := s[:0:0]
	for _, v := range s {
		if v != exclude {
			out = append(out, v)
		}
	}
	return out
}

// appendChangelogEntry appends a dated entry to the skill's CHANGELOG.md.
func appendChangelogEntry(skillDir, message string) error {
	clPath := filepath.Join(skillDir, "CHANGELOG.md")
	today := time.Now().UTC().Format("2006-01-02")
	entry := fmt.Sprintf("\n## %s\n\n- %s\n", today, message)

	var existing string
	if data, err := os.ReadFile(clPath); err == nil {
		existing = string(data)
	} else {
		existing = "# Changelog\n"
	}
	return os.WriteFile(clPath, []byte(existing+entry), 0o644)
}
