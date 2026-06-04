package models

import (
	"fmt"
	"strings"
)

var SupportedPlatforms = map[string]struct{}{
	"codex":          {},
	"gitlab-duo":     {},
	"github-copilot": {},
	"claude":         {},
	"openhands":      {},
	"opencode":       {},
	"ollama":         {},
	"generic":        {},
}

var PlatformAliases = map[string]string{
	"gitlab":                    "gitlab-duo",
	"duo":                       "gitlab-duo",
	"gitlab-duo-agent-platform": "gitlab-duo",
	"copilot":                   "github-copilot",
	"github":                    "github-copilot",
	"github-copilot-chat":       "github-copilot",
	"claude-code":               "claude",
	"open-code":                 "opencode",
	"open-hands":                "openhands",
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
}

type BakeConfig struct {
	Version   string                   `yaml:"version,omitempty"`
	Variables map[string]any           `yaml:"variables,omitempty"`
	Targets   map[string]*TargetConfig `yaml:"targets,omitempty"`
}

type RenderedFile struct {
	Source      string
	Destination string
	Content     string
	Platform    string
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

func ParsePlatforms(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}, nil
	}
	result := []string{}
	seen := map[string]struct{}{}
	for _, item := range strings.Split(value, ",") {
		p, err := NormalizePlatform(item)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		result = append(result, p)
	}
	return result, nil
}
