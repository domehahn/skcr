package models

import (
	"fmt"
	"strings"
)

var SupportedPlatforms = map[string]struct{}{
	"codex":          {},
	"gitlab-duo":     {},
	"github-copilot": {},
	"claude-code":    {},
	"cursor":         {},
	"windsurf":       {},
	"openhands":      {},
	"opencode":       {},
	"ollama":         {},
	"generic":        {},
	"all":            {},
}

var PlatformAliases = map[string]string{
	"gitlab":                    "gitlab-duo",
	"duo":                       "gitlab-duo",
	"gitlab-duo-agent-platform": "gitlab-duo",
	"copilot":                   "github-copilot",
	"github":                    "github-copilot",
	"github-copilot-chat":       "github-copilot",
	"claude":                    "claude-code",
	"open-code":                 "opencode",
	"open-hands":                "openhands",
}

var CanonicalPlatforms = []string{
	"codex",
	"claude-code",
	"gitlab-duo",
	"github-copilot",
	"cursor",
	"windsurf",
	"generic",
	"all",
}

const (
	SkillModeReference = "reference"
	SkillModeCopy      = "copy"
	SkillModeLink      = "link"
	SkillModeEmbed     = "embed"
)

type SkillIntegrationConfig struct {
	Source    string   `yaml:"source,omitempty"`
	Mode      string   `yaml:"mode,omitempty"`
	Platforms []string `yaml:"platforms,omitempty"`
}

type SkillSourceDefaults struct {
	Version        string   `yaml:"version,omitempty"`
	Owner          string   `yaml:"owner,omitempty"`
	License        string   `yaml:"license,omitempty"`
	CompatibleWith []string `yaml:"compatible_with,omitempty"`
	Template       string   `yaml:"template,omitempty"`
}

type SkillSourceDefinition struct {
	Name           string   `yaml:"name"`
	Version        string   `yaml:"version,omitempty"`
	Description    string   `yaml:"description,omitempty"`
	Owner          string   `yaml:"owner,omitempty"`
	License        string   `yaml:"license,omitempty"`
	CompatibleWith []string `yaml:"compatible_with,omitempty"`
	Template       string   `yaml:"template,omitempty"`
}

type SkillSourceConfig struct {
	OutputDir string                  `yaml:"output_dir,omitempty"`
	Defaults  SkillSourceDefaults     `yaml:"defaults,omitempty"`
	Skills    []SkillSourceDefinition `yaml:"skills,omitempty"`
}

type RenderConfig struct {
	SkillSources   *bool `yaml:"skill_sources,omitempty"`
	PlatformFiles  *bool `yaml:"platform_files,omitempty"`
	PlatformSkills *bool `yaml:"platform_skills,omitempty"`
}

type TargetConfig struct {
	Description string         `yaml:"description,omitempty"`
	Inherits    []string       `yaml:"inherits,omitempty"`
	Platforms   []string       `yaml:"platforms,omitempty"`
	Profiles    []string       `yaml:"profiles,omitempty"`
	Skills      []string       `yaml:"skills,omitempty"`
	Flows       []string       `yaml:"flows,omitempty"`
	Rules       map[string]any `yaml:"rules,omitempty"`
	Model       map[string]any `yaml:"model,omitempty"`
	GitLabDuo   map[string]any `yaml:"gitlab_duo,omitempty"`
	Render      *RenderConfig  `yaml:"render,omitempty"`
}

type BakeConfig struct {
	Version      string                   `yaml:"version,omitempty"`
	Variables    map[string]any           `yaml:"variables,omitempty"`
	SkillSources *SkillSourceConfig       `yaml:"skill_sources,omitempty"`
	Skills       *SkillIntegrationConfig  `yaml:"skills,omitempty"`
	Targets      map[string]*TargetConfig `yaml:"targets,omitempty"`
}

type RenderedFile struct {
	Source      string
	Destination string
	Content     string
	Platform    string
	LinkTarget  string
}

func NormalizePlatform(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if alias, ok := PlatformAliases[normalized]; ok {
		normalized = alias
	}
	if _, ok := SupportedPlatforms[normalized]; !ok {
		return "", fmt.Errorf("unsupported platform: %s", value)
	}
	return normalized, nil
}

func NormalizePlatforms(values []string) ([]string, error) {
	result := []string{}
	seen := map[string]struct{}{}
	for _, item := range values {
		p, err := NormalizePlatform(item)
		if err != nil {
			return nil, err
		}
		if p == "all" {
			for _, canonical := range CanonicalPlatforms {
				if canonical == "all" {
					continue
				}
				if _, ok := seen[canonical]; ok {
					continue
				}
				seen[canonical] = struct{}{}
				result = append(result, canonical)
			}
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		result = append(result, p)
	}
	return result, nil
}

func ParsePlatforms(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}, nil
	}
	items := []string{}
	for _, item := range strings.Split(value, ",") {
		items = append(items, item)
	}
	return NormalizePlatforms(items)
}
