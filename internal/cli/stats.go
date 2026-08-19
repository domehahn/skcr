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
)

func newStatsCommand() *cobra.Command {
	var repoPath string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show project statistics: skill count, coverage, platforms, and age",
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

			// Unique skills and platforms.
			skillSeen := map[string]struct{}{}
			platformSeen := map[string]struct{}{}
			for _, t := range cfg.Targets {
				for _, s := range t.Skills {
					skillSeen[s] = struct{}{}
				}
				for _, p := range t.Platforms {
					platformSeen[p] = struct{}{}
				}
			}

			totalSkills := len(skillSeen)
			totalTargets := len(cfg.Targets)

			var platforms []string
			for p := range platformSeen {
				platforms = append(platforms, p)
			}
			sort.Strings(platforms)

			// Per-skill analysis.
			scaffolded, drifted := 0, 0
			var ages []float64
			now := time.Now()

			for name := range skillSeen {
				skillDir := filepath.Join(absTarget, sourceBase, name)
				skillMD := filepath.Join(skillDir, "SKILL.md")
				data, readErr := os.ReadFile(skillMD)
				if readErr != nil {
					continue
				}
				scaffolded++

				fm, ok := parseFrontmatterDates(string(data))
				if !ok {
					continue
				}

				// Version drift check.
				if fm.Version != "" {
					if vfBytes, vfErr := os.ReadFile(filepath.Join(skillDir, "VERSION")); vfErr == nil {
						if strings.TrimSpace(string(vfBytes)) != fm.Version {
							drifted++
						}
					}
				}

				// Age from since date.
				if fm.Since != "" {
					t, parseErr := time.Parse("2006-01-02", fm.Since)
					if parseErr == nil {
						ages = append(ages, now.Sub(t).Hours()/24)
					}
				}
			}

			var avgAgeDays float64
			if len(ages) > 0 {
				sum := 0.0
				for _, a := range ages {
					sum += a
				}
				avgAgeDays = sum / float64(len(ages))
			}

			coveragePct := 0
			if totalSkills > 0 {
				coveragePct = scaffolded * 100 / totalSkills
			}

			type statsOutput struct {
				TotalSkills  int      `json:"total_skills"`
				TotalTargets int      `json:"total_targets"`
				Scaffolded   int      `json:"scaffolded"`
				CoveragePct  int      `json:"coverage_pct"`
				VersionDrift int      `json:"version_drift"`
				Platforms    []string `json:"platforms"`
				AvgAgedays   float64  `json:"avg_age_days,omitempty"`
			}

			out := statsOutput{
				TotalSkills:  totalSkills,
				TotalTargets: totalTargets,
				Scaffolded:   scaffolded,
				CoveragePct:  coveragePct,
				VersionDrift: drifted,
				Platforms:    platforms,
				AvgAgedays:   avgAgeDays,
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Skills\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Total           %d\n", out.TotalSkills)
			fmt.Fprintf(cmd.OutOrStdout(), "  Scaffolded      %d  (%d%%)\n", out.Scaffolded, out.CoveragePct)
			fmt.Fprintf(cmd.OutOrStdout(), "  Version drift   %d\n", out.VersionDrift)
			if out.AvgAgedays > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Avg age         %.0f days\n", out.AvgAgedays)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nTargets\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Total           %d\n", out.TotalTargets)
			fmt.Fprintf(cmd.OutOrStdout(), "\nPlatforms\n")
			if len(out.Platforms) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  (none)\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", strings.Join(out.Platforms, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}
