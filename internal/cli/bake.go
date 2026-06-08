package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/agentic-template-kit/skcr/internal/bake"
	"github.com/agentic-template-kit/skcr/internal/lockfile"
	"github.com/agentic-template-kit/skcr/internal/models"
	"github.com/agentic-template-kit/skcr/internal/renderer"
	"github.com/agentic-template-kit/skcr/internal/skilllock"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

const maxDiffLinesPerFile = 120

var (
	cliAbsPathBake    = filepath.Abs
	cliLoadBakeFile   = bake.LoadBakeFile
	cliResolveTarget  = bake.ResolveTarget
	cliRenderFiles    = renderer.RenderFiles
	cliRenderWithOpts = renderer.RenderFilesWithOptions
	cliLoadLockfile   = lockfile.LoadLockfile
	cliLoadSkillLock  = skilllock.Load
	cliWriteLockfile  = lockfile.WriteLockfile
	cliReadFile       = os.ReadFile
	cliMkdirAllBake   = os.MkdirAll
	cliWriteFile      = os.WriteFile
	cliRemoveBake     = os.Remove
	cliSymlinkBake    = os.Symlink
	cliStatBake       = os.Stat
)

func newBakeCommand() *cobra.Command {
	var target string
	var plan bool
	var detailedDiff bool
	var write bool
	var skillsFrom string
	var skillsMode string
	var platform string

	cmd := &cobra.Command{
		Use:   "bake [target]",
		Short: "Render files for a target and preview or write them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetName := "default"
			if len(args) > 0 {
				targetName = args[0]
			}
			if !plan && !write {
				plan = true
			}

			absTarget, err := cliAbsPathBake(target)
			if err != nil {
				return err
			}

			cfg, err := cliLoadBakeFile(filepath.Join(absTarget, "agentic.bake.yaml"))
			if err != nil {
				return err
			}
			resolved, err := cliResolveTarget(cfg, targetName)
			if err != nil {
				return err
			}
			if platform != "" {
				platforms, err := models.ParsePlatforms(platform)
				if err != nil {
					return err
				}
				resolved.Platforms = filterPlatforms(resolved.Platforms, platforms)
			}
			renderOpts, skillFiles, err := loadSkillRenderOptions(absTarget, cfg, resolved, skillsFrom, skillsMode)
			if err != nil {
				return err
			}
			var files []models.RenderedFile
			if len(renderOpts.LockedSkills) == 0 && renderOpts.SkillsMode == models.SkillModeReference {
				files, err = cliRenderFiles(cfg, resolved)
			} else {
				files, err = cliRenderWithOpts(cfg, resolved, renderOpts)
			}
			if err != nil {
				return err
			}
			files = append(files, skillFiles...)

			lock, err := cliLoadLockfile(absTarget)
			if err != nil {
				return err
			}
			stateFiles := lockfile.ManagedFilesByPath(lock)
			plannedFiles := map[string]models.RenderedFile{}
			for _, file := range files {
				plannedFiles[file.Destination] = file
			}

			fmt.Printf("Bake target: %s\n", targetName)
			fmt.Println("Action\tPlatform\tPath")
			for _, rendered := range sortedRendered(files) {
				path := filepath.Join(absTarget, rendered.Destination)
				action := "create"
				if current, err := cliReadFile(path); err == nil {
					if string(current) == rendered.Content {
						action = "unchanged"
					} else {
						action = "update"
					}
				}
				fmt.Printf("%s\t%s\t%s\n", action, rendered.Platform, rendered.Destination)
			}

			stateCounts := map[string]int{"create": 0, "update": 0, "delete": 0, "noop": 0}
			fmt.Println("\nState Plan (.agentic-template.lock)")
			fmt.Println("Action\tPlatform\tPath")

			for _, path := range sortedKeys(plannedFiles) {
				rendered := plannedFiles[path]
				checksum := renderedChecksum(rendered)
				prev, ok := stateFiles[path]
				action := "create"
				if ok {
					if checksumValue(prev) == checksum {
						action = "noop"
					} else {
						action = "update"
					}
				}
				stateCounts[action]++
				fmt.Printf("%s\t%s\t%s\n", action, rendered.Platform, path)
			}
			for _, path := range sortedMapKeys(stateFiles) {
				if _, exists := plannedFiles[path]; exists {
					continue
				}
				stateCounts["delete"]++
				platform := "-"
				if p, ok := stateFiles[path]["platform"].(string); ok && p != "" {
					platform = p
				}
				fmt.Printf("delete\t%s\t%s\n", platform, path)
			}

			fmt.Printf("\nPlan summary: %d to create, %d to update, %d to delete, %d unchanged in state.\n", stateCounts["create"], stateCounts["update"], stateCounts["delete"], stateCounts["noop"])

			if plan {
				changed := []string{}
				for path, rendered := range plannedFiles {
					current, err := cliReadFile(filepath.Join(absTarget, path))
					if err != nil {
						continue
					}
					if string(current) != rendered.Content {
						changed = append(changed, path)
					}
				}
				sort.Strings(changed)
				for _, path := range changed {
					rendered := plannedFiles[path]
					current, err := cliReadFile(filepath.Join(absTarget, path))
					if err != nil {
						continue
					}
					diffText := unifiedDiff(string(current), rendered.Content, path, !detailedDiff)
					if strings.TrimSpace(diffText) == "" {
						continue
					}
					fmt.Printf("\nDiff: %s\n%s\n", path, diffText)
				}

				deleted := []string{}
				for path := range stateFiles {
					if _, ok := plannedFiles[path]; !ok {
						deleted = append(deleted, path)
					}
				}
				sort.Strings(deleted)
				if len(deleted) > 0 {
					fmt.Println("\nState-only files (would be removed from state if applied):")
					for _, path := range deleted {
						fmt.Println("-", path)
					}
				}

				fmt.Println("\nDry run only. Use --write to write files.")
				return nil
			}

			for _, rendered := range files {
				path := filepath.Join(absTarget, rendered.Destination)
				if err := cliMkdirAllBake(filepath.Dir(path), 0o755); err != nil {
					return err
				}
				if rendered.LinkTarget != "" {
					targetPath := rendered.LinkTarget
					if !filepath.IsAbs(targetPath) {
						targetPath = filepath.Join(absTarget, targetPath)
					}
					if err := cliRemoveBake(path); err != nil && !os.IsNotExist(err) {
						return err
					}
					if err := cliSymlinkBake(targetPath, path); err != nil {
						return err
					}
					continue
				}
				if err := cliWriteFile(path, []byte(rendered.Content), 0o644); err != nil {
					return err
				}
			}

			if err := cliWriteLockfile(absTarget, files, targetName); err != nil {
				return err
			}
			fmt.Printf("Wrote %d files and .agentic-template.lock\n", len(files))
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", ".", "Target repository path")
	cmd.Flags().BoolVar(&plan, "plan", false, "Show generated files without writing")
	cmd.Flags().BoolVar(&detailedDiff, "detailed-diff", false, "Show full unified diffs in --plan")
	cmd.Flags().BoolVar(&write, "write", false, "Write files to target repository")
	cmd.Flags().StringVar(&skillsFrom, "skills-from", "", "Read skpm locked skills from agent-skills.lock")
	cmd.Flags().StringVar(&skillsMode, "skills-mode", "", "Skill integration mode: reference, copy, link, embed")
	cmd.Flags().StringVar(&platform, "platform", "", "Render only the selected canonical platform")

	return cmd
}

func loadSkillRenderOptions(root string, cfg *models.BakeConfig, target *models.TargetConfig, from, mode string) (renderer.Options, []models.RenderedFile, error) {
	opts := renderer.Options{SkillsMode: models.SkillModeReference}
	if cfg != nil && cfg.Skills != nil {
		opts.SkillsMode = cfg.Skills.Mode
	}
	if mode != "" {
		opts.SkillsMode = mode
	}
	if err := validateSkillsMode(opts.SkillsMode); err != nil {
		return opts, nil, err
	}

	source := from
	explicit := source != ""
	if source == "" && cfg != nil && cfg.Skills != nil {
		source = cfg.Skills.Source
	}
	if source == "" {
		return opts, nil, nil
	}
	sourcePath := source
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(root, sourcePath)
	}
	if !explicit {
		if _, err := cliStatBake(sourcePath); os.IsNotExist(err) {
			return opts, nil, nil
		} else if err != nil {
			return opts, nil, err
		}
	}
	state, err := cliLoadSkillLock(sourcePath)
	if err != nil {
		return opts, nil, err
	}
	platforms := target.Platforms
	if cfg != nil && cfg.Skills != nil && len(cfg.Skills.Platforms) > 0 {
		configured, err := models.NormalizePlatforms(cfg.Skills.Platforms)
		if err != nil {
			return opts, nil, err
		}
		platforms = filterPlatforms(platforms, configured)
	}
	filtered := skilllock.FilterByPlatforms(state.Skills, platforms)
	opts.LockedSkills = skilllock.References(filtered)
	files, err := skilllock.SkillFiles(root, filtered, opts.SkillsMode, platforms)
	if err != nil {
		return opts, nil, err
	}
	return opts, files, nil
}

func validateSkillsMode(mode string) error {
	switch mode {
	case "", models.SkillModeReference, models.SkillModeCopy, models.SkillModeLink, models.SkillModeEmbed:
		return nil
	default:
		return fmt.Errorf("unsupported skills mode %q", mode)
	}
}

func filterPlatforms(current, selected []string) []string {
	allowed := map[string]struct{}{}
	for _, platform := range selected {
		allowed[platform] = struct{}{}
	}
	filtered := []string{}
	for _, platform := range current {
		if _, ok := allowed[platform]; ok {
			filtered = append(filtered, platform)
		}
	}
	return filtered
}

func renderedChecksum(rendered models.RenderedFile) string {
	if rendered.LinkTarget != "" {
		return lockfile.Sha256Text("link:" + rendered.LinkTarget)
	}
	return lockfile.Sha256Text(rendered.Content)
}

func sortedRendered(files []models.RenderedFile) []models.RenderedFile {
	copyFiles := make([]models.RenderedFile, len(files))
	copy(copyFiles, files)
	sort.Slice(copyFiles, func(i, j int) bool {
		return copyFiles[i].Destination < copyFiles[j].Destination
	})
	return copyFiles
}

func sortedKeys(m map[string]models.RenderedFile) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapKeys(m map[string]map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func checksumValue(entry map[string]any) string {
	if value, ok := entry["checksum"].(string); ok {
		return value
	}
	return ""
}

func unifiedDiff(oldText, newText, path string, truncate bool) string {
	d := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldText),
		B:        difflib.SplitLines(newText),
		FromFile: "a/" + path,
		ToFile:   "b/" + path,
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(d)
	if !truncate {
		return text
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= maxDiffLinesPerFile {
		return text
	}
	trimmed := append(lines[:maxDiffLinesPerFile], "... (diff truncated)")
	return strings.Join(trimmed, "\n")
}
