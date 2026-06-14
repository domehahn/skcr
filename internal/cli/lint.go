package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// lintKnownStabilities are the accepted stability values.
var lintKnownStabilities = map[string]bool{
	"stable": true, "experimental": true, "deprecated": true, "beta": true,
}

// lintPlaceholders are substrings that indicate placeholder content.
var lintPlaceholders = []string{
	"TODO", "FIXME", "<your-skill>", "<skill-name>", "Example Skill",
	"My Skill", "your skill", "placeholder",
}

func newLintCommand() *cobra.Command {
	var repoPath string
	var jsonOut bool
	var failOnWarn bool
	var minDescLen int

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Check skill content quality: placeholders, short descriptions, unknown stability values",
		Long: `Content-level linting for SKILL.md files. Complements 'audit' (structural
completeness) and 'fmt' (formatting) by checking the quality of the actual
written content.`,
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

			type lintFinding struct {
				Skill   string `json:"skill"`
				Level   string `json:"level"`
				Message string `json:"message"`
			}
			var findings []lintFinding
			add := func(skill, level, msg string) {
				findings = append(findings, lintFinding{skill, level, msg})
			}

			for _, name := range names {
				skillMD := filepath.Join(absTarget, sourceBase, name, "SKILL.md")
				data, readErr := os.ReadFile(skillMD)
				if readErr != nil {
					add(name, "warn", "SKILL.md not found — run: skcr bake --write")
					continue
				}

				content := string(data)

				// Parse frontmatter.
				var fm map[string]any
				if strings.HasPrefix(content, "---\n") {
					end := strings.Index(content[4:], "\n---")
					if end >= 0 {
						_ = yaml.Unmarshal([]byte(content[4:4+end]), &fm)
					}
				}

				desc, _ := fm["description"].(string)
				stab, _ := fm["stability"].(string)

				// Description checks.
				if desc != "" {
					if len(desc) < minDescLen {
						add(name, "warn", fmt.Sprintf("description too short (%d chars, minimum %d)", len(desc), minDescLen))
					}
					if strings.EqualFold(strings.TrimSpace(desc), name) {
						add(name, "warn", "description is identical to the skill name — add meaningful context")
					}
					for _, ph := range lintPlaceholders {
						if strings.Contains(strings.ToLower(desc), strings.ToLower(ph)) {
							add(name, "error", fmt.Sprintf("description contains placeholder text: %q", ph))
							break
						}
					}
				}

				// Stability.
				if stab != "" && !lintKnownStabilities[strings.ToLower(stab)] {
					add(name, "warn", fmt.Sprintf("unknown stability value %q (expected: stable, experimental, beta, deprecated)", stab))
				}

				// Body content length (after frontmatter).
				body := content
				if strings.HasPrefix(content, "---\n") {
					end := strings.Index(content[4:], "\n---")
					if end >= 0 {
						body = strings.TrimSpace(content[4+end+4:])
					}
				}
				if len(body) < 50 {
					add(name, "warn", "SKILL.md body is very short (< 50 chars) — consider adding usage instructions")
				}

				// Placeholder text anywhere in the body.
				bodyLower := strings.ToLower(body)
				for _, ph := range lintPlaceholders {
					if strings.Contains(bodyLower, strings.ToLower(ph)) {
						add(name, "error", fmt.Sprintf("body contains placeholder text: %q", ph))
						break
					}
				}
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(findings)
			}

			icons := map[string]string{"error": "✗", "warn": "!"}
			errors, warns := 0, 0
			for _, f := range findings {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-40s  %s\n", icons[f.Level], f.Skill, f.Message)
				switch f.Level {
				case "error":
					errors++
				case "warn":
					warns++
				}
			}
			if len(findings) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Lint passed: %d skill(s) checked, no issues found.\n", len(names))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d skill(s) checked: %d error(s), %d warning(s).\n", len(names), errors, warns)
			if errors > 0 || (failOnWarn && warns > 0) {
				return fmt.Errorf("lint failed")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Exit non-zero on warnings as well as errors")
	cmd.Flags().IntVar(&minDescLen, "min-desc-len", 20, "Minimum description length in characters")
	return cmd
}
