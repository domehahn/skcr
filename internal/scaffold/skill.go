package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/sklib/spec"
	"gopkg.in/yaml.v3"
)

type SkillOptions struct {
	Name        string
	OutputDir   string
	Version     string
	Description string
	Owner       string
	Stability   string
	Platforms   []string
	License     string
	Force       bool
	DryRun      bool
}

type PlannedFile struct {
	Path    string
	Content string
}

var (
	writeFile = os.WriteFile
	mkdirAll  = os.MkdirAll
	stat      = os.Stat
)

func PlanSkill(opts SkillOptions) ([]PlannedFile, error) {
	if err := validateSkillOptions(&opts); err != nil {
		return nil, err
	}
	root := filepath.Join(opts.OutputDir, opts.Name)
	platformBlock := markdownList(opts.Platforms)
	description := opts.Description
	if description == "" {
		description = "Describe what this skill helps an agent do."
	}

	skillSpec := spec.Skill{
		Name:           opts.Name,
		Version:        opts.Version,
		Description:    description,
		Entrypoint:     spec.DefaultEntrypointValue,
		License:        opts.License,
		CompatibleWith: stringsToSpecPlatforms(opts.Platforms),
	}
	if opts.Owner != "" {
		skillSpec.Owners = []string{opts.Owner}
	}
	skillYAMLBytes, err := yaml.Marshal(skillSpec)
	if err != nil {
		return nil, fmt.Errorf("marshal skill.yaml: %w", err)
	}

	return []PlannedFile{
		{Path: filepath.Join(root, "SKILL.md"), Content: skillMarkdown(opts.Name, description, opts.License, opts.Version, opts.Stability, opts.Owner, opts.Platforms)},
		{Path: filepath.Join(root, "skill.yaml"), Content: string(skillYAMLBytes)},
		{Path: filepath.Join(root, "VERSION"), Content: opts.Version + "\n"},
		{Path: filepath.Join(root, "CHANGELOG.md"), Content: fmt.Sprintf("# Changelog\n\n## %s\n\n- Initial skill scaffold.\n", opts.Version)},
		{Path: filepath.Join(root, "README.md"), Content: fmt.Sprintf("# %s\n\nThis is an AI agent skill scaffolded with `skcr`.\n\n## Version\n\nCurrent version: `%s`\n\n## Compatible platforms\n\n%s\n## Lifecycle\n\nAfter editing this skill, use `skpm` for lifecycle management:\n\n```bash\nskpm validate %s\nskpm package %s\nskpm publish %s\n```\n", opts.Name, opts.Version, platformBlock, opts.Name, opts.Name, opts.Name)},
		{Path: filepath.Join(root, "LICENSE"), Content: licenseText(opts.License)},
		{Path: filepath.Join(root, "tests", "README.md"), Content: fmt.Sprintf("# %s Tests\n\nAdd examples, fixtures, and expected outputs for this skill here.\n", opts.Name)},
	}, nil
}

func WriteSkill(opts SkillOptions) ([]PlannedFile, error) {
	files, err := PlanSkill(opts)
	if err != nil {
		return nil, err
	}
	if opts.DryRun {
		return files, nil
	}
	for _, file := range files {
		if !opts.Force {
			if _, err := stat(file.Path); err == nil {
				return nil, fmt.Errorf("%s already exists. Use --force to overwrite", file.Path)
			} else if !os.IsNotExist(err) {
				return nil, err
			}
		}
		if err := mkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			return nil, err
		}
		if err := writeFile(file.Path, []byte(file.Content), 0o644); err != nil {
			return nil, err
		}
	}
	return files, nil
}

// SkillWriteResult records which files were created or skipped during a safe write.
type SkillWriteResult struct {
	Created []PlannedFile
	Skipped []PlannedFile
}

// WriteSkillSafe writes a skill skeleton, skipping existing files instead of erroring.
// Use --force to overwrite, --dry-run to preview without writing.
func WriteSkillSafe(opts SkillOptions) (*SkillWriteResult, error) {
	files, err := PlanSkill(opts)
	if err != nil {
		return nil, err
	}
	result := &SkillWriteResult{}
	if opts.DryRun {
		result.Created = files
		return result, nil
	}
	for _, file := range files {
		if !opts.Force {
			if _, statErr := stat(file.Path); statErr == nil {
				result.Skipped = append(result.Skipped, file)
				continue
			} else if !os.IsNotExist(statErr) {
				return nil, statErr
			}
		}
		if err := mkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			return nil, err
		}
		if err := writeFile(file.Path, []byte(file.Content), 0o644); err != nil {
			return nil, err
		}
		result.Created = append(result.Created, file)
	}
	return result, nil
}

func validateSkillOptions(opts *SkillOptions) error {
	opts.Name = strings.TrimSpace(opts.Name)
	if err := spec.ValidateSkillName(opts.Name); err != nil {
		return fmt.Errorf("invalid skill name %q: use lowercase letters, digits, and hyphens; do not start or end with a hyphen", opts.Name)
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	normalized, err := spec.NormalizeVersion(opts.Version)
	if err != nil {
		return fmt.Errorf("invalid skill version %q: expected semver like 0.1.0", opts.Version)
	}
	opts.Version = normalized
	if opts.License == "" {
		opts.License = "MIT"
	}
	if opts.Stability == "" {
		opts.Stability = "experimental"
	}
	if opts.Owner == "" {
		opts.Owner = "platform-engineering"
	}
	if len(opts.Platforms) == 0 {
		opts.Platforms = []string{"claude-code", "codex"}
	}
	platforms, err := models.NormalizePlatforms(opts.Platforms)
	if err != nil {
		return err
	}
	opts.Platforms = platforms
	return nil
}

func skillMarkdown(name, description, license, version, stability, owner string, platforms []string) string {
	today := time.Now().Format("2006-01-02")
	if stability == "" {
		stability = "experimental"
	}
	if owner == "" {
		owner = "platform-engineering"
	}

	// Use skill-specific template if one is registered.
	data := skillTemplateData{
		Name:         name,
		Title:        skillTitle(name),
		Description:  description,
		Version:      version,
		Since:        today,
		LastModified: today,
		Owner:        owner,
		Stability:    stability,
		License:      license,
		Platforms:    platforms,
	}
	if rendered, err := renderSkillTemplate(name, data); err == nil && rendered != "" {
		return rendered
	}

	// Generic fallback for unregistered skill names.
	platformBlock := ""
	for _, p := range platforms {
		platformBlock += fmt.Sprintf("  %s: \"unknown\"\n", p)
	}
	if platformBlock == "" {
		platformBlock = "  codex: \"unknown\"\n"
	}

	frontmatter := fmt.Sprintf(`---
name: %s
description: %s
version: "%s"
since: "%s"
last_modified: "%s"
authors:
  - %s
stability: %s
min_platform_version:
%sdeprecated_since:
replaces:
supersedes: []
changelog:
  - version: "%s"
    date: "%s"
    change: "Initial release"
`, name, description, version, today, today, owner, stability, platformBlock, version, today)

	if license != "" {
		frontmatter += fmt.Sprintf("license: %s\n", license)
	}
	frontmatter += "---\n"

	return frontmatter + fmt.Sprintf(`
# %s

## Purpose

%s

## When to Use This Skill

Use this skill when the task matches the description above or when central agent instructions route work to $%s.

## Skill-Specific Operating Model

1. Clarify the goal and constraints.
2. Inspect the minimum relevant repository context.
3. Produce a concise execution plan for non-trivial work.
4. Execute with tools when implementation is requested.
5. Validate the result with repository-native checks.
6. Summarize changed files, validation results, and residual risks.

## Skill-Specific Checklist

- [ ] Confirm the request matches the stated purpose and identify affected files, systems, or workflows.
- [ ] Check concrete risks, edge cases, and validation needs for: %s
- [ ] Verify outputs are actionable and do not rely on generic guidance.

## Decision Rules

- Use this skill only when its purpose directly applies or central routing selects it.
- Escalate to a domain-specific skill when a more targeted skill covers the risk.
- Do not claim production readiness unless checklist, acceptance criteria, and output requirements are satisfied.

## DevSecOps Guardrails

- Do not read secrets, .env files, private keys, production credentials, masked CI/CD variables, or sensitive logs unless explicitly required.
- Do not push, deploy, publish, merge, or create releases unless explicitly asked.
- Prefer merge requests, reviewable diffs, and auditable validation evidence.
- Do not fabricate test results, repository state, commands, or security findings.

## Output Requirements

- State the actions taken or analysis performed.
- List files, systems, or workflows reviewed or changed.
- Report validation performed, findings or risks, and the recommended next step.

## Acceptance Criteria

- The result addresses the specific purpose of %s rather than only restating the request.
- Findings or changes are backed by repository evidence, validation, or clearly stated assumptions.
- Security, governance, and platform compatibility constraints remain intact.

## Anti-Patterns

- Producing generic guidance not grounded in the specific repository context.
- Skipping the skill-specific checklist and jumping directly to recommendations.
- Conflating symptom with root cause without evidence.

## Changelog

### %s - %s

- Initial release.
`, name, description, name, name, name, version, today)
}

func markdownList(values []string) string {
	lines := []string{}
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return strings.Join(lines, "\n") + "\n\n"
}

func stringsToSpecPlatforms(values []string) []spec.Platform {
	out := make([]spec.Platform, len(values))
	for i, v := range values {
		out[i] = spec.Platform(v)
	}
	return out
}

func licenseText(name string) string {
	if strings.EqualFold(name, "MIT") {
		return "MIT License\n\nCopyright (c) YEAR OWNER\n\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files, to deal in the Software without restriction.\n"
	}
	return name + "\n"
}
