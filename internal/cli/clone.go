package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/sklib/spec"
	"github.com/domehahn/skcr/v2/internal/models"
	"github.com/spf13/cobra"
)

func newCloneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone resources within the project",
	}
	cmd.AddCommand(newCloneSkillCommand())
	return cmd
}

func newCloneSkillCommand() *cobra.Command {
	var repoPath string
	var inTargets []string
	var dryRun bool

	cmd := &cobra.Command{
		Use:               "skill <source> <dest>",
		Short:             "Copy a skill within the project and register the clone in bakefile targets",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeSkillNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			srcName, dstName := args[0], args[1]

			if err := spec.ValidateSkillName(dstName); err != nil {
				return fmt.Errorf("invalid destination name %q: use lowercase letters, digits, and hyphens", dstName)
			}
			if srcName == dstName {
				return fmt.Errorf("source and destination names are identical")
			}

			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}
			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			srcDir := filepath.Join(absTarget, sourceBase, srcName)
			dstDir := filepath.Join(absTarget, sourceBase, dstName)

			if _, statErr := cliStatBake(srcDir); statErr != nil {
				return fmt.Errorf("source skill %q not found at %s", srcName, srcDir)
			}
			if _, statErr := cliStatBake(dstDir); statErr == nil {
				return fmt.Errorf("destination %s already exists", dstDir)
			}

			// Determine which targets to add the clone to.
			targetNames := append([]string(nil), inTargets...)
			if len(targetNames) == 0 {
				// Default: same targets that contain the source skill.
				for tn, t := range cfg.Targets {
					for _, s := range t.Skills {
						if s == srcName {
							targetNames = append(targetNames, tn)
							break
						}
					}
				}
				sort.Strings(targetNames)
			}
			if len(targetNames) == 0 {
				// Source not in any target — fall back to all targets.
				for tn := range cfg.Targets {
					targetNames = append(targetNames, tn)
				}
				sort.Strings(targetNames)
			}

			// Validate all target names exist.
			for _, tn := range targetNames {
				if _, ok := cfg.Targets[tn]; !ok {
					return fmt.Errorf("target %q not found in bakefile", tn)
				}
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "would copy  %s/%s/  →  %s/%s/\n", sourceBase, srcName, sourceBase, dstName)
				fmt.Fprintf(cmd.OutOrStdout(), "would register %q in targets: %s\n", dstName, strings.Join(targetNames, ", "))
				return nil
			}

			if err := copyDir(srcDir, dstDir); err != nil {
				return fmt.Errorf("copy %s → %s: %w", srcDir, dstDir, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "copied  %s/%s/  →  %s/%s/\n", sourceBase, srcName, sourceBase, dstName)

			var added []string
			for _, tn := range targetNames {
				t := cfg.Targets[tn]
				found := false
				for _, s := range t.Skills {
					if s == dstName {
						found = true
						break
					}
				}
				if !found {
					t.Skills = append(t.Skills, dstName)
					added = append(added, tn)
				}
			}

			if cfg.SkillSources != nil {
				found := false
				for _, sd := range cfg.SkillSources.Skills {
					if sd.Name == dstName {
						found = true
						break
					}
				}
				if !found {
					// Inherit metadata from source entry.
					var srcDef models.SkillSourceDefinition
					for _, sd := range cfg.SkillSources.Skills {
						if sd.Name == srcName {
							srcDef = sd
							break
						}
					}
					srcDef.Name = dstName
					cfg.SkillSources.Skills = append(cfg.SkillSources.Skills, srcDef)
				}
			}

			if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "registered %q in targets: %s\n", dstName, strings.Join(added, ", "))
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().StringArrayVar(&inTargets, "in-target", nil, "Targets to register clone in (default: same as source)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	_ = cmd.RegisterFlagCompletionFunc("in-target", completeBakeTargets)
	return cmd
}
