package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/domehahn/sklib/spec"
	"github.com/spf13/cobra"
)

func newRenameCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename resources in the bakefile and on disk",
	}
	cmd.AddCommand(newRenameSkillCommand())
	cmd.AddCommand(newRenameTargetCommand())
	return cmd
}

func newRenameTargetCommand() *cobra.Command {
	var repoPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:               "target <old-name> <new-name>",
		Short:             "Rename a target in the bakefile",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeBakeTargets,
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName, newName := args[0], args[1]
			if oldName == newName {
				return fmt.Errorf("old and new name are identical")
			}

			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			t, exists := cfg.Targets[oldName]
			if !exists {
				return fmt.Errorf("target %q not found in bakefile", oldName)
			}
			if _, conflict := cfg.Targets[newName]; conflict {
				return fmt.Errorf("target %q already exists in bakefile", newName)
			}

			// Fix inherits references in all other targets.
			for _, other := range cfg.Targets {
				for i, inh := range other.Inherits {
					if inh == oldName {
						other.Inherits[i] = newName
					}
				}
			}

			delete(cfg.Targets, oldName)
			cfg.Targets[newName] = t

			if dryRun {
				fmt.Printf("Would rename target %q → %q\n", oldName, newName)
				return nil
			}
			if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
				return err
			}
			fmt.Printf("Renamed target %q → %q in agentic.bake.yaml\n", oldName, newName)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newRenameSkillCommand() *cobra.Command {
	var target string
	var dryRun bool

	cmd := &cobra.Command{
		Use:               "skill <old-name> <new-name>",
		Short:             "Rename a skill in all bakefile targets and move its platform directories",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeSkillNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName, newName := args[0], args[1]

			if err := spec.ValidateSkillName(newName); err != nil {
				return fmt.Errorf("invalid new name %q: use lowercase letters, digits, and hyphens", newName)
			}
			if oldName == newName {
				return fmt.Errorf("old and new name are identical")
			}

			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			// Check new name is not already used in any target.
			for tn, t := range cfg.Targets {
				for _, s := range t.Skills {
					if s == newName {
						return fmt.Errorf("skill %q already exists in target %q", newName, tn)
					}
				}
			}

			// Rename in all targets.
			var renamedInTargets []string
			for tn, t := range cfg.Targets {
				for i, s := range t.Skills {
					if s == oldName {
						t.Skills[i] = newName
						renamedInTargets = append(renamedInTargets, tn)
						break
					}
				}
			}

			if len(renamedInTargets) == 0 {
				return fmt.Errorf("skill %q not found in any target", oldName)
			}

			if cfg.SkillSources != nil {
				for i, sd := range cfg.SkillSources.Skills {
					if sd.Name == oldName {
						cfg.SkillSources.Skills[i].Name = newName
						break
					}
				}
			}

			if !dryRun {
				if err := cliDumpBakeFile(cfg, filepath.Join(absTarget, "agentic.bake.yaml")); err != nil {
					return err
				}
			}
			verb := "Renamed"
			if dryRun {
				verb = "Would rename"
			}
			fmt.Printf("%s %q → %q in targets: %s\n", verb, oldName, newName, strings.Join(renamedInTargets, ", "))

			// Move directories in all platform base dirs.
			dirs := allPlatformBaseDirs(cfg)
			moved, missing, conflicts := 0, 0, 0
			for _, baseDir := range dirs {
				srcDir := filepath.Join(absTarget, baseDir, oldName)
				dstDir := filepath.Join(absTarget, baseDir, newName)

				if _, statErr := cliStatBake(srcDir); os.IsNotExist(statErr) {
					missing++
					continue
				}
				if _, statErr := cliStatBake(dstDir); statErr == nil {
					fmt.Printf("conflict  %s/%s/ already exists — skipped\n", baseDir, newName)
					conflicts++
					continue
				}
				if dryRun {
					fmt.Printf("would move  %s/%s/  →  %s/%s/\n", baseDir, oldName, baseDir, newName)
					moved++
					continue
				}
				if err := os.Rename(srcDir, dstDir); err != nil {
					return fmt.Errorf("move %s/%s → %s/%s: %w", baseDir, oldName, baseDir, newName, err)
				}
				fmt.Printf("moved  %s/%s/  →  %s/%s/\n", baseDir, oldName, baseDir, newName)
				moved++
			}

			summary := []string{fmt.Sprintf("%d director(ies) moved", moved)}
			if missing > 0 {
				summary = append(summary, fmt.Sprintf("%d absent", missing))
			}
			if conflicts > 0 {
				summary = append(summary, fmt.Sprintf("%d conflict(s) skipped", conflicts))
			}
			fmt.Printf("\n%s: %s.\n", map[bool]string{true: "Dry run", false: "Done"}[dryRun], strings.Join(summary, ", "))
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	return cmd
}
