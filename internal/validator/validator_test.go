package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/models"
)

func TestValidateHelpers(t *testing.T) {
	if !containsAll("abc def", []string{"abc", "def"}) {
		t.Fatal("containsAll should be true")
	}
	if containsAll("abc", []string{"abc", "def"}) {
		t.Fatal("containsAll should be false")
	}
	validSkill := validSkillFixture()
	if msg := validateSkillMetadata(validSkill); msg != "" {
		t.Fatalf("unexpected validateSkillMetadata msg: %q", msg)
	}
	if msg := validateSkillMetadata("description: ok\n"); !strings.Contains(msg, "frontmatter") {
		t.Fatalf("expected frontmatter error, got %q", msg)
	}
	if msg := validateSkillMetadata(strings.Replace(validSkill, "description: ok", "description: \"\"", 1)); !strings.Contains(msg, "description") {
		t.Fatalf("expected description error, got %q", msg)
	}
	for name, tc := range map[string]string{
		"leading v":                 strings.Replace(validSkill, `version: "1.0.0"`, `version: "v1.0.0"`, 1),
		"missing changelog section": strings.Replace(validSkill, "## Changelog", "## History", 1),
		"missing specific section":  strings.Replace(validSkill, "## Decision Rules", "## Choices", 1),
		"deprecated missing date":   strings.Replace(validSkill, "stability: stable", "stability: deprecated", 1),
		"old last modified":         strings.Replace(validSkill, `last_modified: "2026-06-10"`, `last_modified: "2026-06-09"`, 1),
	} {
		if msg := validateSkillMetadata(tc); msg == "" {
			t.Fatalf("expected metadata error for %s", name)
		}
	}
	if !isEmptyMetadataValue("  \"\"  ") || !isEmptyMetadataValue(" '' ") || isEmptyMetadataValue("x") {
		t.Fatal("isEmptyMetadataValue behavior mismatch")
	}
}

func validSkillFixture() string {
	return `---
name: test-skill
description: ok
version: "1.0.0"
since: "2025-01-01"
last_modified: "2026-06-10"
authors:
  - platform-engineering
stability: stable
min_platform_version:
  codex: "unknown"
deprecated_since:
replaces:
supersedes: []
changelog:
  - version: "1.0.0"
    date: "2026-06-10"
    change: "Initial release"
---

# Test Skill

## Purpose

Validate a test skill fixture.

## When to use

- Use when validator tests need a valid skill.

## Operating model

- Inspect the concrete domain before making recommendations.

## Spec-Driven Change Context

- Preserve proposal, spec, design, task, verification, sync, and archive context.

## Skill-Specific Review Scope

- Review domain-specific scope and affected controls.

## Skill-Specific Checklist

- Verify the domain-specific checks are complete.

## Decision Rules

- Prefer evidence-backed decisions.

## Finding Categories

- Missing evidence.

## Severity Guidance

- High: material risk.

## DevSecOps Guardrails

- Do not fabricate validation results.

## Acceptance Criteria

- The skill produces a domain-specific result.

## Output Requirements

- Report evidence and validation.

## Anti-Patterns

- Generic copy-paste output.

## Changelog

### 1.0.0 - 2026-06-10

- Initial release.
`
}

func TestValidateProjectScenarios(t *testing.T) {
	dir := t.TempDir()

	errs, err := ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 1 || errs[0] != "Missing agentic.bake.yaml" {
		t.Fatalf("unexpected missing bake errors: %#v", errs)
	}

	bakePath := filepath.Join(dir, "agentic.bake.yaml")
	if err := os.WriteFile(bakePath, []byte("targets: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateProject(dir); err == nil {
		t.Fatal("expected yaml parse error")
	}

	readErrDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(readErrDir, "agentic.bake.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateProject(readErrDir); err == nil {
		t.Fatal("expected read error when bakefile path is a directory")
	}

	validBake := `version: "1"
targets:
  t1:
    platforms:
      - invalid-platform
`
	if err := os.WriteFile(bakePath, []byte(validBake), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "skills", "badskill"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skills", "badskill", "SKILL.md"), []byte("name: \"\"\ndescription: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{filepath.Join(dir, ".agents/skills/miss"), filepath.Join(dir, ".claude/skills/miss"), filepath.Join(dir, ".agentic/skills/miss")} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.MkdirAll(filepath.Join(dir, ".gitlab", "duo", "flows"), 0o755); err != nil {
		t.Fatal(err)
	}
	flow := "name: bad\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitlab", "duo", "flows", "f.yaml"), []byte(flow), 0o644); err != nil {
		t.Fatal(err)
	}

	errs, err = ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(errs, "\n")
	checks := []string{
		"unsupported platform",
		"Skill metadata frontmatter is missing",
		"Skill missing SKILL.md",
		"GitLab Duo output missing .gitlab/duo/chat-rules.md",
		"forbidden top-level field",
		"workspace_agent_skills",
		"user_rule",
	}
	for _, c := range checks {
		if !strings.Contains(joined, c) {
			t.Fatalf("expected error containing %q, got:\n%s", c, joined)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, ".gitlab", "duo", "chat-rules.md"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	goodFlow := "version: \"1\"\nworkspace_agent_skills: x\nuser_rule: y\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitlab", "duo", "flows", "f.yaml"), []byte(goodFlow), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, ".agents/skills/miss")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, ".claude/skills/miss")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, ".agentic/skills/miss")); err != nil {
		t.Fatal(err)
	}

	if _, err := ValidateProject(string([]byte{0})); err == nil {
		t.Fatal("expected os.Stat error for invalid target path")
	}
}

func TestValidateSkillSources(t *testing.T) {
	// Use no targets in these tests to isolate skill_sources validation from
	// platform file rendering checks.
	dir := t.TempDir()
	bakePath := filepath.Join(dir, "agentic.bake.yaml")

	// Valid skill_sources with existing directories — should pass.
	bakeContent := `version: "1"
skill_sources:
  output_dir: skills
  defaults:
    compatible_with:
      - codex
  skills:
    - name: valid-skill
`
	if err := os.WriteFile(bakePath, []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "skills", "valid-skill"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skills", "valid-skill", "SKILL.md"),
		[]byte(strings.Replace(validSkillFixture(), "name: test-skill", "name: valid-skill", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	errs, err := ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range errs {
		if strings.Contains(e, "valid-skill") {
			t.Fatalf("unexpected skill_sources error for valid config: %q", e)
		}
	}

	// Missing skill source directory.
	bakeWithMissing := `version: "1"
skill_sources:
  output_dir: skills
  skills:
    - name: missing-skill
`
	if err := os.WriteFile(bakePath, []byte(bakeWithMissing), 0o644); err != nil {
		t.Fatal(err)
	}
	errs, err = ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(errs, "\n"), "missing-skill") {
		t.Fatalf("expected missing-skill error, got: %v", errs)
	}

	// Invalid skill name.
	bakeWithInvalid := `version: "1"
skill_sources:
  skills:
    - name: "Invalid_Name"
`
	if err := os.WriteFile(bakePath, []byte(bakeWithInvalid), 0o644); err != nil {
		t.Fatal(err)
	}
	errs, err = ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(errs, "\n"), "invalid skill name") {
		t.Fatalf("expected invalid name error, got: %v", errs)
	}

	// Duplicate skill names.
	bakeWithDup := `version: "1"
skill_sources:
  skills:
    - name: dup-skill
    - name: dup-skill
`
	if err := os.WriteFile(bakePath, []byte(bakeWithDup), 0o644); err != nil {
		t.Fatal(err)
	}
	errs, err = ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(errs, "\n"), "duplicate") {
		t.Fatalf("expected duplicate error, got: %v", errs)
	}

	// Invalid platform in skill_sources defaults.
	bakeWithBadPlatform := `version: "1"
skill_sources:
  defaults:
    compatible_with:
      - not-a-platform
  skills: []
`
	if err := os.WriteFile(bakePath, []byte(bakeWithBadPlatform), 0o644); err != nil {
		t.Fatal(err)
	}
	errs, err = ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(errs, "\n"), "unsupported platform") {
		t.Fatalf("expected unsupported platform error, got: %v", errs)
	}
}

func TestValidateProjectRejectsDuplicateSkillSpecificBlocks(t *testing.T) {
	dir := t.TempDir()
	bakePath := filepath.Join(dir, "agentic.bake.yaml")
	bakeContent := `version: "1"
skill_sources:
  output_dir: .agents/skills
  skills:
    - name: alpha-skill
    - name: beta-skill
`
	if err := os.WriteFile(bakePath, []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"alpha-skill", "beta-skill"} {
		skillDir := filepath.Join(dir, ".agents", "skills", name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := strings.Replace(validSkillFixture(), "name: test-skill", "name: "+name, 1)
		content = strings.Replace(content, "# Test Skill", "# "+name, 1)
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	errs, err := ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(errs, "\n"), "identical skill-specific content blocks") {
		t.Fatalf("expected duplicate skill-specific block error, got: %v", errs)
	}
}

func TestValidateProjectCoversContinueBranches(t *testing.T) {
	dir := t.TempDir()
	bakePath := filepath.Join(dir, "agentic.bake.yaml")
	content := `version: "1"
targets:
  not-map: "x"
`
	if err := os.WriteFile(bakePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skills", "README.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".gitlab", "duo", "flows", "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitlab", "duo", "flows", "note.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "missing.yaml"), filepath.Join(dir, ".gitlab", "duo", "flows", "broken.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitlab", "duo", "chat-rules.md"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	errs, err := ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors for continue-branch scenario, got: %v", errs)
	}

	if err := os.WriteFile(bakePath, []byte("targets: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs, err = ValidateProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(errs, "\n"), "No targets configured") {
		t.Fatalf("expected no-targets error, got %v", errs)
	}
}

func TestValidateSkillMetadataPublic(t *testing.T) {
	valid := validSkillFixture()
	if msg := ValidateSkillMetadata(valid); msg != "" {
		t.Fatalf("expected no error for valid skill, got %q", msg)
	}
	if msg := ValidateSkillMetadata("no frontmatter"); !strings.Contains(msg, "frontmatter") {
		t.Fatalf("expected frontmatter error, got %q", msg)
	}
	noDesc := strings.Replace(valid, "description: ok", `description: ""`, 1)
	if msg := ValidateSkillMetadata(noDesc); !strings.Contains(msg, "description") {
		t.Fatalf("expected description error for empty description, got %q", msg)
	}
	sinceAfterModified := strings.Replace(valid, `since: "2025-01-01"`, `since: "2027-01-01"`, 1)
	if msg := ValidateSkillMetadata(sinceAfterModified); !strings.Contains(msg, "since") {
		t.Fatalf("expected since > last_modified error, got %q", msg)
	}
}

func TestValidateSkillWarnings(t *testing.T) {
	withUnknown := validSkillFixture()
	warnings := ValidateSkillWarnings(withUnknown)
	if len(warnings) == 0 {
		t.Fatal("expected warning for unknown min_platform_version, got none")
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "unknown") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'unknown' in warnings, got %v", warnings)
	}

	withConcrete := strings.Replace(withUnknown, `codex: "unknown"`, `codex: "0.51.0"`, 1)
	warnings = ValidateSkillWarnings(withConcrete)
	for _, w := range warnings {
		if strings.Contains(w, "codex") && strings.Contains(w, "unknown") {
			t.Fatalf("unexpected unknown warning for concrete version: %q", w)
		}
	}

	if ValidateSkillWarnings("no frontmatter") != nil {
		t.Fatal("expected nil warnings for skill without frontmatter")
	}
}

func TestPlatformSkillBaseDir(t *testing.T) {
	cases := map[string]string{
		"claude-code":    ".claude/skills",
		"github-copilot": ".github/skills",
		"gitlab-duo":     "skills",
		"cursor":         ".cursor/skills",
		"junie":          ".junie/skills",
		"gemini-cli":     ".gemini/skills",
		"roo-code":       ".roo/skills",
		"kiro":           ".kiro/skills",
		"opencode":       ".opencode/skills",
		"openhands":      ".openhands/skills",
		"windsurf":       ".windsurf/skills",
		"ollama":         ".ollama/skills",
		"codex":          ".agents/skills",
		"unknown-tool":   ".agents/skills",
	}
	for platform, want := range cases {
		if got := platformSkillBaseDir(platform); got != want {
			t.Errorf("platformSkillBaseDir(%q) = %q, want %q", platform, got, want)
		}
	}
}

func TestValidateGeneratedState(t *testing.T) {
	dir := t.TempDir()

	// Empty lock + no expected files → no errors.
	if err := os.WriteFile(filepath.Join(dir, ".agentic-template.lock"), []byte(`{"version":"1","target":"default","files":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := validateGeneratedState(dir, nil); len(errs) != 0 {
		t.Fatalf("expected no errors for empty state, got %v", errs)
	}

	// Missing generated file → error.
	rf := models.RenderedFile{Platform: "codex", Destination: "missing.md", Content: "hello"}
	if errs := validateGeneratedState(dir, []models.RenderedFile{rf}); len(errs) == 0 {
		t.Fatal("expected error for missing generated file")
	}

	// Present file with wrong content → checksum mismatch.
	if err := os.WriteFile(filepath.Join(dir, "present.md"), []byte("wrong content"), 0o644); err != nil {
		t.Fatal(err)
	}
	rf2 := models.RenderedFile{Platform: "codex", Destination: "present.md", Content: "expected content"}
	if errs := validateGeneratedState(dir, []models.RenderedFile{rf2}); len(errs) == 0 {
		t.Fatal("expected checksum mismatch error for wrong content")
	}

	// SKILL.md paths are exempt from checksum check.
	skillPath := filepath.Join(dir, ".agents", "skills", "test-skill")
	if err := os.MkdirAll(skillPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte("different"), 0o644); err != nil {
		t.Fatal(err)
	}
	rfSkill := models.RenderedFile{Platform: "codex", Destination: ".agents/skills/test-skill/SKILL.md", Content: "expected"}
	if errs := validateGeneratedState(dir, []models.RenderedFile{rfSkill}); len(errs) != 0 {
		t.Fatalf("SKILL.md should be exempt from checksum check, got %v", errs)
	}
}

func TestValidateSkillLock(t *testing.T) {
	dir := t.TempDir()

	// Missing lock file → error.
	errs := validateSkillLock(dir, "agent-skills.lock", []string{"codex"})
	if len(errs) == 0 {
		t.Fatal("expected error for missing skill lock")
	}

	// Empty lock → no errors.
	lockContent := `version: "1"
skills: []
`
	lockPath := filepath.Join(dir, "agent-skills.lock")
	if err := os.WriteFile(lockPath, []byte(lockContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := validateSkillLock(dir, "agent-skills.lock", []string{"codex"}); len(errs) != 0 {
		t.Fatalf("expected no errors for empty lock, got %v", errs)
	}

	// Locked skill with no install paths → error.
	lockWithSkill := `version: "1"
skills:
  - name: my-skill
    version: "1.0.0"
    compatible_with:
      - codex
    installed_paths: []
`
	if err := os.WriteFile(lockPath, []byte(lockWithSkill), 0o644); err != nil {
		t.Fatal(err)
	}
	errs = validateSkillLock(dir, "agent-skills.lock", []string{"codex"})
	if len(errs) == 0 {
		t.Fatal("expected error for skill with no install paths")
	}

	// Locked skill with missing path → error.
	lockWithMissing := `version: "1"
skills:
  - name: my-skill
    version: "1.0.0"
    compatible_with:
      - codex
    installed_paths:
      - .agents/skills/my-skill
`
	if err := os.WriteFile(lockPath, []byte(lockWithMissing), 0o644); err != nil {
		t.Fatal(err)
	}
	errs = validateSkillLock(dir, "agent-skills.lock", []string{"codex"})
	if len(errs) == 0 {
		t.Fatal("expected error for skill with missing install path")
	}

	// Locked skill with valid SKILL.md present → no errors.
	skillDir := filepath.Join(dir, ".agents", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := validateSkillLock(dir, "agent-skills.lock", []string{"codex"}); len(errs) != 0 {
		t.Fatalf("expected no errors for valid skill install, got %v", errs)
	}
}
