package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateHelpers(t *testing.T) {
	if !containsAll("abc def", []string{"abc", "def"}) {
		t.Fatal("containsAll should be true")
	}
	if containsAll("abc", []string{"abc", "def"}) {
		t.Fatal("containsAll should be false")
	}
	if msg := validateSkillMetadata("name: test\ndescription: ok\n"); msg != "" {
		t.Fatalf("unexpected validateSkillMetadata msg: %q", msg)
	}
	if msg := validateSkillMetadata("description: ok\n"); !strings.Contains(msg, "name") {
		t.Fatalf("expected name error, got %q", msg)
	}
	if msg := validateSkillMetadata("name: x\ndescription: \"\"\n"); !strings.Contains(msg, "description") {
		t.Fatalf("expected description error, got %q", msg)
	}
	if !isEmptyMetadataValue("  \"\"  ") || !isEmptyMetadataValue(" '' ") || isEmptyMetadataValue("x") {
		t.Fatal("isEmptyMetadataValue behavior mismatch")
	}
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
		"Skill metadata name is missing or empty",
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
		[]byte("name: valid-skill\ndescription: A valid skill.\n"), 0o644); err != nil {
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
