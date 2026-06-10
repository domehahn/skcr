package cli

import (
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

// completeBakeTargets returns the target names from the nearest agentic.bake.yaml
// for use as dynamic completions on the bake command's positional argument.
func completeBakeTargets(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dir, _ := cmd.Flags().GetString("target")
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names := make([]string, 0, len(cfg.Targets))
	for name := range cfg.Targets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeSkillNames returns skill names from all bakefile targets for use as
// dynamic completions on commands that accept a skill name argument.
func completeSkillNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dir, _ := cmd.Flags().GetString("target")
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	cfg, err := cliLoadBakeFile(filepath.Join(abs, "agentic.bake.yaml"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	seen := map[string]struct{}{}
	var names []string
	for _, t := range cfg.Targets {
		for _, s := range t.Skills {
			if _, dup := seen[s]; !dup {
				seen[s] = struct{}{}
				names = append(names, s)
			}
		}
	}
	sort.Strings(names)
	return names, cobra.ShellCompDirectiveNoFileComp
}
