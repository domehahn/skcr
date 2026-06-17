package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// lintFlowStep extends flowStep with optional lint-relevant fields.
type lintFlowStep struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`
	Skill     string   `yaml:"skill"`
	DependsOn []string `yaml:"depends_on"`
}

type lintFlowManifest struct {
	Name    string         `yaml:"name"`
	Version string         `yaml:"version"`
	Steps   []lintFlowStep `yaml:"steps"`
}

type flowLintIssue struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func newFlowLintCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "lint <name>",
		Short: "Static lint a flow.yaml for structural problems",
		Long: `Checks a flow.yaml for:
  - Duplicate step IDs or names
  - Steps with no skill: field
  - Undefined depends_on references
  - Dependency cycles`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := validateSkillArg(name); err != nil {
				return err
			}

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

			var fm lintFlowManifest
			if err := yaml.Unmarshal(data, &fm); err != nil {
				return fmt.Errorf("parse flow.yaml: %w", err)
			}

			issues := lintFlow(fm, abs, sourceBase)

			if jsonOut {
				type jsonResult struct {
					Flow   string          `json:"flow"`
					Issues []flowLintIssue `json:"issues"`
					OK     bool            `json:"ok"`
				}
				result := jsonResult{Flow: name, OK: len(issues) == 0, Issues: issues}
				if result.Issues == nil {
					result.Issues = []flowLintIssue{}
				}
				if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
					return err
				}
				if len(issues) > 0 {
					return fmt.Errorf("flow has %d issue(s)", len(issues))
				}
				return nil
			}

			if len(issues) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "flow %q: OK (%d steps, no issues)\n", name, len(fm.Steps))
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "flow %q: %d issue(s)\n\n", name, len(issues))
			for _, iss := range issues {
				fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", iss.Severity, iss.Message)
			}
			return fmt.Errorf("flow has %d issue(s)", len(issues))
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func lintFlow(fm lintFlowManifest, repoRoot, sourceBase string) []flowLintIssue {
	var issues []flowLintIssue

	seenIDs := map[string]int{}
	allIDs := map[string]struct{}{}

	for i, step := range fm.Steps {
		id := stepID(step, i)
		allIDs[id] = struct{}{}
		// Register both id: and name: so DependsOn can reference either.
		if step.ID != "" && step.Name != "" {
			allIDs[step.Name] = struct{}{}
		}

		seenIDs[id]++
		if seenIDs[id] == 2 {
			issues = append(issues, flowLintIssue{"error", fmt.Sprintf("duplicate step ID/name %q", id)})
		}

		if step.Skill == "" {
			issues = append(issues, flowLintIssue{"error", fmt.Sprintf("step %q has no skill: field", id)})
		} else {
			skillPath := filepath.Join(repoRoot, sourceBase, step.Skill)
			if _, err := os.Stat(skillPath); err != nil {
				issues = append(issues, flowLintIssue{"error", fmt.Sprintf("step %q references unknown skill %q", id, step.Skill)})
			}
		}
	}

	// Check depends_on references after all IDs are collected.
	for i, step := range fm.Steps {
		id := stepID(step, i)
		for _, dep := range step.DependsOn {
			if _, ok := allIDs[dep]; !ok {
				issues = append(issues, flowLintIssue{"error", fmt.Sprintf("step %q depends_on undefined step %q", id, dep)})
			}
		}
	}

	if hasCycle(fm.Steps) {
		issues = append(issues, flowLintIssue{"error", "dependency cycle detected among steps"})
	}

	return issues
}

func stepID(step lintFlowStep, idx int) string {
	if step.ID != "" {
		return step.ID
	}
	if step.Name != "" {
		return step.Name
	}
	return fmt.Sprintf("step[%d]", idx)
}

// hasCycle performs DFS cycle detection on the depends_on graph.
func hasCycle(steps []lintFlowStep) bool {
	// Build normalised adjacency: register both id: and name: so DependsOn
	// references work regardless of which alias the author used.
	adj := map[string][]string{}
	for i, step := range steps {
		id := stepID(step, i)
		adj[id] = step.DependsOn
		if step.ID != "" && step.Name != "" {
			adj[step.Name] = step.DependsOn
		}
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var dfs func(id string) bool
	dfs = func(id string) bool {
		color[id] = gray
		for _, dep := range adj[id] {
			if color[dep] == gray {
				return true
			}
			if color[dep] == white && dfs(dep) {
				return true
			}
		}
		color[id] = black
		return false
	}

	for id := range adj {
		if color[id] == white && dfs(id) {
			return true
		}
	}
	return false
}
