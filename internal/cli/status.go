package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var target string

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

			// Collect unique platform dirs in a stable order (agents first).
			const agentsBase = ".agents/skills"
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
			const colWidth = 16
			header := fmt.Sprintf("%-*s", nameWidth, "Skill")
			for _, d := range dirs {
				header += fmt.Sprintf("  %-*s", colWidth, shortDirLabel(d))
			}
			fmt.Println(header)
			fmt.Println(strings.Repeat("─", len(header)))

			inSync, outOfSync, missing := 0, 0, 0
			for _, s := range skills {
				canonicalPath := filepath.Join(absTarget, agentsBase, s, "SKILL.md")
				canonicalData, _ := os.ReadFile(canonicalPath)

				row := fmt.Sprintf("%-*s", nameWidth, s)
				for _, d := range dirs {
					skillMD := filepath.Join(absTarget, d, s, "SKILL.md")
					data, statErr := os.ReadFile(skillMD)
					var cell string
					switch {
					case os.IsNotExist(statErr):
						cell = "✗"
						missing++
					case d == agentsBase || string(data) == string(canonicalData):
						cell = "✓"
						inSync++
					default:
						cell = "~" // differs from canonical — run skcr sync
						outOfSync++
					}
					row += fmt.Sprintf("  %-*s", colWidth, cell)
				}
				fmt.Println(row)
			}

			fmt.Println()
			legend := []string{fmt.Sprintf("%d ✓ in sync", inSync)}
			if outOfSync > 0 {
				legend = append(legend, fmt.Sprintf("%d ~ differs (run skcr sync)", outOfSync))
			}
			if missing > 0 {
				legend = append(legend, fmt.Sprintf("%d ✗ missing (run skcr bake --write)", missing))
			}
			fmt.Println(strings.Join(legend, "  ·  "))
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	return cmd
}

func shortDirLabel(dir string) string {
	parts := strings.SplitN(dir, "/", 2)
	label := strings.TrimPrefix(parts[0], ".")
	if label == "agents" {
		return "agents" // .agents/skills → agents
	}
	return label
}
