package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/domehahn/skcr/internal/bake"
	"github.com/domehahn/skcr/internal/lockfile"
	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/renderer"
	"github.com/domehahn/skcr/internal/scaffold"
	"github.com/domehahn/skcr/internal/skilllock"
	"github.com/domehahn/skcr/internal/skillversion"
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
		Use:               "bake [target]",
		Short:             "Render files for a target and preview or write them",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeBakeTargets,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			targetName := "default"
			if len(args) > 0 {
				targetName = args[0]
			} else if _, ok := cfg.Targets[targetName]; !ok {
				// No explicit arg and no "default" target — try "all", then sole target, then error.
				if _, ok := cfg.Targets["all"]; ok {
					targetName = "all"
				} else {
					names := make([]string, 0, len(cfg.Targets))
					for n := range cfg.Targets {
						names = append(names, n)
					}
					if len(names) == 1 {
						targetName = names[0]
					} else {
						sort.Strings(names)
						return fmt.Errorf("no %q target found; specify one of: %s", "default", strings.Join(names, ", "))
					}
				}
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
			if renderPlatformFilesEnabled(resolved) {
				files, err = cliRenderWithOpts(cfg, resolved, renderOpts)
				if err != nil {
					return err
				}
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

			if len(resolved.Skills) > 0 {
				// Collect unique base directories to scaffold.
				dirSeen := map[string]struct{}{}
				var planDirs []string
				for _, p := range resolved.Platforms {
					d := canonicalPlatformSkillBaseDir(p)
					if _, dup := dirSeen[d]; !dup {
						dirSeen[d] = struct{}{}
						planDirs = append(planDirs, d)
					}
				}
				if _, ok := dirSeen[".agents/skills"]; !ok {
					planDirs = append(planDirs, ".agents/skills")
				}

				fmt.Printf("\nSkill Sources Plan (%d director(ies))\n", len(planDirs))
				fmt.Println("Action\tDir\tSkill")
				for _, baseDir := range planDirs {
					for _, name := range resolved.Skills {
						skillDir := filepath.Join(absTarget, baseDir, name)
						if _, statErr := cliStatBake(skillDir); os.IsNotExist(statErr) {
							fmt.Printf("create\t%s\t%s\n", baseDir, name)
						} else {
							fmt.Printf("exists\t%s\t%s\n", baseDir, name)
						}
					}
				}
			}

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

			if len(resolved.Skills) > 0 {
				created, skipped, err := scaffoldTargetSkills(absTarget, resolved.Skills, cfg.SkillSources, resolved.Platforms, false)
				if err != nil {
					return err
				}
				if created > 0 || skipped > 0 {
					fmt.Printf("Skill sources: %d created, %d skipped (existing files preserved)\n", created, skipped)
				}
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
			if err := syncRenderedSkillArtifacts(absTarget, files); err != nil {
				return err
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

func syncRenderedSkillArtifacts(root string, files []models.RenderedFile) error {
	seen := map[string]struct{}{}
	for _, rendered := range files {
		if filepath.Base(rendered.Destination) != "SKILL.md" {
			continue
		}
		skillDir := filepath.Dir(filepath.Join(root, rendered.Destination))
		if _, ok := seen[skillDir]; ok {
			continue
		}
		seen[skillDir] = struct{}{}
		if _, err := skillversion.SyncArtifacts(skillDir); err != nil {
			return fmt.Errorf("sync skill version artifacts for %s: %w", rendered.Destination, err)
		}
	}
	return nil
}

// canonicalPlatformSkillBaseDir returns the base directory where skills are stored for a platform.
// Platforms with dedicated directories use those; all others share .agents/skills/.
func canonicalPlatformSkillBaseDir(platform string) string {
	switch platform {
	case "claude-code":
		return ".claude/skills"
	case "github-copilot":
		return ".github/skills"
	case "cursor":
		return ".cursor/skills"
	case "junie":
		return ".junie/skills"
	case "gemini-cli":
		return ".gemini/skills"
	case "roo-code":
		return ".roo/skills"
	case "kiro":
		return ".kiro/skills"
	case "opencode":
		return ".opencode/skills"
	case "openhands":
		return ".openhands/skills"
	case "windsurf":
		return ".windsurf/skills"
	case "gitlab-duo":
		return "skills"
	case "ollama":
		return ".ollama/skills"
	default:
		return ".agents/skills"
	}
}

// skillSourceOutputDir returns the configured output directory for skill sources,
// defaulting to .agents/skills if not set.
func skillSourceOutputDir(ss *models.SkillSourceConfig) string {
	if ss != nil && ss.OutputDir != "" {
		return ss.OutputDir
	}
	return ".agents/skills"
}

// scaffoldTargetSkills creates the full skill directory structure for every skill in every
// platform-specific skills directory derived from the active platforms. Existing files are
// skipped so user edits are never overwritten.
func scaffoldTargetSkills(root string, skillNames []string, ss *models.SkillSourceConfig, platforms []string, force bool) (created, skipped int, err error) {
	if len(skillNames) == 0 {
		return 0, 0, nil
	}

	// Collect unique base directories across all platforms.
	dirSeen := map[string]struct{}{}
	var baseDirs []string
	for _, p := range platforms {
		d := canonicalPlatformSkillBaseDir(p)
		if _, dup := dirSeen[d]; !dup {
			dirSeen[d] = struct{}{}
			baseDirs = append(baseDirs, d)
		}
	}
	// Always include .agents/skills/ as the universal fallback.
	if _, ok := dirSeen[".agents/skills"]; !ok {
		baseDirs = append(baseDirs, ".agents/skills")
	}

	fakeSS := &models.SkillSourceConfig{}
	if ss != nil {
		fakeSS.Defaults = ss.Defaults
	}

	// Build description lookup from skill_sources.skills definitions so that
	// scaffolded skill.yaml gets the real description instead of the placeholder.
	descLookup := map[string]string{}
	if ss != nil {
		for _, def := range ss.Skills {
			if def.Description != "" {
				descLookup[def.Name] = def.Description
			}
		}
	}

	skillSeen := map[string]struct{}{}
	for _, name := range skillNames {
		if _, dup := skillSeen[name]; dup {
			continue
		}
		skillSeen[name] = struct{}{}

		for _, baseDir := range baseDirs {
			absDir := baseDir
			if !filepath.IsAbs(absDir) {
				absDir = filepath.Join(root, baseDir)
			}
			fakeSS.OutputDir = absDir
			opts := skillDefToScaffoldOpts(models.SkillSourceDefinition{Name: name, Description: descLookup[name]}, fakeSS, absDir, false, force)
			result, writeErr := scaffold.WriteSkillSafe(opts)
			if writeErr != nil {
				return created, skipped, fmt.Errorf("skill %s in %s: %w", name, baseDir, writeErr)
			}
			created += len(result.Created)
			skipped += len(result.Skipped)
		}
	}
	return created, skipped, nil
}

// resolveDefaultTarget picks the best target for commands that need a single resolved view
// of the bakefile (status, sync). Priority: "default" → "all" → sole target → error.
func resolveDefaultTarget(cfg *models.BakeConfig) (*models.TargetConfig, error) {
	for _, name := range []string{"default", "all"} {
		if _, ok := cfg.Targets[name]; ok {
			return cliResolveTarget(cfg, name)
		}
	}
	names := make([]string, 0, len(cfg.Targets))
	for n := range cfg.Targets {
		names = append(names, n)
	}
	if len(names) == 1 {
		return cliResolveTarget(cfg, names[0])
	}
	sort.Strings(names)
	return nil, fmt.Errorf("no default target; specify one of: %s", strings.Join(names, ", "))
}

func renderPlatformFilesEnabled(resolved *models.TargetConfig) bool {
	if resolved.Render != nil && resolved.Render.PlatformFiles != nil {
		return *resolved.Render.PlatformFiles
	}
	return true
}

func loadSkillRenderOptions(root string, cfg *models.BakeConfig, target *models.TargetConfig, from, mode string) (renderer.Options, []models.RenderedFile, error) {
	opts := renderer.Options{SkillsMode: models.SkillModeReference, Root: root}
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
