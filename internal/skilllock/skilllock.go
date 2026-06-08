package skilllock

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/domehahn/skcr/internal/models"
	"gopkg.in/yaml.v3"
)

type LockedSkill struct {
	Name           string
	Version        string
	Source         string
	Checksum       string
	CompatibleWith []string
	InstalledPaths []string
	Metadata       map[string]any
}

type LockState struct {
	Skills []LockedSkill
}

func Load(path string) (*LockState, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent-skills.lock not found at %s; run skpm lock or skpm install", path)
		}
		return nil, err
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("invalid agent-skills.lock at %s: %w; run skpm verify", path, err)
	}
	if raw == nil {
		return &LockState{}, nil
	}
	skills, err := parseSkills(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid agent-skills.lock at %s: %w; run skpm verify", path, err)
	}
	return &LockState{Skills: skills}, nil
}

func FilterByPlatforms(skills []LockedSkill, platforms []string) []LockedSkill {
	wanted := map[string]struct{}{}
	for _, platform := range platforms {
		normalized, err := models.NormalizePlatform(platform)
		if err != nil {
			continue
		}
		wanted[normalized] = struct{}{}
	}
	if len(wanted) == 0 {
		return skills
	}

	result := []LockedSkill{}
	for _, skill := range skills {
		if len(skill.CompatibleWith) == 0 {
			result = append(result, skill)
			continue
		}
		for _, platform := range skill.CompatibleWith {
			normalized, err := models.NormalizePlatform(platform)
			if err != nil {
				continue
			}
			if _, ok := wanted[normalized]; ok {
				result = append(result, skill)
				break
			}
		}
	}
	return result
}

func References(skills []LockedSkill) []map[string]any {
	items := make([]map[string]any, 0, len(skills))
	for _, skill := range skills {
		items = append(items, map[string]any{
			"name":            skill.Name,
			"version":         skill.Version,
			"source":          skill.Source,
			"checksum":        skill.Checksum,
			"compatible_with": skill.CompatibleWith,
			"installed_paths": skill.InstalledPaths,
			"path":            preferredSkillPath(skill),
		})
	}
	return items
}

func SkillFiles(root string, skills []LockedSkill, mode string, platforms []string) ([]models.RenderedFile, error) {
	if mode != models.SkillModeCopy && mode != models.SkillModeLink {
		return nil, nil
	}
	files := []models.RenderedFile{}
	for _, skill := range FilterByPlatforms(skills, platforms) {
		source := preferredSkillPath(skill)
		if source == "" {
			return nil, fmt.Errorf("locked skill %q has no installed path; run skpm install", skill.Name)
		}
		sourcePath := source
		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(root, sourcePath)
		}
		if _, err := os.Stat(sourcePath); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("locked skill %q path %s is missing; run skpm install", skill.Name, source)
			}
			return nil, err
		}
		content := ""
		if mode == models.SkillModeCopy {
			payload, err := os.ReadFile(sourcePath)
			if err != nil {
				return nil, err
			}
			content = string(payload)
		}
		for _, platform := range platforms {
			dest := PlatformSkillDestination(platform, skill.Name)
			if dest == "" {
				continue
			}
			rendered := models.RenderedFile{
				Source:      source,
				Destination: dest,
				Content:     content,
				Platform:    platform,
			}
			if mode == models.SkillModeLink {
				rendered.LinkTarget = source
			}
			files = append(files, rendered)
		}
	}
	return files, nil
}

func PlatformSkillDestination(platform, name string) string {
	switch platform {
	case "codex":
		return filepath.ToSlash(filepath.Join(".agents", "skills", name, "SKILL.md"))
	case "claude-code":
		return filepath.ToSlash(filepath.Join(".claude", "skills", name, "SKILL.md"))
	case "gitlab-duo":
		return filepath.ToSlash(filepath.Join("skills", name, "SKILL.md"))
	case "github-copilot":
		return filepath.ToSlash(filepath.Join(".github", "prompts", name+".prompt.md"))
	case "cursor", "windsurf", "generic":
		return filepath.ToSlash(filepath.Join(".agentic", "skills", name, "SKILL.md"))
	default:
		return ""
	}
}

func preferredSkillPath(skill LockedSkill) string {
	for _, p := range skill.InstalledPaths {
		if strings.HasSuffix(filepath.ToSlash(p), "/SKILL.md") || filepath.Base(p) == "SKILL.md" {
			return p
		}
	}
	if len(skill.InstalledPaths) == 0 {
		return ""
	}
	return filepath.ToSlash(filepath.Join(skill.InstalledPaths[0], "SKILL.md"))
}

func parseSkills(raw map[string]any) ([]LockedSkill, error) {
	value, ok := raw["skills"]
	if !ok {
		value = raw["locked_skills"]
	}
	if value == nil {
		return []LockedSkill{}, nil
	}
	switch typed := value.(type) {
	case []any:
		result := []LockedSkill{}
		for _, item := range typed {
			m, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("skills entries must be mappings")
			}
			skill, err := parseSkill(m, "")
			if err != nil {
				return nil, err
			}
			result = append(result, skill)
		}
		return result, nil
	case map[string]any:
		result := []LockedSkill{}
		for name, item := range typed {
			m, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("skill %q must be a mapping", name)
			}
			skill, err := parseSkill(m, name)
			if err != nil {
				return nil, err
			}
			result = append(result, skill)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("skills must be a list or mapping")
	}
}

func parseSkill(m map[string]any, fallbackName string) (LockedSkill, error) {
	name := stringValue(m, "name")
	if name == "" {
		name = fallbackName
	}
	if name == "" {
		return LockedSkill{}, fmt.Errorf("locked skill is missing name")
	}
	return LockedSkill{
		Name:           name,
		Version:        stringValue(m, "version"),
		Source:         stringValue(m, "source"),
		Checksum:       stringValue(m, "checksum"),
		CompatibleWith: normalizeStringList(firstValue(m, "compatible_with", "compatibleWith", "platforms")),
		InstalledPaths: normalizeStringList(firstValue(m, "installed_paths", "installedPaths", "installed_path", "path")),
		Metadata:       metadataValue(m),
	}, nil
}

func firstValue(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			return value
		}
	}
	return nil
}

func stringValue(m map[string]any, key string) string {
	value, _ := m[key].(string)
	return value
}

func normalizeStringList(value any) []string {
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return []string{}
		}
		return []string{typed}
	case []any:
		result := []string{}
		for _, item := range typed {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return typed
	default:
		return []string{}
	}
}

func metadataValue(m map[string]any) map[string]any {
	metadata, ok := m["metadata"].(map[string]any)
	if !ok || metadata == nil {
		return map[string]any{}
	}
	return metadata
}
