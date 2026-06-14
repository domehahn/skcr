package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newAuditCommand() *cobra.Command {
	var repoPath string
	var maxAgeDays int
	var jsonOut bool
	var failOnWarn bool

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Check skill metadata quality across all scaffolded skills",
		Long: `Audits skill metadata completeness and freshness. Unlike 'doctor' (which
checks structural health), 'audit' enforces team metadata standards:
required fields, version consistency, and staleness thresholds.`,
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

			type finding struct {
				Skill   string `json:"skill"`
				Level   string `json:"level"`
				Message string `json:"message"`
			}
			var findings []finding
			add := func(skill, level, msg string) {
				findings = append(findings, finding{skill, level, msg})
			}

			now := time.Now()
			staleThreshold := time.Duration(maxAgeDays) * 24 * time.Hour

			for _, name := range names {
				skillDir := filepath.Join(absTarget, sourceBase, name)
				skillMDPath := filepath.Join(skillDir, "SKILL.md")
				data, readErr := os.ReadFile(skillMDPath)
				if readErr != nil {
					add(name, "warn", "SKILL.md not found — run: skcr bake --write")
					continue
				}

				var fm map[string]any
				if strings.HasPrefix(string(data), "---\n") {
					end := strings.Index(string(data)[4:], "\n---")
					if end >= 0 {
						_ = yaml.Unmarshal([]byte(string(data)[4:4+end]), &fm)
					}
				}

				// requireAny warns if none of the candidate keys has a non-empty value.
				// Supports field aliasing: scaffold writes "authors" while the standard uses "owner".
				requireAny := func(primary string, aliases ...string) {
					all := append([]string{primary}, aliases...)
					for _, key := range all {
						v, ok := fm[key]
						if ok && v != nil && v != "" {
							return
						}
					}
					add(name, "warn", fmt.Sprintf("missing field: %s", primary))
				}
				requireAny("description")
				requireAny("owner", "authors")
				requireAny("license")
				requireAny("compatible_with", "min_platform_version")
				requireAny("version")

				if lm, ok := fm["last_modified"].(string); ok && lm != "" && maxAgeDays > 0 {
					t, parseErr := time.Parse("2006-01-02", lm)
					if parseErr == nil && now.Sub(t) > staleThreshold {
						add(name, "warn", fmt.Sprintf("last_modified %s is older than %d days", lm, maxAgeDays))
					}
				}

				// Version drift: VERSION file vs SKILL.md.
				if ver, ok := fm["version"].(string); ok && ver != "" {
					versionFile, vfErr := os.ReadFile(filepath.Join(skillDir, "VERSION"))
					if vfErr == nil {
						if vf := strings.TrimSpace(string(versionFile)); vf != ver {
							add(name, "error", fmt.Sprintf("VERSION file (%s) differs from SKILL.md version (%s) — run: skcr validate --fix", vf, ver))
						}
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
				fmt.Fprintf(cmd.OutOrStdout(), "Audit passed: %d skill(s) checked, no issues found.\n", len(names))
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d skill(s) checked: %d error(s), %d warning(s).\n", len(names), errors, warns)
			if errors > 0 || (failOnWarn && warns > 0) {
				return fmt.Errorf("audit failed")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().IntVar(&maxAgeDays, "max-age", 365, "Warn if last_modified is older than this many days (0 = disable)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Exit non-zero on warnings as well as errors")
	return cmd
}
