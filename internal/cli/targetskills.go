package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

func newTargetSkillsCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "target-skills <target>",
		Short: "List all skills included in a bakefile target after resolving inherits",
		Long: `Resolves the full inherits: chain for <target> and prints every skill
that would be included when running 'skcr bake <target>'.

Duplicate skills (inherited from multiple parents) are deduplicated.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetName := args[0]

			abs, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("resolve target: %w", err)
			}

			cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
			if err != nil {
				return fmt.Errorf("load bakefile: %w", err)
			}

			resolved, err := cliResolveTarget(cfg, targetName)
			if err != nil {
				return fmt.Errorf("resolve target %q: %w", targetName, err)
			}

			skills := deduplicateStrings(resolved.Skills)
			sort.Strings(skills)

			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"target": targetName,
					"skills": skills,
				})
			}

			if len(skills) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Target %q has no skills.\n", targetName)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Target: %s (%d skill(s))\n\n", targetName, len(skills))
			for _, s := range skills {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", s)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", ".", "Path to the repository root containing agentic.bake.yaml")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func deduplicateStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
