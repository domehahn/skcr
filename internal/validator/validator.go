package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/domehahn/skcr/internal/bake"
	"github.com/domehahn/skcr/internal/lockfile"
	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/renderer"
	"github.com/domehahn/skcr/internal/skilllock"
	"github.com/domehahn/sklib/spec"
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

	for _, baseDir := range skillMetadataBaseDirs(cfg) {
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
				if errMsg := validateSkillMetadataForName(string(text), entry.Name()); errMsg != "" {
					errors = append(errors, fmt.Sprintf("%s: %s", errMsg, skillFile))
				}
			}
		}
	}
	errors = append(errors, validateCanonicalSkillSpecificBlocks(target, cfg)...)

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

func validateCanonicalSkillSpecificBlocks(target string, cfg *models.BakeConfig) []string {
	sourceDir := ".agents/skills"
	if cfg != nil && cfg.SkillSources != nil && cfg.SkillSources.OutputDir != "" {
		sourceDir = cfg.SkillSources.OutputDir
	}
	skillsDir := filepath.Join(target, sourceDir)
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}
	seen := map[string]string{}
	errors := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		payload, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		_, body, ok := splitSkillFrontmatter(string(payload))
		if !ok {
			continue
		}
		block := normalizedSkillSpecificBlock(body)
		if block == "" {
			continue
		}
		if previous, ok := seen[block]; ok {
			errors = append(errors, fmt.Sprintf("Skills %s and %s have identical skill-specific content blocks", previous, entry.Name()))
			continue
		}
		seen[block] = entry.Name()
	}
	return errors
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
		if isSkillMarkdownPath(path) {
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

func isSkillMarkdownPath(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) < 3 || parts[len(parts)-1] != "SKILL.md" {
		return false
	}
	for _, part := range parts[:len(parts)-2] {
		if part == "skills" {
			return true
		}
	}
	return false
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
	bodyChangelogHeadingRE = regexp.MustCompile(`(?m)^## Changelog\s*$`)
	bodyChangelogEntryRE   = regexp.MustCompile(`(?m)^###\s+([0-9A-Za-z.+-]+)\s+-\s+(\d{4}-\d{2}-\d{2})\s*$`)
	skillReadinessHeadings = []string{
		"## Purpose",
		"## When to use",
		"## Operating model",
		"## Spec-Driven Change Context",
		"## Skill-Specific Review Scope",
		"## Skill-Specific Checklist",
		"## Decision Rules",
		"## Finding Categories",
		"## Severity Guidance",
		"## DevSecOps Guardrails",
		"## Acceptance Criteria",
		"## Output Requirements",
		"## Anti-Patterns",
		"## Changelog",
	}
	skillSpecificBlockRE = regexp.MustCompile(`(?s)## Skill-Specific Review Scope\s*(.*?)\n## Changelog\s*`)
)

func validateSkillMetadata(content string) string {
	return validateSkillMetadataForName(content, "")
}

func ValidateSkillMetadata(content string) string {
	return validateSkillMetadata(content)
}

func validateSkillMetadataForName(content, expectedName string) string {
	frontmatter, body, ok := splitSkillFrontmatter(content)
	if !ok {
		return "Skill metadata frontmatter is missing"
	}

	raw := map[string]any{}
	if err := yaml.Unmarshal([]byte(frontmatter), &raw); err != nil {
		return "Skill metadata frontmatter is invalid YAML"
	}

	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return "Skill metadata frontmatter is invalid YAML"
	}

	errs := []string{}
	for _, field := range []string{
		"name",
		"description",
		"version",
		"since",
		"last_modified",
		"authors",
		"stability",
		"min_platform_version",
		"deprecated_since",
		"replaces",
		"supersedes",
		"changelog",
	} {
		if _, ok := raw[field]; !ok {
			errs = append(errs, "missing required field: "+field)
		}
	}
	if isEmptyMetadataValue(fm.Description) {
		errs = append(errs, "missing required field: description")
	}
	if expectedName != "" && fm.Name != "" && fm.Name != expectedName {
		errs = append(errs, "name does not match skill directory: "+fm.Name+" != "+expectedName)
	}
	errs = append(errs, spec.ValidateSkillMDFrontmatter(fm)...)
	if !bodyChangelogHeadingRE.MatchString(body) {
		errs = append(errs, "missing required body section: ## Changelog")
	} else {
		bodyEntries := bodyChangelogEntryRE.FindAllStringSubmatch(body, -1)
		if len(bodyEntries) == 0 {
			errs = append(errs, "body Changelog has no version entries")
		} else if fm.Version != "" && bodyEntries[0][1] != fm.Version {
			errs = append(errs, "version mismatch: frontmatter version "+fm.Version+" does not match newest body changelog entry "+bodyEntries[0][1])
		}
	}
	for _, heading := range skillReadinessHeadings {
		if !hasMarkdownHeading(body, heading) {
			errs = append(errs, "missing required body section: "+heading)
		}
	}
	if len(errs) == 0 {
		return ""
	}
	return "Skill metadata invalid: " + strings.Join(dedupeStrings(errs), "; ")
}

func ValidateSkillWarnings(content string) []string {
	frontmatter, _, ok := splitSkillFrontmatter(content)
	if !ok {
		return nil
	}
	fm := spec.SkillMDFrontmatter{}
	if err := yaml.Unmarshal([]byte(frontmatter), &fm); err != nil {
		return nil
	}
	warnings := []string{}
	for platform, version := range fm.MinPlatformVer {
		if strings.EqualFold(strings.TrimSpace(version), "unknown") {
			warnings = append(warnings, "min_platform_version for "+platform+" is unknown; production routing must treat this as unverified compatibility")
		}
	}
	return dedupeStrings(warnings)
}

func hasMarkdownHeading(body, heading string) bool {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(heading) + `\s*$`)
	return re.MatchString(body)
}

func normalizedSkillSpecificBlock(body string) string {
	match := skillSpecificBlockRE.FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}
	fields := strings.Fields(strings.ToLower(match[1]))
	return strings.Join(fields, " ")
}

func isEmptyMetadataValue(value string) bool {
	v := strings.TrimSpace(value)
	v = strings.Trim(v, `"'`)
	return strings.TrimSpace(v) == ""
}

func splitSkillFrontmatter(content string) (frontmatter, body string, ok bool) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content, false
	}
	rest := strings.TrimPrefix(content, "---\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", content, false
	}
	frontmatter = rest[:end]
	bodyStart := end + len("\n---")
	if len(rest) > bodyStart && rest[bodyStart] == '\r' {
		bodyStart++
	}
	if len(rest) > bodyStart && rest[bodyStart] == '\n' {
		bodyStart++
	}
	return frontmatter, rest[bodyStart:], true
}

func skillMetadataBaseDirs(cfg *models.BakeConfig) []string {
	baseDirs := []string{
		".agents/skills",
		"skills",
		".claude/skills",
		".github/skills",
		".opencode/skills",
		".openhands/skills",
		".ollama/skills",
		".cursor/skills",
		".roo/skills",
		".kiro/skills",
		".junie/skills",
		".gemini/skills",
		".windsurf/skills",
		".agentic/skills",
	}
	if cfg != nil {
		for _, target := range cfg.Targets {
			for _, platform := range target.Platforms {
				if dir := platformSkillBaseDir(platform); dir != "" {
					baseDirs = append(baseDirs, dir)
				}
			}
		}
		if cfg.SkillSources != nil && cfg.SkillSources.OutputDir != "" {
			baseDirs = append(baseDirs, cfg.SkillSources.OutputDir)
		}
	}
	slices.Sort(baseDirs)
	return slices.Compact(baseDirs)
}

func platformSkillBaseDir(platform string) string {
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

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := []string{}
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
