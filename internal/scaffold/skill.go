package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/domehahn/skcr/internal/models"
	platformcompat "github.com/domehahn/skcr/internal/platforms"
	"github.com/domehahn/sklib/spec"
	"gopkg.in/yaml.v3"
)

type SkillOptions struct {
	Name        string
	OutputDir   string
	Version     string
	Since       string
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
	if opts.Since == "" {
		if existing := readExistingSince(filepath.Join(root, "SKILL.md")); existing != "" {
			opts.Since = existing
		}
	}
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
		{Path: filepath.Join(root, "SKILL.md"), Content: skillMarkdown(opts.Name, description, opts.License, opts.Version, opts.Stability, opts.Owner, opts.Since, opts.Platforms)},
		{Path: filepath.Join(root, "skill.yaml"), Content: string(skillYAMLBytes)},
		{Path: filepath.Join(root, "VERSION"), Content: opts.Version + "\n"},
		{Path: filepath.Join(root, "CHANGELOG.md"), Content: fmt.Sprintf("# Changelog\n\n## %s\n\n- Initial skill scaffold.\n", opts.Version)},
		{Path: filepath.Join(root, "README.md"), Content: fmt.Sprintf("# %s\n\nThis is an AI agent skill scaffolded with `skcr`.\n\n## Version\n\nCurrent version: `%s`\n\n## Compatible platforms\n\n%s\n## Lifecycle\n\nAfter editing this skill, use `skpm` for lifecycle management:\n\n```bash\nskpm validate %s\nskpm package %s\nskpm publish %s\n```\n", opts.Name, opts.Version, platformBlock, opts.Name, opts.Name, opts.Name)},
		{Path: filepath.Join(root, "LICENSE"), Content: licenseText(opts.License)},
		{Path: filepath.Join(root, "scripts", "README.md"), Content: fmt.Sprintf("# %s Scripts\n\nPlace executable helper scripts for this skill here. Keep scripts self-contained, document dependencies, and reference them from `SKILL.md` only when the agent should run them.\n", opts.Name)},
		{Path: filepath.Join(root, "references", "README.md"), Content: fmt.Sprintf("# %s References\n\nPlace focused supplemental documentation for this skill here. Agents should load these files on demand via relative links from `SKILL.md`.\n", opts.Name)},
		{Path: filepath.Join(root, "assets", "README.md"), Content: fmt.Sprintf("# %s Assets\n\nPlace templates, static resources, schemas, diagrams, example payloads, or lookup tables for this skill here.\n", opts.Name)},
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

// ReadExistingSince returns the since date from an existing SKILL.md frontmatter,
// or empty string if the file doesn't exist or has no valid since field.
func ReadExistingSince(path string) string {
	return readExistingSince(path)
}

func readExistingSince(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "since:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "since:"))
			val = strings.Trim(val, `"`)
			if val != "" && val != "YYYY-MM-DD" {
				return val
			}
		}
	}
	return ""
}

func skillMarkdown(name, description, license, version, stability, owner, since string, platforms []string) string {
	today := time.Now().Format("2006-01-02")
	if stability == "" {
		stability = "experimental"
	}
	if owner == "" {
		owner = "platform-engineering"
	}
	if since == "" {
		since = today
	}

	// Use skill-specific template if one is registered.
	data := skillTemplateData{
		Name:         name,
		Title:        skillTitle(name),
		Description:  description,
		Version:      version,
		Since:        since,
		LastModified: today,
		Owner:        owner,
		Stability:    stability,
		License:      license,
		Platforms:    platforms,
		MinPlatforms: platformcompat.AllMinVersions(),
	}
	if rendered, err := renderSkillTemplate(name, data); err == nil && rendered != "" {
		return rendered
	}

	// Generic fallback for unregistered skill names.
	platformBlock := minPlatformVersionBlock(platforms)

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
`, name, description, version, since, today, owner, stability, platformBlock, version, today)

	if license != "" {
		frontmatter += fmt.Sprintf("license: %s\n", license)
	}
	frontmatter += "---\n"

	return frontmatter + fmt.Sprintf(`
# %s

## Purpose

%s

## When to use

- The task matches this skill's purpose or central routing selects $%s.
- Repository evidence must be turned into concrete findings, implementation guidance, or validation steps.
- Security, governance, platform compatibility, release, or operational risk must be considered.
- The result needs a clear pass, conditional pass, block, or implementation-ready recommendation.
- A production-ready skill body is required rather than generic placeholder guidance.

## Operating model

1. Clarify the goal, scope, constraints, and expected output.
2. Inspect the minimum relevant repository context before making claims.
3. Apply the skill-specific review scope and checklist.
4. Classify findings using severity guidance.
5. Validate with repository-native checks where practical.
6. Summarize evidence, residual risk, and next actions.

## Spec-Driven Change Context

- Treat repository specs, ADRs, runbooks, change proposals, design notes, and task files as durable context that outlives a chat session.
- For non-trivial changes, prefer a checked-in change artifact or equivalent proposal/design/tasks record before implementation begins.
- Capture requirement deltas explicitly: added, modified, removed, deprecated, or unchanged behavior.
- Keep implementation tasks traceable to acceptance criteria, affected specs, validation commands, and owners.
- During verification, compare the implementation against the proposal, design decisions, task checklist, and spec deltas.
- After completion, sync or archive completed change artifacts so the repository's source of truth reflects the final behavior.
- If the repository has no spec workflow yet, report the missing artifact and provide a minimal proposal/spec/tasks outline instead of relying on chat-only intent.

## Skill-Specific Review Scope

- The stated purpose and domain boundaries of %s.
- Repository files, generated outputs, workflows, or controls directly affected by the request.
- Security, governance, platform compatibility, validation, and rollback implications.
- Assumptions, dependencies, owners, and open decisions that affect safe completion.
- Evidence needed to support findings, recommendations, or implementation changes.

## Skill-Specific Checklist

- [ ] Confirm the request actually matches %s and identify affected files, systems, or workflows.
- [ ] Check the concrete risks, edge cases, and validation needs implied by: %s
- [ ] Verify outputs are actionable for the target platform and do not rely on generic copy-paste guidance.
- [ ] Confirm platform compatibility metadata is present and does not invent versions.
- [ ] Check security and governance guardrails before claiming completion.
- [ ] Identify required validation commands or explain why validation cannot run.
- [ ] Check whether generated or synchronized files need updating.
- [ ] Verify changelog and version metadata when instructions change materially.
- [ ] Identify residual risks and follow-up owners.
- [ ] Confirm output is specific to this skill's purpose.

## Decision Rules

- Use this skill only when its purpose directly applies or central routing selects it.
- Escalate to a more specific reviewer skill when the dominant risk is security, secrets, CI/CD, dependencies, IaC, observability, release readiness, or compliance.
- Do not claim production readiness unless checklist, severity guidance, acceptance criteria, validation, and output requirements are satisfied.
- When required repository proof is unavailable, state the limitation and reduce confidence instead of assuming success.
- If a recommendation changes security, release, governance, or platform behavior, require explicit review.

## Finding Categories

- Unsupported claim or recommendation for the requested domain.
- Unsafe implementation, workflow, policy, or operational recommendation.
- Missing validation, test, sync, or generated-output consistency.
- Security, compliance, privacy, release, or operational risk.
- Unclear ownership, assumption, decision, or residual risk.

## Severity Guidance

- Critical: immediate risk to secrets, regulated data, production safety, release integrity, or destructive behavior.
- High: credible security, compliance, compatibility, rollback, or operational risk requiring owner action before merge or release.
- Medium: meaningful maintainability, validation, documentation, or process gap that should be tracked.
- Low: advisory improvement or clarity issue with limited immediate impact.

## DevSecOps Guardrails

- Do not read secrets, .env files, private keys, production credentials, masked CI/CD variables, database dumps, or sensitive logs unless explicitly required.
- Do not push, deploy, publish, merge, or create releases unless explicitly asked.
- Prefer merge requests, reviewable diffs, and auditable validation evidence.
- Prefer least privilege, minimal changes, and explicit rollback notes.
- Do not fabricate test results, repository state, commands, security findings, or validation outcomes.
- Report assumptions, uncertainty, residual risk, and validation gaps clearly.

## Output Requirements

- State the actions taken or analysis performed.
- List files, systems, workflows, controls, or artifacts reviewed or changed.
- Report validation performed, findings or risks, and the recommended next step.
- Include severity, evidence, and remediation for each significant finding.
- Separate confirmed facts from assumptions and open questions.

## Acceptance Criteria

- The result addresses the specific purpose of %s rather than only restating the request.
- Findings or changes are backed by repository evidence, validation, or clearly stated assumptions.
- Security, governance, and platform compatibility constraints remain intact.
- Required sections, version metadata, and changelog are present.
- Residual risks and validation gaps are explicit.

## Anti-Patterns

- Creating a skill that only differs by name and description.
- Using generic checklist language that does not mention the skill domain.
- Claiming validation passed without running or naming validation.
- Hiding residual risk, assumptions, or missing evidence.
- Weakening governance or DevSecOps guardrails to simplify the output.

## Changelog

### %s - %s

- Initial release.
`, name, description, name, name, name, description, name, version, today)
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

func minPlatformVersionBlock(selected []string) string {
	entries := platformcompat.MinVersionsFor(selected)
	if len(entries) == 0 {
		for _, platform := range selected {
			entries = append(entries, platformcompat.CompatibilityEntry{
				Name:       platform,
				MinVersion: platformcompat.MinVersionOrUnknown(platform),
			})
		}
	}
	if len(entries) == 0 {
		entries = platformcompat.AllMinVersions()
	}
	var b strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&b, "  %s: \"%s\"\n", entry.Name, entry.MinVersion)
	}
	return b.String()
}

func licenseText(name string) string {
	if strings.EqualFold(name, "MIT") {
		return "MIT License\n\nCopyright (c) YEAR OWNER\n\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files, to deal in the Software without restriction.\n"
	}
	return name + "\n"
}
