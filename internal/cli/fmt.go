package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// fmtKeyOrder defines canonical SKILL.md frontmatter key ordering.
var fmtKeyOrder = []string{
	"name", "description", "version", "since", "last_modified",
	"authors", "stability", "min_platform_version", "license", "changelog",
}

func newFmtCommand() *cobra.Command {
	var repoPath string
	var dryRun bool
	var skillFilter string

	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "Normalize SKILL.md frontmatter key order and whitespace",
		Long: `Reformats SKILL.md frontmatter in canonical key order, trims whitespace
from string values, and converts flow-style lists to block style.
Like gofmt, but for skill metadata. Idempotent.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			seen := map[string]struct{}{}
			var names []string
			for _, t := range cfg.Targets {
				for _, s := range t.Skills {
					if _, dup := seen[s]; !dup {
						seen[s] = struct{}{}
						names = append(names, s)
					}
				}
			}
			sort.Strings(names)

			if skillFilter != "" {
				names = []string{skillFilter}
			}

			changed, skipped := 0, 0
			for _, name := range names {
				skillMD := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
				data, readErr := os.ReadFile(skillMD)
				if readErr != nil {
					skipped++
					continue
				}
				normalized, didChange, fmtErr := normalizeFrontmatter(string(data))
				if fmtErr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "skip %s: %v\n", name, fmtErr)
					skipped++
					continue
				}
				if !didChange {
					fmt.Fprintf(cmd.OutOrStdout(), "ok      %s\n", name)
					continue
				}
				if dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "would-fmt  %s\n", name)
					changed++
					continue
				}
				if writeErr := os.WriteFile(skillMD, []byte(normalized), 0644); writeErr != nil {
					return fmt.Errorf("write %s: %w", skillMD, writeErr)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "fmt     %s\n", name)
				changed++
			}

			verb := "formatted"
			if dryRun {
				verb = "would format"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d %s, %d skipped (no SKILL.md).\n", changed, verb, skipped)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without writing")
	cmd.Flags().StringVar(&skillFilter, "skill", "", "Format only this skill")
	_ = cmd.RegisterFlagCompletionFunc("skill", completeSkillNames)
	return cmd
}

// normalizeFrontmatter rewrites SKILL.md frontmatter in canonical key order with
// whitespace-trimmed string values. Returns (normalized content, changed, error).
func normalizeFrontmatter(content string) (string, bool, error) {
	if !strings.HasPrefix(content, "---\n") {
		return content, false, nil
	}
	idx := strings.Index(content[4:], "\n---")
	if idx < 0 {
		return content, false, nil
	}
	fmRaw := content[4 : 4+idx]
	body := content[4+idx+4:] // skip the closing "\n---"

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return content, false, err
	}

	// Trim whitespace from all string values.
	for k, v := range fm {
		if s, ok := v.(string); ok {
			fm[k] = strings.TrimSpace(s)
		}
	}

	var sb strings.Builder
	sb.WriteString("---\n")

	written := map[string]bool{}
	for _, key := range fmtKeyOrder {
		v, ok := fm[key]
		if !ok {
			continue
		}
		chunk, err := yaml.Marshal(map[string]any{key: v})
		if err != nil {
			return content, false, err
		}
		sb.WriteString(string(chunk))
		written[key] = true
	}

	// Remaining unknown keys in alphabetical order.
	var extra []string
	for k := range fm {
		if !written[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	for _, key := range extra {
		chunk, err := yaml.Marshal(map[string]any{key: fm[key]})
		if err != nil {
			return content, false, err
		}
		sb.WriteString(string(chunk))
	}

	sb.WriteString("---")
	sb.WriteString(body)

	normalized := sb.String()
	return normalized, normalized != content, nil
}
