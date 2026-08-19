package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/v2/internal/bake"
	"github.com/domehahn/skcr/v2/internal/renderer"
	"github.com/domehahn/skcr/v2/internal/skillversion"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Show detailed information about a bakefile resource",
	}
	cmd.AddCommand(newInspectSkillCommand())
	cmd.AddCommand(newInspectTargetCommand())
	return cmd
}

func newInspectTargetCommand() *cobra.Command {
	var repoPath string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:               "target <name>",
		Short:             "Show resolved configuration for a bake target",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeBakeTargets,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			absTarget, err := filepath.Abs(repoPath)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}

			raw, exists := cfg.Targets[name]
			if !exists {
				return fmt.Errorf("target %q not found in bakefile", name)
			}

			resolved, err := bake.ResolveTarget(cfg, name)
			if err != nil {
				return err
			}

			files, err := renderer.RenderFiles(cfg, resolved)
			if err != nil {
				return err
			}

			sort.Strings(resolved.Skills)

			if jsonOut {
				out := map[string]any{
					"name":           name,
					"description":    raw.Description,
					"platforms":      resolved.Platforms,
					"inherits":       raw.Inherits,
					"skills":         resolved.Skills,
					"rendered_files": len(files),
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Target:         %s\n", name)
			if raw.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Description:    %s\n", raw.Description)
			}
			if len(raw.Inherits) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Inherits:       %s\n", strings.Join(raw.Inherits, ", "))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Platforms:      %s\n", strings.Join(resolved.Platforms, ", "))
			fmt.Fprintf(cmd.OutOrStdout(), "Skills (%d):     %s\n", len(resolved.Skills), strings.Join(resolved.Skills, ", "))
			fmt.Fprintf(cmd.OutOrStdout(), "Rendered files: %d\n", len(files))
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoPath, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func newInspectSkillCommand() *cobra.Command {
	var target string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:               "skill <name>",
		Short:             "Show full metadata for a skill",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeSkillNames,
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

			sourceBase := skillSourceOutputDir(cfg.SkillSources)
			skillDir := filepath.Join(absTarget, sourceBase, name)

			// Collect which bake targets include this skill.
			var inTargets []string
			for tn, t := range cfg.Targets {
				for _, s := range t.Skills {
					if s == name {
						inTargets = append(inTargets, tn)
						break
					}
				}
			}
			sort.Strings(inTargets)

			// Read bakefile skill_sources entry if present.
			var sourceEntry map[string]any
			if cfg.SkillSources != nil {
				for _, sd := range cfg.SkillSources.Skills {
					if sd.Name == name {
						sourceEntry = map[string]any{
							"description":     sd.Description,
							"owner":           sd.Owner,
							"initial_version": sd.InitialVersion,
							"compatible_with": sd.CompatibleWith,
							"license":         sd.License,
						}
						break
					}
				}
			}

			// Read SKILL.md frontmatter.
			skillMDPath := filepath.Join(skillDir, "SKILL.md")
			skillMDData, readErr := os.ReadFile(skillMDPath)
			var frontmatter map[string]any
			if readErr == nil {
				content := string(skillMDData)
				if strings.HasPrefix(content, "---\n") {
					end := strings.Index(content[4:], "\n---")
					if end >= 0 {
						raw := content[4 : 4+end]
						_ = yaml.Unmarshal([]byte(raw), &frontmatter)
					}
				}
			}

			// Read VERSION file.
			versionFileData, _ := os.ReadFile(filepath.Join(skillDir, "VERSION"))
			versionFile := strings.TrimSpace(string(versionFileData))

			// Read CHANGELOG entries (first 5 lines after header).
			changelog, _ := skillversion.Changelog(skillDir)

			if jsonOut {
				out := map[string]any{
					"name":          name,
					"path":          skillDir,
					"in_targets":    inTargets,
					"version_file":  versionFile,
					"frontmatter":   frontmatter,
					"skill_sources": sourceEntry,
					"changelog":     changelog,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Skill:     %s\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "Path:      %s\n", skillDir)
			if len(inTargets) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Targets:   %s\n", strings.Join(inTargets, ", "))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Targets:   (none — not in any bake target)\n")
			}
			fmt.Fprintln(cmd.OutOrStdout())

			if readErr != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "SKILL.md:  not found (%v)\n", readErr)
			} else {
				printFrontmatterField(cmd, frontmatter, "version")
				printFrontmatterField(cmd, frontmatter, "since")
				printFrontmatterField(cmd, frontmatter, "last_modified")
				printFrontmatterField(cmd, frontmatter, "description")
				printFrontmatterField(cmd, frontmatter, "owner")
				printFrontmatterField(cmd, frontmatter, "license")
				printFrontmatterField(cmd, frontmatter, "compatible_with")
			}

			if versionFile != "" && versionFile != frontmatter["version"] {
				fmt.Fprintf(cmd.OutOrStdout(), "\nWARN: VERSION file (%s) differs from SKILL.md version (%v) — run: skcr version sync\n", versionFile, frontmatter["version"])
			}

			if len(changelog) > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "Changelog (recent):")
				limit := 5
				if len(changelog) < limit {
					limit = len(changelog)
				}
				for _, e := range changelog[:limit] {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s  %s  %s\n", e.Date, e.Version, e.Change)
				}
				if len(changelog) > 5 {
					fmt.Fprintf(cmd.OutOrStdout(), "  … %d more entries\n", len(changelog)-5)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print JSON output")
	return cmd
}

func printFrontmatterField(cmd *cobra.Command, fm map[string]any, key string) {
	val, ok := fm[key]
	if !ok || val == nil || val == "" {
		return
	}
	label := strings.ReplaceAll(key, "_", " ")
	label = strings.ToTitle(label[:1]) + label[1:]
	switch v := val.(type) {
	case []any:
		strs := make([]string, 0, len(v))
		for _, item := range v {
			strs = append(strs, fmt.Sprintf("%v", item))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-14s %s\n", label+":", strings.Join(strs, ", "))
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "%-14s %v\n", label+":", v)
	}
}
