package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/internal/bake"
	"github.com/domehahn/skcr/internal/renderer"
	"github.com/domehahn/skcr/internal/skilllock"
	"github.com/spf13/cobra"
)

func newListTargetsCommand() *cobra.Command {
	var target string
	var withSkills bool
	var skillsFrom string

	cmd := &cobra.Command{
		Use:   "list-targets",
		Short: "List available bake targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filepath.Join(target, "agentic.bake.yaml")
			cfg, err := bake.LoadBakeFile(path)
			if err != nil {
				return err
			}
			var skillState *skilllock.LockState
			if withSkills || skillsFrom != "" {
				source := skillsFrom
				if source == "" && cfg.Skills != nil {
					source = cfg.Skills.Source
				}
				if source == "" {
					source = "agent-skills.lock"
				}
				sourcePath := source
				if !filepath.IsAbs(sourcePath) {
					sourcePath = filepath.Join(target, sourcePath)
				}
				skillState, err = skilllock.Load(sourcePath)
				if err != nil {
					return err
				}
			}

			names := make([]string, 0, len(cfg.Targets))
			for name := range cfg.Targets {
				names = append(names, name)
			}
			sort.Strings(names)

			if skillState != nil {
				fmt.Println("Target\tDescription\tPlatforms / Inherits\tGenerated files\tCompatible skills")
			} else {
				fmt.Println("Target\tDescription\tPlatforms / Inherits")
			}
			for _, name := range names {
				t := cfg.Targets[name]
				value := strings.Join(t.Platforms, ", ")
				if value == "" {
					value = strings.Join(t.Inherits, ", ")
				}
				if skillState == nil {
					fmt.Printf("%s\t%s\t%s\n", name, t.Description, value)
					continue
				}
				resolved, err := bake.ResolveTarget(cfg, name)
				if err != nil {
					return err
				}
				files, err := renderer.RenderFiles(cfg, resolved)
				if err != nil {
					return err
				}
				compatible := skilllock.FilterByPlatforms(skillState.Skills, resolved.Platforms)
				skillNames := []string{}
				for _, skill := range compatible {
					skillNames = append(skillNames, skill.Name)
				}
				sort.Strings(skillNames)
				fmt.Printf("%s\t%s\t%s\t%d\t%s\n", name, t.Description, value, len(files), strings.Join(skillNames, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().BoolVar(&withSkills, "with-skills", false, "Show compatible locked skills")
	cmd.Flags().StringVar(&skillsFrom, "skills-from", "", "Read skpm locked skills from agent-skills.lock")
	return cmd
}
