package skilllock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/models"
)

func TestLoadFilterReferencesAndSkillFiles(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".agents", "skills", "secure-code-review")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: secure-code-review\ndescription: ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(dir, "agent-skills.lock")
	lock := `skills:
  secure-code-review:
    version: v1.2.3
    source: registry
    checksum: sha256:test
    compatible_with:
      - codex
    installed_paths:
      - .agents/skills/secure-code-review
`
	if err := os.WriteFile(lockPath, []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	state, err := Load(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Skills) != 1 || state.Skills[0].Name != "secure-code-review" {
		t.Fatalf("unexpected state: %#v", state)
	}
	if got := FilterByPlatforms(state.Skills, []string{"gitlab-duo"}); len(got) != 0 {
		t.Fatalf("expected filtered out skill, got %#v", got)
	}
	filtered := FilterByPlatforms(state.Skills, []string{"codex"})
	refs := References(filtered)
	if refs[0]["path"] != ".agents/skills/secure-code-review/SKILL.md" {
		t.Fatalf("unexpected reference path: %#v", refs[0])
	}
	files, err := SkillFiles(dir, filtered, models.SkillModeCopy, []string{"codex"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || !strings.Contains(files[0].Content, "secure-code-review") {
		t.Fatalf("unexpected copy files: %#v", files)
	}
	links, err := SkillFiles(dir, filtered, models.SkillModeLink, []string{"codex"})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].LinkTarget == "" {
		t.Fatalf("expected link target, got %#v", links)
	}
}

func TestLoadErrorsAreSkpmOriented(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "agent-skills.lock")); err == nil || !strings.Contains(err.Error(), "skpm") {
		t.Fatalf("expected skpm missing lock hint, got %v", err)
	}
	bad := filepath.Join(t.TempDir(), "bad.lock")
	if err := os.WriteFile(bad, []byte("skills: [x]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bad); err == nil || !strings.Contains(err.Error(), "skpm verify") {
		t.Fatalf("expected skpm verify hint, got %v", err)
	}
}

func TestPlatformSkillDestinationUsesCapabilityMatrix(t *testing.T) {
	got := PlatformSkillDestination("antigravity", "security-reviewer")
	want := ".agent/skills/security-reviewer/SKILL.md"
	if got != want {
		t.Fatalf("PlatformSkillDestination antigravity = %q, want %q", got, want)
	}
	got = PlatformSkillDestination("qwen", "security-reviewer")
	want = ".qwen/skills/security-reviewer/SKILL.md"
	if got != want {
		t.Fatalf("PlatformSkillDestination qwen = %q, want %q", got, want)
	}
}
