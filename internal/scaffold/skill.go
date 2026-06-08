package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/domehahn/skcr/internal/models"
)

type SkillOptions struct {
	Name        string
	OutputDir   string
	Version     string
	Description string
	Owner       string
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
	validSkillName    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	validSkillVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`)
	writeFile         = os.WriteFile
	mkdirAll          = os.MkdirAll
	stat              = os.Stat
)

func PlanSkill(opts SkillOptions) ([]PlannedFile, error) {
	if err := validateSkillOptions(&opts); err != nil {
		return nil, err
	}
	root := filepath.Join(opts.OutputDir, opts.Name)
	platformBlock := markdownList(opts.Platforms)
	yamlPlatforms := yamlList(opts.Platforms)
	description := opts.Description
	if description == "" {
		description = "Describe what this skill helps an agent do."
	}

	return []PlannedFile{
		{Path: filepath.Join(root, "SKILL.md"), Content: skillMarkdown(opts.Name)},
		{Path: filepath.Join(root, "skill.yaml"), Content: fmt.Sprintf("name: %s\nversion: %s\ndescription: %s\nowner: %s\nlicense: %s\ncompatible_with:\n%s", opts.Name, opts.Version, quoteYAML(description), quoteYAML(opts.Owner), quoteYAML(opts.License), yamlPlatforms)},
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

func validateSkillOptions(opts *SkillOptions) error {
	opts.Name = strings.TrimSpace(opts.Name)
	if !validSkillName.MatchString(opts.Name) {
		return fmt.Errorf("invalid skill name %q: use lowercase letters, digits, and hyphens; do not start or end with a hyphen", opts.Name)
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "."
	}
	if opts.Version == "" {
		opts.Version = "0.1.0"
	}
	if !validSkillVersion.MatchString(opts.Version) {
		return fmt.Errorf("invalid skill version %q: expected semver like 0.1.0", opts.Version)
	}
	if opts.License == "" {
		opts.License = "MIT"
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

func skillMarkdown(name string) string {
	return fmt.Sprintf(`# %s

## Purpose

Describe what this skill helps an agent do.

## When to use this skill

Use this skill when...

## Inputs

- Repository context
- User request
- Relevant files or diffs

## Workflow

1. Inspect the request.
2. Identify the relevant context.
3. Apply the skill-specific rules.
4. Produce a clear, actionable result.

## Output

Describe the expected output format.
`, name)
}

func yamlList(values []string) string {
	lines := []string{}
	for _, value := range values {
		lines = append(lines, "  - "+value)
	}
	return strings.Join(lines, "\n") + "\n"
}

func markdownList(values []string) string {
	lines := []string{}
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return strings.Join(lines, "\n") + "\n\n"
}

func quoteYAML(value string) string {
	if value == "" {
		return `""`
	}
	escaped := strings.ReplaceAll(value, `"`, `\"`)
	return `"` + escaped + `"`
}

func licenseText(name string) string {
	if strings.EqualFold(name, "MIT") {
		return "MIT License\n\nCopyright (c) YEAR OWNER\n\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files, to deal in the Software without restriction.\n"
	}
	return name + "\n"
}
