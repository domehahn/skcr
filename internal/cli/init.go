package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/domehahn/skcr/internal/bake"
	"github.com/domehahn/skcr/internal/models"
	"github.com/spf13/cobra"
)

var (
	cliAbsPath            = filepath.Abs
	cliMkdirAll           = os.MkdirAll
	cliParsePlatforms     = models.ParsePlatforms
	cliBuildInitialConfig = bake.BuildInitialConfig
	cliDumpBakeFile       = bake.DumpBakeFile
)

func newInitCommand() *cobra.Command {
	var target string
	var platform string
	var preset string
	var projectName string
	var ownerTeam string
	var language string
	var governanceLevel string
	var force bool
	var skills []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new agentic.bake.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			absTarget, err := cliAbsPath(target)
			if err != nil {
				return err
			}
			if err := cliMkdirAll(absTarget, 0o755); err != nil {
				return err
			}

			bakePath := filepath.Join(absTarget, "agentic.bake.yaml")
			if _, err := os.Stat(bakePath); err == nil && !force {
				return fmt.Errorf("%s already exists. Use --force to overwrite", bakePath)
			}

			platforms, err := cliParsePlatforms(platform)
			if err != nil {
				return err
			}
			if projectName == "" {
				projectName = filepath.Base(absTarget)
			}
			cfg, err := cliBuildInitialConfig(platforms, projectName, ownerTeam, language, governanceLevel, preset)
			if err != nil {
				return err
			}

			for _, name := range skills {
				if cfg.SkillSources == nil {
					cfg.SkillSources = &models.SkillSourceConfig{
						OutputDir: "skills",
						Skills:    []models.SkillSourceDefinition{},
					}
				}
				cfg.SkillSources.Skills = append(cfg.SkillSources.Skills, models.SkillSourceDefinition{
					Name: name,
				})
			}

			if err := cliDumpBakeFile(cfg, bakePath); err != nil {
				return err
			}

			fmt.Println("Created", bakePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().StringVar(&platform, "platform", "", "Comma-separated platforms, e.g. gitlab-duo,codex,github-copilot")
	cmd.Flags().StringVar(&preset, "preset", "", "Preset: minimal, gitlab, enterprise, local-ai, all")
	cmd.Flags().StringVar(&projectName, "project-name", "", "Project name used in rendered templates")
	cmd.Flags().StringVar(&ownerTeam, "owner-team", "platform-engineering", "Owning team")
	cmd.Flags().StringVar(&language, "language", "de", "Default documentation language")
	cmd.Flags().StringVar(&governanceLevel, "governance-level", "standard", "relaxed, standard, strict")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing agentic.bake.yaml")
	cmd.Flags().StringArrayVar(&skills, "skill", []string{}, "Add a skill entry to skill_sources.skills; may be repeated")

	return cmd
}
