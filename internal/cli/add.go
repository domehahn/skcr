package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/sklib/spec"
	"github.com/domehahn/skcr/internal/bake"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add resources to the bakefile",
	}
	cmd.AddCommand(newAddSkillCommand())
	return cmd
}

func newAddSkillCommand() *cobra.Command {
	var target string
	var inTargets []string
	var noScaffold bool

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Add a skill to bakefile targets and scaffold its directory structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := spec.ValidateSkillName(name); err != nil {
				return fmt.Errorf("invalid skill name %q: use lowercase letters, digits, and hyphens", name)
			}

			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			targetNames := append([]string(nil), inTargets...)
			if len(targetNames) == 0 {
				for n := range cfg.Targets {
					targetNames = append(targetNames, n)
				}
				sort.Strings(targetNames)
			}

			var added, alreadyIn []string
			for _, tn := range targetNames {
				t, ok := cfg.Targets[tn]
				if !ok {
					return fmt.Errorf("target %q not found in bakefile", tn)
				}
				found := false
				for _, s := range t.Skills {
					if s == name {
						found = true
						break
					}
				}
				if found {
					alreadyIn = append(alreadyIn, tn)
				} else {
					t.Skills = append(t.Skills, name)
					added = append(added, tn)
				}
			}

			if len(added) == 0 {
				fmt.Printf("Skill %q is already present in all targets.\n", name)
				return nil
			}

			if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
				return err
			}
			fmt.Printf("Added %q to targets: %s\n", name, strings.Join(added, ", "))
			if len(alreadyIn) > 0 {
				fmt.Printf("Already present in:  %s\n", strings.Join(alreadyIn, ", "))
			}

			if noScaffold {
				return nil
			}

			// Collect unique platforms across all affected targets.
			platSeen := map[string]struct{}{}
			var platforms []string
			for _, tn := range added {
				resolved, err := bake.ResolveTarget(cfg, tn)
				if err != nil {
					continue
				}
				for _, p := range resolved.Platforms {
					if _, dup := platSeen[p]; !dup {
						platSeen[p] = struct{}{}
						platforms = append(platforms, p)
					}
				}
			}

			created, skipped, err := scaffoldTargetSkills(absTarget, []string{name}, cfg.SkillSources, platforms, false)
			if err != nil {
				return err
			}
			fmt.Printf("Scaffold: %d files created, %d already existed.\n", created, skipped)
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().StringArrayVar(&inTargets, "in-target", nil, "Bake target(s) to add skill to (default: all)")
	cmd.Flags().BoolVar(&noScaffold, "no-scaffold", false, "Update bakefile only, skip directory scaffolding")
	_ = cmd.RegisterFlagCompletionFunc("in-target", completeBakeTargets)
	return cmd
}
