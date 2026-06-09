package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/domehahn/sklib/spec"
	"github.com/domehahn/skcr/internal/bake"
	"github.com/domehahn/skcr/internal/lockfile"
	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/renderer"
	"github.com/domehahn/skcr/internal/skilllock"
	"gopkg.in/yaml.v3"
)


type Options struct {
	AgainstLock string
	Skills      bool
	Platform    string
	CI          bool
}

func ValidateProject(target string) ([]string, error) {
	return ValidateProjectWithOptions(target, Options{})
}

func ValidateProjectWithOptions(target string, opts Options) ([]string, error) {
	errors := []string{}

	bakePath := filepath.Join(target, "agentic.bake.yaml")
	if _, err := os.Stat(bakePath); err != nil {
		if os.IsNotExist(err) {
			return []string{"Missing agentic.bake.yaml"}, nil
		}
		return nil, err
	}

	payload, err := os.ReadFile(bakePath)
	if err != nil {
		return nil, err
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}

	targets, _ := raw["targets"].(map[string]any)
	if len(targets) == 0 {
		errors = append(errors, "No targets configured in agentic.bake.yaml")
	}

	for name, cfgRaw := range targets {
		cfg, ok := cfgRaw.(map[string]any)
		if !ok {
			continue
		}
		platforms, _ := cfg["platforms"].([]any)
		for _, p := range platforms {
			platform, _ := p.(string)
			if _, err := models.NormalizePlatform(platform); err != nil {
				errors = append(errors, fmt.Sprintf("Target %s: unsupported platform %s", name, platform))
			}
		}
	}

	cfg, cfgErr := bake.LoadBakeFile(bakePath)
	if cfgErr != nil {
		cfg = nil
	}

	if cfg != nil && cfg.SkillSources != nil {
		errors = append(errors, validateSkillSources(target, cfg.SkillSources)...)
	}

	for _, baseDir := range []string{"skills", ".agents/skills", ".claude/skills", ".agentic/skills"} {
		skillsDir := filepath.Join(target, baseDir)
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				skillFile := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
				text, err := os.ReadFile(skillFile)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Skill missing SKILL.md: %s", filepath.Dir(skillFile)))
					continue
				}
				if errMsg := validateSkillMetadata(string(text)); errMsg != "" {
					errors = append(errors, fmt.Sprintf("%s: %s", errMsg, skillFile))
				}
			}
		}
	}

	gitlabDuo := filepath.Join(target, ".gitlab", "duo")
	if stat, err := os.Stat(gitlabDuo); err == nil && stat.IsDir() {
		chatRules := filepath.Join(gitlabDuo, "chat-rules.md")
		if _, err := os.Stat(chatRules); err != nil {
			errors = append(errors, "GitLab Duo output missing .gitlab/duo/chat-rules.md")
		}

		flowDir := filepath.Join(gitlabDuo, "flows")
		if entries, err := os.ReadDir(flowDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
					continue
				}
				flowPath := filepath.Join(flowDir, entry.Name())
				text, err := os.ReadFile(flowPath)
				if err != nil {
					continue
				}
				s := string(text)
				for _, forbidden := range []string{"name:", "description:", "product_group:"} {
					if len(s) >= len(forbidden) && s[:len(forbidden)] == forbidden {
						errors = append(errors, fmt.Sprintf("GitLab custom flow contains forbidden top-level field %s: %s", forbidden, flowPath))
					}
				}
				if !containsAll(s, []string{"workspace_agent_skills"}) {
					errors = append(errors, fmt.Sprintf("Flow does not pass workspace_agent_skills: %s", flowPath))
				}
				if !containsAll(s, []string{"user_rule"}) {
					errors = append(errors, fmt.Sprintf("Flow does not pass user_rule: %s", flowPath))
				}
			}
		}
	}

	if cfg != nil {
		targetName := "default"
		if _, ok := cfg.Targets[targetName]; !ok {
			for name := range cfg.Targets {
				targetName = name
				break
			}
		}
		if targetName != "" {
			resolved, err := bake.ResolveTarget(cfg, targetName)
			if err != nil {
				errors = append(errors, err.Error())
			} else {
				if opts.Platform != "" {
					platforms, err := models.ParsePlatforms(opts.Platform)
					if err != nil {
						return nil, err
					}
					resolved.Platforms = filterPlatforms(resolved.Platforms, platforms)
				}
				renderOpts := renderer.Options{}
				if opts.AgainstLock != "" || opts.Skills {
					source := opts.AgainstLock
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
					state, err := skilllock.Load(sourcePath)
					if err != nil {
						errors = append(errors, err.Error())
					} else {
						filtered := skilllock.FilterByPlatforms(state.Skills, resolved.Platforms)
						renderOpts.LockedSkills = skilllock.References(filtered)
						if cfg.Skills != nil {
							renderOpts.SkillsMode = cfg.Skills.Mode
						}
					}
				}
				files, err := renderer.RenderFilesWithOptions(cfg, resolved, renderOpts)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Template rendering failed: %v", err))
				} else {
					errors = append(errors, validateGeneratedState(target, files)...)
				}
				if opts.AgainstLock != "" || opts.Skills {
					source := opts.AgainstLock
					if source == "" && cfg.Skills != nil {
						source = cfg.Skills.Source
					}
					if source == "" {
						source = "agent-skills.lock"
					}
					errors = append(errors, validateSkillLock(target, source, resolved.Platforms)...)
				}
			}
		}
	}

	return errors, nil
}

func validateSkillSources(target string, ss *models.SkillSourceConfig) []string {
	errors := []string{}
	if ss.OutputDir == "" {
		ss.OutputDir = ".agents/skills"
	}
	seen := map[string]struct{}{}
	for _, skillDef := range ss.Skills {
		if err := spec.ValidateSkillName(skillDef.Name); err != nil {
			errors = append(errors, fmt.Sprintf("skill_sources: invalid skill name %q (use lowercase letters, digits, and hyphens)", skillDef.Name))
		}
		if _, dup := seen[skillDef.Name]; dup {
			errors = append(errors, fmt.Sprintf("skill_sources: duplicate skill name %q", skillDef.Name))
		}
		seen[skillDef.Name] = struct{}{}
		for _, p := range skillDef.CompatibleWith {
			if _, err := models.NormalizePlatform(p); err != nil {
				errors = append(errors, fmt.Sprintf("skill_sources: skill %q has unsupported platform %q", skillDef.Name, p))
			}
		}
	}
	for _, p := range ss.Defaults.CompatibleWith {
		if _, err := models.NormalizePlatform(p); err != nil {
			errors = append(errors, fmt.Sprintf("skill_sources.defaults: unsupported platform %q", p))
		}
	}
	if len(errors) > 0 {
		return errors
	}
	// Check that skill source directories exist for configured skills.
	for _, skillDef := range ss.Skills {
		skillDir := filepath.Join(target, ss.OutputDir, skillDef.Name)
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("skill_sources: skill directory missing: %s (run: skcr scaffold skills)", filepath.Join(ss.OutputDir, skillDef.Name)))
		}
	}
	return errors
}

func validateGeneratedState(target string, files []models.RenderedFile) []string {
	errors := []string{}
	lock, err := lockfile.LoadLockfile(target)
	if err != nil {
		return append(errors, err.Error())
	}
	stateFiles := lockfile.ManagedFilesByPath(lock)
	expected := map[string]models.RenderedFile{}
	for _, file := range files {
		expected[file.Destination] = file
	}
	for path, file := range expected {
		fullPath := filepath.Join(target, path)
		payload, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("Generated file missing: %s", path))
				continue
			}
			errors = append(errors, err.Error())
			continue
		}
		if file.LinkTarget == "" && lockfile.Sha256Text(string(payload)) != lockfile.Sha256Text(file.Content) {
			errors = append(errors, fmt.Sprintf("Generated file checksum mismatch: %s", path))
		}
		if entry, ok := stateFiles[path]; ok {
			if checksum, _ := entry["checksum"].(string); checksum != renderedChecksum(file) {
				errors = append(errors, fmt.Sprintf("Lockfile checksum mismatch: %s", path))
			}
		}
	}
	for path := range stateFiles {
		if _, ok := expected[path]; !ok {
			errors = append(errors, fmt.Sprintf("Stale generated file remains in lock state: %s", path))
		}
		if _, err := os.Stat(filepath.Join(target, path)); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("Generated file from lock is missing: %s", path))
		}
	}
	return errors
}

func renderedChecksum(file models.RenderedFile) string {
	if file.LinkTarget != "" {
		return lockfile.Sha256Text("link:" + file.LinkTarget)
	}
	return lockfile.Sha256Text(file.Content)
}

func validateSkillLock(target, source string, platforms []string) []string {
	errors := []string{}
	sourcePath := source
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(target, sourcePath)
	}
	state, err := skilllock.Load(sourcePath)
	if err != nil {
		return append(errors, err.Error())
	}
	compatible := skilllock.FilterByPlatforms(state.Skills, platforms)
	for _, skill := range compatible {
		if len(skill.InstalledPaths) == 0 {
			errors = append(errors, fmt.Sprintf("Locked skill %s has no install path; run skpm install", skill.Name))
			continue
		}
		for _, p := range skill.InstalledPaths {
			checkPath := p
			if !filepath.IsAbs(checkPath) {
				checkPath = filepath.Join(target, checkPath)
			}
			stat, err := os.Stat(checkPath)
			if err != nil {
				if os.IsNotExist(err) {
					errors = append(errors, fmt.Sprintf("Locked skill path missing: %s; run skpm install", p))
					continue
				}
				errors = append(errors, err.Error())
				continue
			}
			skillFile := checkPath
			if stat.IsDir() {
				skillFile = filepath.Join(checkPath, "SKILL.md")
			}
			if _, err := os.Stat(skillFile); err != nil {
				if os.IsNotExist(err) {
					errors = append(errors, fmt.Sprintf("Locked skill SKILL.md missing: %s; run skpm install", skillFile))
					continue
				}
				errors = append(errors, err.Error())
			}
		}
		for _, platform := range skill.CompatibleWith {
			if _, err := models.NormalizePlatform(platform); err != nil {
				errors = append(errors, fmt.Sprintf("Locked skill %s has unsupported platform %s; run skpm verify", skill.Name, platform))
			}
		}
	}
	return errors
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

func containsAll(s string, terms []string) bool {
	for _, t := range terms {
		if !strings.Contains(s, t) {
			return false
		}
	}
	return true
}

var (
	nameRegex        = regexp.MustCompile(`(?m)^name:\s*(.+?)\s*$`)
	descriptionRegex = regexp.MustCompile(`(?m)^description:\s*(.+?)\s*$`)
)

func validateSkillMetadata(content string) string {
	nameMatch := nameRegex.FindStringSubmatch(content)
	if len(nameMatch) < 2 || isEmptyMetadataValue(nameMatch[1]) {
		return "Skill metadata name is missing or empty"
	}

	descriptionMatch := descriptionRegex.FindStringSubmatch(content)
	if len(descriptionMatch) < 2 || isEmptyMetadataValue(descriptionMatch[1]) {
		return "Skill metadata description is missing or empty"
	}

	return ""
}

func isEmptyMetadataValue(value string) bool {
	v := strings.TrimSpace(value)
	v = strings.Trim(v, `"'`)
	return strings.TrimSpace(v) == ""
}
