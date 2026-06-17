package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type coverageRow struct {
	Skill    string `json:"skill"`
	Cases    int    `json:"cases"`
	Covered  bool   `json:"covered"`
	HasError string `json:"error,omitempty"`
}

func newCoverageCommand() *cobra.Command {
	var repoPath string
	var bakeTarget string
	var failUnder int
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Report which skills have skill-tests.yaml and how many test cases each has",
		Long: `Scans each skill directory for a skill-tests.yaml manifest and reports:
  - which skills are covered (have tests)
  - how many test cases each covered skill declares
  - overall coverage percentage

Use --fail-under <pct> to exit non-zero when coverage falls below a threshold (CI).`,
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

			seen := map[string]struct{}{}
			var skills []string
			for tName, t := range cfg.Targets {
				if bakeTarget != "" && tName != bakeTarget {
					continue
				}
				for _, s := range t.Skills {
					if _, dup := seen[s]; !dup {
						seen[s] = struct{}{}
						skills = append(skills, s)
					}
				}
			}
			sort.Strings(skills)

			var rows []coverageRow
			for _, skillName := range skills {
				manifestPath := filepath.Join(abs, sourceBase, skillName, "skill-tests.yaml")
				data, readErr := os.ReadFile(manifestPath)
				if readErr != nil {
					rows = append(rows, coverageRow{Skill: skillName, Covered: false})
					continue
				}
				var manifest struct {
					Cases []any `yaml:"cases"`
				}
				if err := yaml.Unmarshal(data, &manifest); err != nil {
					rows = append(rows, coverageRow{Skill: skillName, Covered: false, HasError: err.Error()})
					continue
				}
				rows = append(rows, coverageRow{Skill: skillName, Cases: len(manifest.Cases), Covered: len(manifest.Cases) > 0})
			}

			covered := 0
			totalCases := 0
			for _, r := range rows {
				if r.Covered {
					covered++
					totalCases += r.Cases
				}
			}

			pct := 0
			if len(rows) > 0 {
				pct = covered * 100 / len(rows)
			}

			if jsonOut {
				out := map[string]any{
					"skills":        rows,
					"covered":       covered,
					"total":         len(rows),
					"percent":       pct,
					"total_cases":   totalCases,
				}
				data, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Test coverage: %d/%d skills (%d%%)\n\n", covered, len(rows), pct)
				for _, r := range rows {
					if r.Covered {
						fmt.Fprintf(cmd.OutOrStdout(), "  ✓  %-28s  %d case(s)\n", r.Skill, r.Cases)
					} else if r.HasError != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "  !  %-28s  invalid skill-tests.yaml\n", r.Skill)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  -  %-28s  no tests\n", r.Skill)
					}
				}
			}

			if failUnder > 0 && pct < failUnder {
				return fmt.Errorf("coverage %d%% is below threshold %d%%", pct, failUnder)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringVar(&bakeTarget, "bake-target", "", "Restrict to a single bakefile target")
	cmd.Flags().IntVar(&failUnder, "fail-under", 0, "Exit non-zero if coverage < this percentage")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
