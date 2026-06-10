package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/internal/models"
	"github.com/spf13/cobra"
)

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove resources from the bakefile",
	}
	cmd.AddCommand(newRemoveSkillCommand())
	return cmd
}

func newRemoveSkillCommand() *cobra.Command {
	var target string
	var inTargets []string
	var deleteDirs bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Remove a skill from bakefile targets and optionally delete its directories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

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

			var removed, notFound []string
			for _, tn := range targetNames {
				t, ok := cfg.Targets[tn]
				if !ok {
					return fmt.Errorf("target %q not found in bakefile", tn)
				}
				newSkills := make([]string, 0, len(t.Skills))
				found := false
				for _, s := range t.Skills {
					if s == name {
						found = true
					} else {
						newSkills = append(newSkills, s)
					}
				}
				if found {
					t.Skills = newSkills
					removed = append(removed, tn)
				} else {
					notFound = append(notFound, tn)
				}
			}

			if len(removed) == 0 {
				fmt.Printf("Skill %q not found in any target.\n", name)
				return nil
			}

			if !dryRun {
				if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
					return err
				}
			}
			verb := "Removed"
			if dryRun {
				verb = "Would remove"
			}
			fmt.Printf("%s %q from targets: %s\n", verb, name, strings.Join(removed, ", "))
			if len(notFound) > 0 {
				fmt.Printf("Not present in: %s\n", strings.Join(notFound, ", "))
			}

			if !deleteDirs {
				if !dryRun {
					fmt.Printf("Skill directories preserved. Use --delete-dirs to remove them.\n")
				}
				return nil
			}

			dirs := allPlatformBaseDirs(cfg)
			deleted, missing := 0, 0
			for _, baseDir := range dirs {
				skillDir := filepath.Join(absTarget, baseDir, name)
				if _, statErr := cliStatBake(skillDir); os.IsNotExist(statErr) {
					missing++
					continue
				}
				if dryRun {
					fmt.Printf("would delete  %s/%s/\n", baseDir, name)
					deleted++
					continue
				}
				if err := os.RemoveAll(skillDir); err != nil {
					return fmt.Errorf("delete %s/%s: %w", baseDir, name, err)
				}
				fmt.Printf("deleted  %s/%s/\n", baseDir, name)
				deleted++
			}

			if dryRun {
				fmt.Printf("\nDry run: would delete %d director(ies), %d already absent.\n", deleted, missing)
			} else {
				fmt.Printf("Deleted %d director(ies), %d already absent.\n", deleted, missing)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().StringArrayVar(&inTargets, "in-target", nil, "Bake target(s) to remove skill from (default: all)")
	cmd.Flags().BoolVar(&deleteDirs, "delete-dirs", false, "Also delete skill directories from all platform dirs")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	return cmd
}

// allPlatformBaseDirs returns all unique skill base directories across every
// target in the bakefile, always including .agents/skills as the universal fallback.
func allPlatformBaseDirs(cfg *models.BakeConfig) []string {
	const agentsBase = ".agents/skills"
	dirSeen := map[string]struct{}{agentsBase: {}}
	dirs := []string{agentsBase}
	for _, t := range cfg.Targets {
		for _, p := range t.Platforms {
			d := canonicalPlatformSkillBaseDir(p)
			if _, dup := dirSeen[d]; !dup {
				dirSeen[d] = struct{}{}
				dirs = append(dirs, d)
			}
		}
	}
	return dirs
}
