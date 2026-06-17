package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newFlowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flow",
		Short: "Manage multi-step skill flows",
	}
	cmd.AddCommand(newFlowValidateCommand())
	cmd.AddCommand(newFlowListCommand())
	cmd.AddCommand(newFlowRunCommand())
	cmd.AddCommand(newFlowLintCommand())
	return cmd
}

type flowManifest struct {
	Name    string     `yaml:"name"`
	Version string     `yaml:"version"`
	Steps   []flowStep `yaml:"steps"`
}

type flowStep struct {
	Name  string `yaml:"name"`
	Skill string `yaml:"skill"`
}

func newFlowValidateCommand() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "validate <name>",
		Short: "Validate a flow.yaml, checking all referenced skills exist",
		Long: `Reads <name>/flow.yaml from the skill sources directory and verifies that
every step's skill: field resolves to an existing skill directory.

Exits non-zero if any referenced skill is missing.`,
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
			flowDir := filepath.Join(abs, sourceBase, name)
			flowFile := filepath.Join(flowDir, "flow.yaml")

			data, err := os.ReadFile(flowFile)
			if err != nil {
				return fmt.Errorf("read %s: %w", flowFile, err)
			}

			var fm flowManifest
			if err := yaml.Unmarshal(data, &fm); err != nil {
				return fmt.Errorf("parse flow.yaml: %w", err)
			}

			if len(fm.Steps) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "flow %q has no steps defined.\n", name)
				return nil
			}

			var missing []string
			for _, step := range fm.Steps {
				if step.Skill == "" {
					continue
				}
				skillDir := filepath.Join(abs, sourceBase, step.Skill)
				if _, err := os.Stat(skillDir); os.IsNotExist(err) {
					missing = append(missing, fmt.Sprintf("  step %q: skill %q not found in %s", step.Name, step.Skill, sourceBase))
				}
			}

			if len(missing) > 0 {
				for _, m := range missing {
					fmt.Fprintln(cmd.OutOrStdout(), m)
				}
				return fmt.Errorf("flow %q has %d missing skill(s)", name, len(missing))
			}

			fmt.Fprintf(cmd.OutOrStdout(), "flow %q is valid (%d step(s) OK)\n", name, len(fm.Steps))
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	return cmd
}
