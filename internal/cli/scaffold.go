package cli

import (
	"fmt"
	"path/filepath"

	"github.com/domehahn/skcr/internal/scaffold"
	"github.com/spf13/cobra"
)

var (
	cliAbsPathScaffold = filepath.Abs
	cliWriteSkill      = scaffold.WriteSkill
)

func newScaffoldCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Scaffold versionable agentic project structures",
	}
	cmd.AddCommand(newScaffoldSkillCommand())
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
			if dryRun {
				fmt.Printf("Skill scaffold plan: %s\n", args[0])
				for _, file := range files {
					fmt.Println("create", file.Path)
				}
				fmt.Println("Dry run only. Re-run without --dry-run to write files.")
				return nil
			}
			fmt.Printf("Created skill scaffold %s with %d files\n", filepath.Join(absOutput, args[0]), len(files))
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
