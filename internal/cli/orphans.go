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

func newOrphansCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "orphans",
		Short: "List skill directories not referenced by any target or flow",
		Long: `Resolves all targets (through inherits:) and collects every referenced skill
and flow step, then walks skill_sources.output_dir and reports directories that
are not referenced anywhere. Safe to delete.`,
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

			// Collect all referenced skills from every target.
			referenced := map[string]struct{}{}
			for targetName := range cfg.Targets {
				tc, err := cliResolveTarget(cfg, targetName)
				if err != nil {
					continue
				}
				for _, s := range tc.Skills {
					referenced[s] = struct{}{}
				}
				for _, f := range tc.Flows {
					if validateSkillArg(f) != nil {
						continue
					}
					// Count the flow dir itself as referenced.
					referenced[f] = struct{}{}
					// Also collect skills referenced inside flow steps.
					flowFile := filepath.Join(abs, sourceBase, f, "flow.yaml")
					if data, err := os.ReadFile(flowFile); err == nil {
						var fm struct {
							Steps []struct {
								Skill string `yaml:"skill"`
							} `yaml:"steps"`
						}
						if yaml.Unmarshal(data, &fm) == nil {
							for _, step := range fm.Steps {
								if step.Skill != "" {
									referenced[step.Skill] = struct{}{}
								}
							}
						}
					}
				}
			}

			entries, err := os.ReadDir(sourceDir)
			if err != nil {
				return fmt.Errorf("read skills dir: %w", err)
			}

			var orphans []string
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				if _, ok := referenced[e.Name()]; !ok {
					orphans = append(orphans, e.Name())
				}
			}
			sort.Strings(orphans)

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"orphans": orphans,
					"count":   len(orphans),
				})
			}

			if len(orphans) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No orphaned skill directories found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%d orphaned skill directory/directories:\n\n", len(orphans))
			for _, o := range orphans {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s/%s\n", sourceBase, o)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
