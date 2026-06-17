package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newFlowRunCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Print a step-by-step execution plan for a flow (dry run)",
		Long: `Reads <name>/flow.yaml and prints each step in execution order showing
the skill name and any declared inputs. No skills are actually invoked.

Useful for reviewing the data flow of a multi-step flow before running it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			flowFile := filepath.Join(abs, sourceBase, name, "flow.yaml")

			data, err := os.ReadFile(flowFile)
			if err != nil {
				return fmt.Errorf("read %s: %w", flowFile, err)
			}

			var fm flowManifest
			if err := yaml.Unmarshal(data, &fm); err != nil {
				return fmt.Errorf("parse flow.yaml: %w", err)
			}

			if len(fm.Steps) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Flow %q has no steps defined.\n", name)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Flow: %s", name)
			if fm.Version != "" {
				fmt.Fprintf(cmd.OutOrStdout(), " (%s)", fm.Version)
			}
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout())

			for i, step := range fm.Steps {
				skillLabel := step.Skill
				if skillLabel == "" {
					skillLabel = "(unset)"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  Step %d: %s\n", i+1, step.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "    skill: %s\n", skillLabel)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n%d step(s) — dry run only, no skills were invoked.\n", len(fm.Steps))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	return cmd
}
