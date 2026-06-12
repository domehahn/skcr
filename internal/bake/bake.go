package bake

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/domehahn/skcr/internal/catalog"
	"github.com/domehahn/skcr/internal/models"
	"gopkg.in/yaml.v3"
)

var bakeYAMLMarshal = yaml.Marshal

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
	normalizeSkillConfig(cfg)
	normalizeSkillSourceConfig(cfg)
	if cfg.Targets == nil {
		cfg.Targets = map[string]*models.TargetConfig{}
	}
	for _, target := range cfg.Targets {
		if err := normalizeTarget(target); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func DumpBakeFile(config *models.BakeConfig, path string) error {
	payload, err := bakeYAMLMarshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func normalizeSkillConfig(cfg *models.BakeConfig) {
	if cfg.Skills == nil {
		cfg.Skills = &models.SkillIntegrationConfig{}
	}
	if cfg.Skills.Source == "" {
		cfg.Skills.Source = "agent-skills.lock"
	}
	if cfg.Skills.Mode == "" {
		cfg.Skills.Mode = models.SkillModeReference
	}
	if cfg.Skills.Platforms == nil {
		cfg.Skills.Platforms = []string{}
	} else {
		platforms, err := models.NormalizePlatforms(cfg.Skills.Platforms)
		if err == nil {
			cfg.Skills.Platforms = platforms
		}
	}
}

func normalizeSkillSourceConfig(cfg *models.BakeConfig) {
	if cfg.SkillSources == nil {
		return
	}
	ss := cfg.SkillSources
	if ss.OutputDir == "" {
		ss.OutputDir = ".agents/skills"
	}
	if ss.Defaults.Version == "" {
		ss.Defaults.Version = "0.1.0"
	}
	if ss.Defaults.License == "" {
		ss.Defaults.License = "MIT"
	}
	if len(ss.Defaults.CompatibleWith) > 0 {
		if normalized, err := models.NormalizePlatforms(ss.Defaults.CompatibleWith); err == nil {
			ss.Defaults.CompatibleWith = normalized
		}
	}
	for i := range ss.Skills {
		if len(ss.Skills[i].CompatibleWith) > 0 {
			if normalized, err := models.NormalizePlatforms(ss.Skills[i].CompatibleWith); err == nil {
				ss.Skills[i].CompatibleWith = normalized
			}
		}
	}
}

func normalizeTarget(target *models.TargetConfig) error {
	if target == nil {
		return nil
	}
	if target.Inherits == nil {
		target.Inherits = []string{}
	}
	if target.Platforms == nil {
		target.Platforms = []string{}
	} else {
		platforms, err := models.NormalizePlatforms(target.Platforms)
		if err != nil {
			return err
		}
		target.Platforms = platforms
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
	if target.GitLabDuo == nil {
		target.GitLabDuo = map[string]any{}
	}
	switch target.Delivery {
	case "", "skills", "commands", "both":
	default:
		return fmt.Errorf("unsupported delivery mode: %s", target.Delivery)
	}
	return nil
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
		if err := normalizeTarget(target); err != nil {
			return nil, err
		}
		activeStack = append(activeStack, name)
		merged := &models.TargetConfig{}
		if err := normalizeTarget(merged); err != nil {
			return nil, err
		}
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
			merged.GitLabDuo = deepMerge(merged.GitLabDuo, parentTarget.GitLabDuo)
			if merged.Delivery == "" {
				merged.Delivery = parentTarget.Delivery
			}
		}

		merged.Platforms = mergeUnique(merged.Platforms, target.Platforms)
		merged.Profiles = mergeUnique(merged.Profiles, target.Profiles)
		merged.Skills = mergeUnique(merged.Skills, target.Skills)
		merged.Flows = mergeUnique(merged.Flows, target.Flows)
		merged.Rules = deepMerge(merged.Rules, target.Rules)
		merged.Model = deepMerge(merged.Model, target.Model)
		merged.GitLabDuo = deepMerge(merged.GitLabDuo, target.GitLabDuo)
		if target.Delivery != "" {
			merged.Delivery = target.Delivery
		}

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
	defaultRuntimePlatforms := models.AllConcretePlatforms()
	var err error
	platforms, err = models.NormalizePlatforms(platforms)
	if err != nil {
		return nil, err
	}
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
			platforms = slices.Clone(defaultRuntimePlatforms)
		default:
			return nil, fmt.Errorf("unsupported preset: %s", preset)
		}
		platforms, err = models.NormalizePlatforms(platforms)
		if err != nil {
			return nil, err
		}
	}

	if len(platforms) == 0 {
		platforms = slices.Clone(defaultRuntimePlatforms)
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
			Delivery:    "both",
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
		}
	}

	if slices.Contains(platforms, "github-copilot") {
		targets["copilot"] = &models.TargetConfig{
			Description: "GitHub Copilot repository instructions and prompt files",
			Platforms:   []string{"github-copilot"},
			Profiles:    []string{"base", "devsecops", "documentation"},
			Delivery:    "both",
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
		}
	}

	if slices.Contains(platforms, "claude-code") {
		targets["claude"] = &models.TargetConfig{
			Description: "Claude Code CLAUDE.md and project skills",
			Platforms:   []string{"claude-code"},
			Profiles:    []string{"base", "devsecops"},
			Delivery:    "both",
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
		}
	}

	if slices.Contains(platforms, "gitlab-duo") {
		targets["gitlab"] = &models.TargetConfig{
			Description: "GitLab Duo Agent Platform setup with AGENTS.md, project-level skills, custom rules, and flow templates",
			Platforms:   []string{"gitlab-duo"},
			Profiles:    []string{"base", "gitlab-governance", "devsecops", "documentation"},
			Delivery:    "both",
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
			Flows:       slices.Clone(catalog.DevsecopsFlows),
			GitLabDuo: map[string]any{
				"slash_command": true,
			},
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
			Delivery:    "both",
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
			Model: map[string]any{
				"provider":      "ollama",
				"default_model": "qwen2.5-coder:7b",
				"base_url":      "http://localhost:11434",
			},
		}
	}

	skillsOnlyPlatforms := []string{}
	renderedTargets := map[string]struct{}{
		"codex":          {},
		"github-copilot": {},
		"claude-code":    {},
		"gitlab-duo":     {},
		"opencode":       {},
		"openhands":      {},
		"ollama":         {},
	}
	for _, p := range platforms {
		if _, rendered := renderedTargets[p]; !rendered {
			skillsOnlyPlatforms = append(skillsOnlyPlatforms, p)
		}
	}
	if len(skillsOnlyPlatforms) > 0 {
		targets["skills-only"] = &models.TargetConfig{
			Description: "Skill directory scaffolds for platforms without dedicated instruction templates",
			Platforms:   skillsOnlyPlatforms,
			Profiles:    []string{"base", "devsecops"},
			Delivery:    "skills",
			Skills:      slices.Clone(catalog.CoreSkills),
			Rules:       deepMerge(map[string]any{}, catalog.BaseRules),
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

	defaultPlatforms := []string{"codex"}
	if len(platforms) > 0 {
		defaultPlatforms = platforms
	}

	return &models.BakeConfig{
		Version:   "1",
		Variables: variables,
		SkillSources: &models.SkillSourceConfig{
			Defaults: models.SkillSourceDefaults{
				Version:        "0.1.0",
				Owner:          ownerTeam,
				License:        "MIT",
				CompatibleWith: defaultPlatforms,
			},
		},
		Skills: &models.SkillIntegrationConfig{
			Source: "agent-skills.lock",
			Mode:   models.SkillModeReference,
		},
		Targets: targets,
	}, nil
}
