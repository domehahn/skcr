package cli

import (
	"encoding/json"
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
	var withBakefileSkills bool
	var skillsFrom string
	var jsonOut bool

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

			type targetEntry struct {
				Name             string   `json:"name"`
				Description      string   `json:"description,omitempty"`
				Platforms        []string `json:"platforms,omitempty"`
				Inherits         []string `json:"inherits,omitempty"`
				Skills           []string `json:"skills,omitempty"`
				RenderedFiles    int      `json:"rendered_files,omitempty"`
				CompatibleSkills []string `json:"compatible_skills,omitempty"`
			}

			var entries []targetEntry
			for _, name := range names {
				t := cfg.Targets[name]
				skills := append([]string(nil), t.Skills...)
				sort.Strings(skills)

				entry := targetEntry{
					Name:        name,
					Description: t.Description,
					Platforms:   t.Platforms,
					Inherits:    t.Inherits,
					Skills:      skills,
				}

				if skillState != nil {
					resolved, err := bake.ResolveTarget(cfg, name)
					if err != nil {
						return err
					}
					files, err := renderer.RenderFiles(cfg, resolved)
					if err != nil {
						return err
					}
					entry.RenderedFiles = len(files)
					compatible := skilllock.FilterByPlatforms(skillState.Skills, resolved.Platforms)
					for _, sk := range compatible {
						entry.CompatibleSkills = append(entry.CompatibleSkills, sk.Name)
					}
					sort.Strings(entry.CompatibleSkills)
				}
				entries = append(entries, entry)
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}

			switch {
			case skillState != nil:
				fmt.Println("Target\tDescription\tPlatforms / Inherits\tGenerated files\tCompatible skills")
			case withBakefileSkills:
				fmt.Println("Target\tDescription\tPlatforms / Inherits\tSkills")
			default:
				fmt.Println("Target\tDescription\tPlatforms / Inherits")
			}
			for _, e := range entries {
				value := strings.Join(e.Platforms, ", ")
				if value == "" {
					value = strings.Join(e.Inherits, ", ")
				}
				switch {
				case skillState != nil:
					fmt.Printf("%s\t%s\t%s\t%d\t%s\n", e.Name, e.Description, value, e.RenderedFiles, strings.Join(e.CompatibleSkills, ", "))
				case withBakefileSkills:
					fmt.Printf("%s\t%s\t%s\t%s\n", e.Name, e.Description, value, strings.Join(e.Skills, ", "))
				default:
					fmt.Printf("%s\t%s\t%s\n", e.Name, e.Description, value)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().BoolVar(&withSkills, "with-skills", false, "Show compatible locked skills")
	cmd.Flags().BoolVar(&withBakefileSkills, "with-bakefile-skills", false, "Show skills declared in each bakefile target")
	cmd.Flags().StringVar(&skillsFrom, "skills-from", "", "Read skpm locked skills from agent-skills.lock")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}
