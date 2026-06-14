package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/domehahn/skcr/internal/skillversion"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type reportFinding struct {
	Category string `json:"category"`
	Skill    string `json:"skill,omitempty"`
	Level    string `json:"level"`
	Message  string `json:"message"`
}

func newReportCommand() *cobra.Command {
	var repoPath string
	var jsonOut bool
	var failOnWarn bool

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Full project health report combining audit, validate, and structure checks",
		Long: `Runs audit (metadata quality), version validation, and structural checks
in a single pass and produces a consolidated report. Suitable for CI pipelines.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			var findings []reportFinding
			add := func(cat, skill, level, msg string) {
				findings = append(findings, reportFinding{cat, skill, level, msg})
			}

			// ── Bakefile ─────────────────────────────────────────────────────
			bakePath := filepath.Join(absTarget, "agentic.bake.yaml")
			cfg, loadErr := cliLoadBakeFile(bakePath)
			if loadErr != nil {
				add("bakefile", "", "error", fmt.Sprintf("cannot parse agentic.bake.yaml: %v", loadErr))
				return emitReport(cmd, findings, jsonOut, failOnWarn)
			}
			add("bakefile", "", "ok", "agentic.bake.yaml is valid")

			sourceBase := skillSourceOutputDir(cfg.SkillSources)

			// Collect unique skills.
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

			now := time.Now()

			for _, name := range names {
				skillDir := filepath.Join(absTarget, sourceBase, name)
				skillMDPath := filepath.Join(skillDir, "SKILL.md")
				data, readErr := os.ReadFile(skillMDPath)
				if readErr != nil {
					add("scaffold", name, "warn", "SKILL.md not found — run: skcr bake --write")
					continue
				}

				// Parse frontmatter.
				var fm map[string]any
				content := string(data)
				if strings.HasPrefix(content, "---\n") {
					end := strings.Index(content[4:], "\n---")
					if end >= 0 {
						_ = yaml.Unmarshal([]byte(content[4:4+end]), &fm)
					}
				}

				// Metadata completeness (audit-equivalent).
				checkAny := func(primary string, aliases ...string) {
					all := append([]string{primary}, aliases...)
					for _, key := range all {
						if v, ok := fm[key]; ok && v != nil && v != "" {
							return
						}
					}
					add("metadata", name, "warn", fmt.Sprintf("missing field: %s", primary))
				}
				checkAny("description")
				checkAny("owner", "authors")
				checkAny("version")

				// Staleness (> 1 year).
				if lm, ok := fm["last_modified"].(string); ok && lm != "" {
					t, parseErr := time.Parse("2006-01-02", lm)
					if parseErr == nil && now.Sub(t) > 365*24*time.Hour {
						add("metadata", name, "warn", fmt.Sprintf("last_modified %s is older than 365 days", lm))
					}
				}

				// Version consistency via skillversion.Check.
				infos, checkErr := skillversion.Check(skillDir)
				if checkErr == nil {
					for _, info := range infos {
						for _, e := range info.Errors {
							add("version", name, "error", e)
						}
						for _, w := range info.Warnings {
							add("version", name, "warn", w)
						}
					}
				}

				// VERSION file drift vs SKILL.md.
				if ver, ok := fm["version"].(string); ok && ver != "" {
					if vfBytes, vfErr := os.ReadFile(filepath.Join(skillDir, "VERSION")); vfErr == nil {
						if vf := strings.TrimSpace(string(vfBytes)); vf != ver {
							add("version", name, "error", fmt.Sprintf("VERSION file (%s) differs from SKILL.md version (%s) — run: skcr validate --fix", vf, ver))
						}
					}
				}
			}

			return emitReport(cmd, findings, jsonOut, failOnWarn)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Exit non-zero on warnings as well as errors")
	return cmd
}

func emitReport(cmd *cobra.Command, findings []reportFinding, jsonOut, failOnWarn bool) error {
	if jsonOut {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(findings)
	}

	icons := map[string]string{"ok": "✓", "error": "✗", "warn": "!"}
	errors, warns := 0, 0
	prevCat := ""
	for _, f := range findings {
		if f.Level == "ok" {
			continue
		}
		if f.Category != prevCat {
			fmt.Fprintf(cmd.OutOrStdout(), "\n[%s]\n", f.Category)
			prevCat = f.Category
		}
		skill := f.Skill
		if skill == "" {
			skill = "-"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-40s  %s\n", icons[f.Level], skill, f.Message)
		switch f.Level {
		case "error":
			errors++
		case "warn":
			warns++
		}
	}

	// Count ok items.
	oks := 0
	for _, f := range findings {
		if f.Level == "ok" {
			oks++
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nReport: %d ok, %d error(s), %d warning(s).\n", oks, errors, warns)
	if errors > 0 || (failOnWarn && warns > 0) {
		return fmt.Errorf("report found issues")
	}
	return nil
}
