package bake

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/agentic-template-kit/skcr/internal/catalog"
	"github.com/agentic-template-kit/skcr/internal/models"
	"gopkg.in/yaml.v3"
)

func mergeUnique(base, add []string) []string {
	result := slices.Clone(base)
	for _, item := range add {
		if !slices.Contains(result, item) {
			result = append(result, item)
		}
	}
	return result
}

func deepMerge(base, add map[string]any) map[string]any {
	result := map[string]any{}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range add {
		left, lok := result[k].(map[string]any)
		right, rok := v.(map[string]any)
		if lok && rok {
			result[k] = deepMerge(left, right)
			continue
		}
		result[k] = v
	}
	return result
}

func LoadBakeFile(path string) (*models.BakeConfig, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &models.BakeConfig{}
	if err := yaml.Unmarshal(payload, cfg); err != nil {
		return nil, err
	}
	if cfg.Version == "" {
		cfg.Version = "1"
	}
	if cfg.Variables == nil {
		cfg.Variables = map[string]any{}
	}
	if cfg.Targets == nil {
		cfg.Targets = map[string]*models.TargetConfig{}
	}
	for _, target := range cfg.Targets {
		normalizeTarget(target)
	}
	return cfg, nil
}

func DumpBakeFile(config *models.BakeConfig, path string) error {
	payload, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func normalizeTarget(target *models.TargetConfig) {
	if target == nil {
		return
	}
	if target.Inherits == nil {
		target.Inherits = []string{}
	}
	if target.Platforms == nil {
		target.Platforms = []string{}
	}
	if target.Profiles == nil {
		target.Profiles = []string{}
	}
	if target.Skills == nil {
		target.Skills = []string{}
	}
	if target.Flows == nil {
		target.Flows = []string{}
	}
	if target.Rules == nil {
		target.Rules = map[string]any{}
	}
	if target.Model == nil {
		target.Model = map[string]any{}
	}
}

func ResolveTarget(config *models.BakeConfig, targetName string) (*models.TargetConfig, error) {
	if _, ok := config.Targets[targetName]; !ok {
		return nil, fmt.Errorf("unknown target %q", targetName)
	}

	activeStack := []string{}

	var resolve func(name string) (*models.TargetConfig, error)
	resolve = func(name string) (*models.TargetConfig, error) {
		if slices.Contains(activeStack, name) {
			return nil, fmt.Errorf("circular target inheritance detected: %v -> %s", activeStack, name)
		}
		target, ok := config.Targets[name]
		if !ok {
			return nil, fmt.Errorf("target %q not found", name)
		}
		normalizeTarget(target)
		activeStack = append(activeStack, name)
		merged := &models.TargetConfig{}
		normalizeTarget(merged)
		merged.Description = target.Description

		for _, parent := range target.Inherits {
			parentTarget, err := resolve(parent)
			if err != nil {
				return nil, err
			}
			merged.Platforms = mergeUnique(merged.Platforms, parentTarget.Platforms)
			merged.Profiles = mergeUnique(merged.Profiles, parentTarget.Profiles)
			merged.Skills = mergeUnique(merged.Skills, parentTarget.Skills)
			merged.Flows = mergeUnique(merged.Flows, parentTarget.Flows)
			merged.Rules = deepMerge(merged.Rules, parentTarget.Rules)
			merged.Model = deepMerge(merged.Model, parentTarget.Model)
		}

		merged.Platforms = mergeUnique(merged.Platforms, target.Platforms)
		merged.Profiles = mergeUnique(merged.Profiles, target.Profiles)
		merged.Skills = mergeUnique(merged.Skills, target.Skills)
		merged.Flows = mergeUnique(merged.Flows, target.Flows)
		merged.Rules = deepMerge(merged.Rules, target.Rules)
		merged.Model = deepMerge(merged.Model, target.Model)

		activeStack = activeStack[:len(activeStack)-1]
		return merged, nil
	}

	return resolve(targetName)
}

func BuildInitialConfig(
	platforms []string,
	projectName,
	ownerTeam,
	language,
	governanceLevel string,
	preset string,
) (*models.BakeConfig, error) {
	if preset != "" {
		switch preset {
		case "minimal":
			platforms = []string{"codex"}
		case "gitlab":
			platforms = []string{"gitlab-duo"}
		case "enterprise":
			platforms = []string{"gitlab-duo", "codex", "github-copilot"}
		case "local-ai":
			platforms = []string{"opencode", "openhands", "ollama"}
		case "all":
			platforms = []string{"codex", "github-copilot", "claude", "gitlab-duo", "opencode", "openhands", "ollama"}
		default:
			return nil, fmt.Errorf("unsupported preset: %s", preset)
		}
	}

	if len(platforms) == 0 {
		platforms = []string{"codex", "github-copilot", "claude", "gitlab-duo", "opencode", "openhands", "ollama"}
	}

	variables := map[string]any{
		"project_name":     projectName,
		"owner_team":       ownerTeam,
		"default_language": language,
		"governance_level": governanceLevel,
	}

	targets := map[string]*models.TargetConfig{}

	if slices.Contains(platforms, "codex") {
		targets["codex"] = &models.TargetConfig{
			Description: "Codex AGENTS.md and project skills",
			Platforms:   []string{"codex"},
			Profiles:    []string{"base", "devsecops"},
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
		}
	}

	if slices.Contains(platforms, "github-copilot") {
		targets["copilot"] = &models.TargetConfig{
			Description: "GitHub Copilot repository instructions and prompt files",
			Platforms:   []string{"github-copilot"},
			Profiles:    []string{"base", "devsecops", "documentation"},
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
		}
	}

	if slices.Contains(platforms, "claude") {
		targets["claude"] = &models.TargetConfig{
			Description: "Claude Code CLAUDE.md and project skills",
			Platforms:   []string{"claude"},
			Profiles:    []string{"base", "devsecops"},
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
		}
	}

	if slices.Contains(platforms, "gitlab-duo") {
		targets["gitlab"] = &models.TargetConfig{
			Description: "GitLab Duo Agent Platform setup with AGENTS.md, project-level skills, custom rules, and flow templates",
			Platforms:   []string{"gitlab-duo"},
			Profiles:    []string{"base", "gitlab-governance", "devsecops", "documentation"},
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
			Flows:       slices.Clone(catalog.DevsecopsFlows),
		}
	}

	localPlatforms := []string{}
	for _, p := range []string{"opencode", "openhands", "ollama"} {
		if slices.Contains(platforms, p) {
			localPlatforms = append(localPlatforms, p)
		}
	}
	if len(localPlatforms) > 0 {
		targets["local-ai"] = &models.TargetConfig{
			Description: "Local Ollama/OpenCode/OpenHands setup",
			Platforms:   localPlatforms,
			Profiles:    []string{"base", "local-models"},
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
			Model: map[string]any{
				"provider":      "ollama",
				"default_model": "qwen2.5-coder:7b",
				"base_url":      "http://localhost:11434",
			},
		}
	}

	if len(targets) == 0 {
		return nil, errors.New("no targets generated from selected platforms")
	}

	targetNames := make([]string, 0, len(targets))
	for name := range targets {
		targetNames = append(targetNames, name)
	}
	slices.Sort(targetNames)

	targets["default"] = &models.TargetConfig{
		Description: "Default target for this repository",
		Inherits:    slices.Clone(targetNames),
	}
	targets["all"] = &models.TargetConfig{
		Description: "Generate all configured platform artifacts",
		Inherits:    slices.Clone(targetNames),
	}

	return &models.BakeConfig{
		Version:   "1",
		Variables: variables,
		Targets:   targets,
	}, nil
}
