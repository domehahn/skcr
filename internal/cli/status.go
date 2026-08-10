package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show skill scaffold status across all platform directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			resolved, err := resolveDefaultTarget(cfg)
			if err != nil {
				return err
			}

			// Collect unique platform dirs in a stable order (canonical source first).
			agentsBase := skillSourceOutputDir(cfg.SkillSources)
			dirSeen := map[string]struct{}{}
			dirs := []string{agentsBase}
			dirSeen[agentsBase] = struct{}{}
			for _, p := range resolved.Platforms {
				d := canonicalPlatformSkillBaseDir(p)
				if _, dup := dirSeen[d]; !dup {
					dirSeen[d] = struct{}{}
					dirs = append(dirs, d)
				}
			}

			// Collect unique skills from ALL targets (not just resolved) so nothing is hidden.
			skillSeen := map[string]struct{}{}
			var skills []string
			for _, t := range cfg.Targets {
				for _, s := range t.Skills {
					if _, dup := skillSeen[s]; !dup {
						skillSeen[s] = struct{}{}
						skills = append(skills, s)
					}
				}
			}
			sort.Strings(skills)

			if len(skills) == 0 {
				fmt.Println("No skills defined in any target.")
				return nil
			}

			// Print header.
			const nameWidth = 32
			const verWidth = 10
			const colWidth = 16
			header := fmt.Sprintf("%-*s  %-*s", nameWidth, "Skill", verWidth, "Version")
			for _, d := range dirs {
				header += fmt.Sprintf("  %-*s", colWidth, shortDirLabel(d))
			}
			fmt.Println(header)
			fmt.Println(strings.Repeat("─", len(header)))

			type skillStatus struct {
				Name    string            `json:"name"`
				Version string            `json:"version"`
				Dirs    map[string]string `json:"dirs"`
			}

			var rows []skillStatus
			inSync, outOfSync, missing, invalid := 0, 0, 0, 0
			for _, s := range skills {
				canonicalDir := filepath.Join(absTarget, agentsBase, s)
				canonicalPath := filepath.Join(canonicalDir, "SKILL.md")
				canonicalData, _ := os.ReadFile(canonicalPath)
				descriptorErrors := skillmeta.ValidateDirectory(canonicalDir)

				ver := "-"
				if fm, ok := parseFrontmatterDates(string(canonicalData)); ok && fm.Version != "" {
					ver = fm.Version
				}

				dirStatus := map[string]string{}
				row := fmt.Sprintf("%-*s  %-*s", nameWidth, s, verWidth, ver)
				for _, d := range dirs {
					skillDir := filepath.Join(absTarget, d, s)
					_, statErr := os.Stat(filepath.Join(skillDir, "SKILL.md"))
					var cell, cellJSON string
					switch {
					case os.IsNotExist(statErr):
						cell, cellJSON = "✗", "missing"
						missing++
					case d == agentsBase && len(descriptorErrors) > 0:
						cell, cellJSON = "!", "invalid"
						invalid++
					case d == agentsBase || skillArtifactsEqual(canonicalDir, skillDir):
						cell, cellJSON = "✓", "ok"
						inSync++
					default:
						cell, cellJSON = "~", "differs"
						outOfSync++
					}
					row += fmt.Sprintf("  %-*s", colWidth, cell)
					dirStatus[d] = cellJSON
				}
				rows = append(rows, skillStatus{Name: s, Version: ver, Dirs: dirStatus})
				if !jsonOut {
					fmt.Println(row)
				}
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rows)
			}

			fmt.Println()
			legend := []string{fmt.Sprintf("%d ✓ in sync", inSync)}
			if outOfSync > 0 {
				legend = append(legend, fmt.Sprintf("%d ~ differs (run skcr sync)", outOfSync))
			}
			if invalid > 0 {
				legend = append(legend, fmt.Sprintf("%d ! invalid contract (run skcr validate)", invalid))
			}
			if missing > 0 {
				legend = append(legend, fmt.Sprintf("%d ✗ missing (run skcr bake --write)", missing))
			}
			fmt.Println(strings.Join(legend, "  ·  "))
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func skillArtifactsEqual(canonicalDir, destinationDir string) bool {
	artifacts, err := canonicalSkillArtifacts(canonicalDir)
	if err != nil {
		return false
	}
	for _, rel := range artifacts {
		left, err := os.ReadFile(filepath.Join(canonicalDir, rel))
		if err != nil {
			return false
		}
		right, err := os.ReadFile(filepath.Join(destinationDir, rel))
		if err != nil || string(left) != string(right) {
			return false
		}
	}
	return true
}

func shortDirLabel(dir string) string {
	parts := strings.SplitN(dir, "/", 2)
	label := strings.TrimPrefix(parts[0], ".")
	if label == "agents" {
		return "agents" // .agents/skills → agents
	}
	return label
}
