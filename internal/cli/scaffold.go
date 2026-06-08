package cli

import (
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/scaffold"
	"github.com/spf13/cobra"
)

var (
	cliAbsPathScaffold = filepath.Abs
	cliWriteSkill      = scaffold.WriteSkill
	cliWriteSkillSafe  = scaffold.WriteSkillSafe
)

func newScaffoldCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Scaffold versionable agentic project structures",
	}
	cmd.AddCommand(newScaffoldSkillCommand())
	cmd.AddCommand(newScaffoldSkillsCommand())
	return cmd
}

func newScaffoldSkillCommand() *cobra.Command {
	var outputDir string
	var version string
	var description string
	var owner string
	var platforms []string
	var license string
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Scaffold a versionable AI agent skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			absOutput, err := cliAbsPathScaffold(outputDir)
			if err != nil {
				return err
			}
			files, err := cliWriteSkill(scaffold.SkillOptions{
				Name:        args[0],
				OutputDir:   absOutput,
				Version:     version,
				Description: description,
				Owner:       owner,
				Platforms:   platforms,
				License:     license,
				Force:       force,
				DryRun:      dryRun,
			})
			if err != nil {
				return err
			}
			skillPath := filepath.Join(outputDir, args[0])
			if dryRun {
				fmt.Printf("Skill scaffold plan: %s\n", args[0])
				for _, file := range files {
					fmt.Println("  create", file.Path)
				}
				fmt.Println("Dry run only. Re-run without --dry-run to write files.")
				return nil
			}
			fmt.Printf("Created %s/\n", skillPath)
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  skpm validate %s\n", skillPath)
			fmt.Printf("  skpm package %s\n", skillPath)
			fmt.Printf("  skpm publish %s --source <registry>\n", skillPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "Directory where the skill directory should be created")
	cmd.Flags().StringVar(&version, "version", "0.1.0", "Initial semver version")
	cmd.Flags().StringVar(&description, "description", "", "Skill description for skill.yaml")
	cmd.Flags().StringVar(&owner, "owner", "", "Skill owner")
	cmd.Flags().StringArrayVar(&platforms, "platform", []string{"claude-code", "codex"}, "Compatible platform; may be repeated")
	cmd.Flags().StringVar(&license, "license", "MIT", "Skill license")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing scaffold files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview files without writing")
	return cmd
}

func newScaffoldSkillsCommand() *cobra.Command {
	var target string
	var bakefile string
	var dryRun bool
	var force bool

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Scaffold canonical skill source directories from agentic.bake.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := cliAbsPathScaffold(target)
			if err != nil {
				return err
			}
			bakefilePath := bakefile
			if bakefilePath == "" {
				bakefilePath = filepath.Join(absTarget, "agentic.bake.yaml")
			}
			cfg, err := cliLoadBakeFile(bakefilePath)
			if err != nil {
				return err
			}
			if cfg.SkillSources == nil || len(cfg.SkillSources.Skills) == 0 {
				fmt.Println("No skill_sources.skills defined in agentic.bake.yaml. Nothing to scaffold.")
				return nil
			}

			ss := cfg.SkillSources
			outputDir := filepath.Join(absTarget, ss.OutputDir)

			seen := map[string]struct{}{}
			for _, skillDef := range ss.Skills {
				if _, dup := seen[skillDef.Name]; dup {
					return fmt.Errorf("duplicate skill name %q in skill_sources", skillDef.Name)
				}
				seen[skillDef.Name] = struct{}{}
			}

			anyCreated := false
			for _, skillDef := range ss.Skills {
				opts := skillDefToScaffoldOpts(skillDef, ss, outputDir, dryRun, force)
				result, err := cliWriteSkillSafe(opts)
				if err != nil {
					return fmt.Errorf("skill %s: %w", skillDef.Name, err)
				}
				skillPath := filepath.Join(ss.OutputDir, skillDef.Name)
				if dryRun {
					fmt.Printf("Skill scaffold plan: %s\n", skillDef.Name)
					for _, f := range result.Created {
						fmt.Println("  create", f.Path)
					}
					continue
				}
				for _, f := range result.Created {
					fmt.Println("create", f.Path)
					anyCreated = true
				}
				for _, f := range result.Skipped {
					fmt.Println("skip  ", f.Path)
				}
				if len(result.Created) > 0 || len(result.Skipped) > 0 {
					fmt.Printf("\nNext steps for %s:\n", skillPath)
					fmt.Printf("  skpm validate %s\n", skillPath)
					fmt.Printf("  skpm package %s\n", skillPath)
					fmt.Printf("  skpm publish %s --source <registry>\n\n", skillPath)
				}
			}
			if dryRun {
				fmt.Println("Dry run only. Re-run without --dry-run to write files.")
			} else if !anyCreated {
				fmt.Println("All skill source files already exist. Use --force to overwrite.")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().StringVar(&bakefile, "bakefile", "", "Path to agentic.bake.yaml (default: <target>/agentic.bake.yaml)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview files without writing")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing scaffold files")
	return cmd
}

// skillDefToScaffoldOpts merges skill source config defaults into a SkillOptions struct.
func skillDefToScaffoldOpts(def models.SkillSourceDefinition, ss *models.SkillSourceConfig, outputDir string, dryRun, force bool) scaffold.SkillOptions {
	opts := scaffold.SkillOptions{
		Name:      def.Name,
		OutputDir: outputDir,
		DryRun:    dryRun,
		Force:     force,
	}
	// Apply definition fields, falling back to defaults.
	opts.Version = firstNonEmpty(def.Version, ss.Defaults.Version, "0.1.0")
	opts.Owner = firstNonEmpty(def.Owner, ss.Defaults.Owner)
	opts.License = firstNonEmpty(def.License, ss.Defaults.License, "MIT")
	opts.Description = def.Description

	if len(def.CompatibleWith) > 0 {
		opts.Platforms = def.CompatibleWith
	} else if len(ss.Defaults.CompatibleWith) > 0 {
		opts.Platforms = ss.Defaults.CompatibleWith
	}
	return opts
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
