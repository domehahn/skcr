package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteSkillScaffold(t *testing.T) {
	dir := t.TempDir()
	files, err := WriteSkill(SkillOptions{
		Name:        "secure-code-review",
		OutputDir:   dir,
		Version:     "0.1.0",
		Description: "Security-focused code review skill",
		Owner:       "platform-engineering",
		Platforms:   []string{"codex", "claude-code", "gitlab-duo"},
		License:     "MIT",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 7 {
		t.Fatalf("expected 7 scaffold files, got %d", len(files))
	}
	for _, rel := range []string{"SKILL.md", "skill.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("tests", "README.md")} {
		if _, err := os.Stat(filepath.Join(dir, "secure-code-review", rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	version, err := os.ReadFile(filepath.Join(dir, "secure-code-review", "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if string(version) != "0.1.0\n" {
		t.Fatalf("unexpected VERSION: %q", version)
	}
	meta, err := os.ReadFile(filepath.Join(dir, "secure-code-review", "skill.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"name: secure-code-review", "version: 0.1.0", "- codex", "- claude-code", "- gitlab-duo"} {
		if !strings.Contains(string(meta), want) {
			t.Fatalf("skill.yaml missing %q:\n%s", want, meta)
		}
	}
}

func TestPlanSkillDefaultsDryRunAndValidation(t *testing.T) {
	dir := t.TempDir()
	files, err := WriteSkill(SkillOptions{Name: "test-generator", OutputDir: dir, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 7 {
		t.Fatalf("expected planned files, got %d", len(files))
	}
	if _, err := os.Stat(filepath.Join(dir, "test-generator")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create directory, err=%v", err)
	}
	if _, err := PlanSkill(SkillOptions{Name: "SecureCodeReview"}); err == nil {
		t.Fatal("expected invalid uppercase name error")
	}
	if _, err := PlanSkill(SkillOptions{Name: "secure_code_review"}); err == nil {
		t.Fatal("expected invalid underscore name error")
	}
	if _, err := PlanSkill(SkillOptions{Name: "-secure-code-review"}); err == nil {
		t.Fatal("expected invalid leading hyphen error")
	}
	if _, err := PlanSkill(SkillOptions{Name: "secure-code-review-", Version: "0.1.0"}); err == nil {
		t.Fatal("expected invalid trailing hyphen error")
	}
	if _, err := PlanSkill(SkillOptions{Name: "secure-code-review", Version: "v1"}); err == nil {
		t.Fatal("expected invalid semver error")
	}
	if _, err := PlanSkill(SkillOptions{Name: "secure-code-review", Platforms: []string{"bad"}}); err == nil {
		t.Fatal("expected invalid platform error")
	}
}

func TestWriteSkillForce(t *testing.T) {
	dir := t.TempDir()
	opts := SkillOptions{Name: "secure-code-review", OutputDir: dir}
	if _, err := WriteSkill(opts); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteSkill(opts); err == nil {
		t.Fatal("expected existing file error")
	}
	opts.Force = true
	if _, err := WriteSkill(opts); err != nil {
		t.Fatalf("expected force overwrite to succeed: %v", err)
	}
}

func TestWriteSkillSafe(t *testing.T) {
	dir := t.TempDir()
	opts := SkillOptions{Name: "my-safe-skill", OutputDir: dir}

	// First write: all files created.
	result, err := WriteSkillSafe(opts)
	if err != nil {
		t.Fatalf("WriteSkillSafe first write: %v", err)
	}
	if len(result.Created) != 7 {
		t.Fatalf("expected 7 created files, got %d", len(result.Created))
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("expected 0 skipped files, got %d", len(result.Skipped))
	}

	// Second write without force: all files skipped, no error.
	result2, err := WriteSkillSafe(opts)
	if err != nil {
		t.Fatalf("WriteSkillSafe second write (no-force): %v", err)
	}
	if len(result2.Created) != 0 {
		t.Fatalf("expected 0 created (all exist), got %d", len(result2.Created))
	}
	if len(result2.Skipped) != 7 {
		t.Fatalf("expected 7 skipped, got %d", len(result2.Skipped))
	}

	// With force: all files overwritten.
	opts.Force = true
	result3, err := WriteSkillSafe(opts)
	if err != nil {
		t.Fatalf("WriteSkillSafe force: %v", err)
	}
	if len(result3.Created) != 7 {
		t.Fatalf("expected 7 created with force, got %d", len(result3.Created))
	}
	if len(result3.Skipped) != 0 {
		t.Fatalf("expected 0 skipped with force, got %d", len(result3.Skipped))
	}
}

func TestWriteSkillSafeDryRun(t *testing.T) {
	dir := t.TempDir()
	opts := SkillOptions{Name: "dry-skill", OutputDir: dir, DryRun: true}
	result, err := WriteSkillSafe(opts)
	if err != nil {
		t.Fatalf("WriteSkillSafe dry-run: %v", err)
	}
	if len(result.Created) != 7 {
		t.Fatalf("expected 7 planned files, got %d", len(result.Created))
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("expected 0 skipped in dry-run, got %d", len(result.Skipped))
	}
	if _, err := os.Stat(filepath.Join(dir, "dry-skill")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write files, err=%v", err)
	}
}

func TestWriteSkillSafeInvalidOptions(t *testing.T) {
	dir := t.TempDir()
	_, err := WriteSkillSafe(SkillOptions{Name: "Invalid_Name", OutputDir: dir})
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
}
