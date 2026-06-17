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

type depsResult struct {
	Skill   string   `json:"skill"`
	Targets []string `json:"targets"`
	Flows   []string `json:"flows,omitempty"`
}

func newDepsCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "deps <skill>",
		Short: "Show which bakefile targets and flows depend on a skill",
		Long: `Reverse dependency lookup: lists every bakefile target whose skills: list
includes <skill>, and every flow.yaml whose steps reference <skill>.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			// Find targets.
			var targets []string
			for name, tc := range cfg.Targets {
				for _, s := range tc.Skills {
					if s == skillName {
						targets = append(targets, name)
						break
					}
				}
			}
			sort.Strings(targets)

			// Find flows that reference this skill.
			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			sourceDir := filepath.Join(abs, sourceBase)

			var flows []string
			entries, _ := os.ReadDir(sourceDir)
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				data, err := os.ReadFile(filepath.Join(sourceDir, entry.Name(), "flow.yaml"))
				if err != nil {
					continue
				}
				var fm flowManifest
				if err := yaml.Unmarshal(data, &fm); err != nil {
					continue
				}
				for _, step := range fm.Steps {
					if step.Skill == skillName {
						flows = append(flows, entry.Name())
						break
					}
				}
			}
			sort.Strings(flows)

			result := depsResult{Skill: skillName, Targets: targets, Flows: flows}

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
			}

			if len(targets) == 0 && len(flows) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No targets or flows depend on %q.\n", skillName)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Skill: %s\n", skillName)
			if len(targets) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nTargets:")
				for _, t := range targets {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", t)
				}
			}
			if len(flows) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nFlows:")
				for _, f := range flows {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", f)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
