package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newScaffoldFlowCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "flow <name>",
		Short: "Scaffold a multi-step skill flow directory with a flow.yaml manifest",
		Long: `Creates a <name>/ directory under skill_sources.output_dir containing:
  flow.yaml  — flow manifest with steps scaffold
  README.md  — usage guide placeholder

Each step should be an existing skill referenced by name.`,
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

			flowYAML := buildFlowYAML(name)
			readmeMD := buildFlowREADME(name)

			files := map[string]string{
				filepath.Join(flowDir, "flow.yaml"): flowYAML,
				filepath.Join(flowDir, "README.md"): readmeMD,
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would scaffold flow %q in %s\n", name, flowDir)
				for p := range files {
					rel, _ := filepath.Rel(abs, p)
					fmt.Fprintf(cmd.OutOrStdout(), "  + %s\n", rel)
				}
				return nil
			}

			if err := os.MkdirAll(flowDir, 0o755); err != nil {
				return fmt.Errorf("create flow directory: %w", err)
			}

			created := 0
			for p, content := range files {
				if _, err := os.Stat(p); err == nil {
					rel, _ := filepath.Rel(abs, p)
					fmt.Fprintf(cmd.OutOrStdout(), "  skip  %s (already exists)\n", rel)
					continue
				}
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", p, err)
				}
				rel, _ := filepath.Rel(abs, p)
				fmt.Fprintf(cmd.OutOrStdout(), "  +  %s\n", rel)
				created++
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\nScaffolded flow %q (%d file(s) created).\n", name, created)
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be created without writing")
	return cmd
}

func buildFlowYAML(name string) string {
	return strings.Join([]string{
		"# flow.yaml — multi-step skill flow",
		"name: " + name,
		"version: 0.1.0",
		"description: \"\"",
		"steps:",
		"  - name: step-1",
		"    skill: \"\"",
		"    input: {}",
		"  - name: step-2",
		"    skill: \"\"",
		"    input: {}",
	}, "\n") + "\n"
}

func buildFlowREADME(name string) string {
	return strings.Join([]string{
		"# " + name,
		"",
		"Multi-step skill flow.",
		"",
		"## Steps",
		"",
		"| Step | Skill | Description |",
		"|------|-------|-------------|",
		"| step-1 | | |",
		"| step-2 | | |",
		"",
		"## Usage",
		"",
		"```yaml",
		"# reference this flow from a bakefile target",
		"targets:",
		"  default:",
		"    flows:",
		"      - " + name,
		"```",
	}, "\n") + "\n"
}
