package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type flowSummary struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Steps   int    `json:"steps"`
}

func newFlowListCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all skill flows in the skill sources directory",
		Long: `Scans skill_sources.output_dir for directories containing a flow.yaml and
prints a summary of each flow's name, version, and step count.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			sourceDir := filepath.Join(abs, sourceBase)

			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				return fmt.Errorf("read skill sources dir: %w", err)
			}

			var flows []flowSummary
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				flowFile := filepath.Join(sourceDir, entry.Name(), "flow.yaml")
				data, err := os.ReadFile(flowFile)
				if err != nil {
					continue
				}
				var fm flowManifest
				if err := yaml.Unmarshal(data, &fm); err != nil {
					continue
				}
				flows = append(flows, flowSummary{
					Name:    entry.Name(),
					Version: fm.Version,
					Steps:   len(fm.Steps),
				})
			}

			if len(flows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No flows found.")
				return nil
			}

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(flows)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "  %-28s  %-10s  steps\n", "flow", "version")
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", "──────────────────────────────────────────────")
			for _, f := range flows {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-28s  %-10s  %d\n", f.Name, f.Version, f.Steps)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d flow(s)\n", len(flows))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
