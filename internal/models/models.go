package models

import (
	"fmt"
	"strings"

	"github.com/domehahn/sklib/spec"
)

// Platform type aliases from sklib/spec for backward compatibility.
type Platform = spec.Platform

// SupportedPlatforms is the set of valid platform identifiers (including "all").
// Derived from sklib/spec so it stays in sync automatically.
var SupportedPlatforms = func() map[string]struct{} {
	m := map[string]struct{}{"all": {}}
	for _, p := range spec.ExpandAllPlatforms([]spec.Platform{spec.PlatformAll}) {
		m[string(p)] = struct{}{}
	}
	for _, p := range ExtendedAgentPlatforms {
		m[p] = struct{}{}
	}
	return m
}()

// PlatformAliases maps non-canonical platform names to canonical ones.
// These are skcr-specific aliases not covered by sklib/spec.
var PlatformAliases = map[string]string{
	"gitlab-duo-agent-platform": "gitlab-duo",
	"github-copilot-chat":       "github-copilot",
	"amazon":                    "amazon-q",
	"amazon q":                  "amazon-q",
	"amazon q developer":        "amazon-q",
	"amazon-q-developer":        "amazon-q",
	"auggie-cli":                "auggie",
	"cline-code":                "cline",
	"kilo-code":                 "kilocode",
	"kilo":                      "kilocode",
	"roo":                       "roo-code",
	"qwen-code":                 "qwen",
	"qoder-ide":                 "qoder",
}

var ExtendedAgentPlatforms = []string{
	"amazon-q",
	"antigravity",
	"auggie",
	"bob",
	"cline",
	"codebuddy",
	"continue",
	"costrict",
	"crush",
	"factory",
	"forgecode",
	"iflow",
	"kilocode",
	"kimi",
	"lingma",
	"pi",
	"qoder",
	"qwen",
}

// CanonicalPlatforms is the ordered list of all concrete canonical platforms followed by "all".
// Derived from sklib/spec so it stays in sync automatically.
var CanonicalPlatforms = append(append(
	spec.PlatformStrings(spec.ExpandAllPlatforms([]spec.Platform{spec.PlatformAll})),
	ExtendedAgentPlatforms...),
	"all")

func AllConcretePlatforms() []string {
	out := make([]string, 0, len(CanonicalPlatforms))
	seen := map[string]struct{}{}
	for _, platform := range CanonicalPlatforms {
		if platform == "all" {
			continue
		}
		if _, ok := seen[platform]; ok {
			continue
		}
		seen[platform] = struct{}{}
		out = append(out, platform)
	}
	return out
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
	Delivery    string         `yaml:"delivery,omitempty"`
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

// NormalizePlatform normalizes a platform string to its canonical form.
// sklib/spec handles the common aliases; PlatformAliases covers skcr-specific ones.
func NormalizePlatform(value string) (string, error) {
	p, err := spec.NormalizePlatform(value)
	if err != nil {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if alias, ok := PlatformAliases[normalized]; ok {
			return alias, nil
		}
		for _, platform := range ExtendedAgentPlatforms {
			if normalized == platform {
				return platform, nil
			}
		}
		return "", fmt.Errorf("unsupported platform: %s", value)
	}
	return string(p), nil
}

// NormalizePlatforms normalizes and deduplicates a slice of platform strings.
// "all" is expanded to every concrete canonical platform via sklib/spec.ExpandAllPlatforms.
func NormalizePlatforms(values []string) ([]string, error) {
	result := []string{}
	seen := map[string]struct{}{}
	for _, item := range values {
		p, err := NormalizePlatform(item)
		if err != nil {
			return nil, err
		}
		if p == "all" {
			for _, canonical := range spec.ExpandAllPlatforms([]spec.Platform{spec.PlatformAll}) {
				s := string(canonical)
				if _, ok := seen[s]; !ok {
					seen[s] = struct{}{}
					result = append(result, s)
				}
			}
			for _, s := range ExtendedAgentPlatforms {
				if _, ok := seen[s]; !ok {
					seen[s] = struct{}{}
					result = append(result, s)
				}
			}
			continue
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			result = append(result, p)
		}
	}
	return result, nil
}

// ParsePlatforms parses a comma-separated platform string.
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
