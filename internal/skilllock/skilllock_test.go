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

func TestFilterByPlatformsEdgeCases(t *testing.T) {
	skills := []LockedSkill{
		{Name: "a", CompatibleWith: []string{}},
		{Name: "b", CompatibleWith: []string{"codex", "claude-code"}},
		{Name: "c", CompatibleWith: []string{"gitlab-duo"}},
	}

	// Empty wanted platforms → all returned.
	got := FilterByPlatforms(skills, []string{})
	if len(got) != 3 {
		t.Fatalf("empty filter: expected all 3, got %d", len(got))
	}

	// Skill with no CompatibleWith always included.
	got = FilterByPlatforms(skills, []string{"codex"})
	names := make([]string, len(got))
	for i, s := range got {
		names[i] = s.Name
	}
	if len(got) != 2 {
		t.Fatalf("expected skills a and b for codex filter, got %v", names)
	}

	// Invalid platform name: silently skipped, all returned.
	got = FilterByPlatforms(skills, []string{"NOT_A_PLATFORM"})
	if len(got) != 3 {
		t.Fatalf("invalid platform should result in no filter (all included), got %d", len(got))
	}
}

func TestSkillFilesErrorPaths(t *testing.T) {
	dir := t.TempDir()

	// Missing installed path → error.
	skills := []LockedSkill{{
		Name:           "my-skill",
		InstalledPaths: []string{".agents/skills/my-skill/SKILL.md"},
		CompatibleWith: []string{"codex"},
	}}
	if _, err := SkillFiles(dir, skills, models.SkillModeCopy, []string{"codex"}); err == nil {
		t.Fatal("expected error for missing installed path")
	}

	// Empty installed paths → error.
	skills2 := []LockedSkill{{
		Name:           "my-skill",
		InstalledPaths: []string{},
		CompatibleWith: []string{"codex"},
	}}
	if _, err := SkillFiles(dir, skills2, models.SkillModeCopy, []string{"codex"}); err == nil {
		t.Fatal("expected error for empty installed paths")
	}

	// Unknown mode → nil (no error, no files).
	files, err := SkillFiles(dir, skills, "unknown-mode", []string{"codex"})
	if err != nil || files != nil {
		t.Fatalf("unknown mode: expected nil result, got %v %v", files, err)
	}
}

func TestLoadEmptyLockFile(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "agent-skills.lock")
	if err := os.WriteFile(empty, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	state, err := Load(empty)
	if err != nil {
		t.Fatalf("empty lock should load cleanly: %v", err)
	}
	if len(state.Skills) != 0 {
		t.Fatalf("empty lock should have 0 skills, got %d", len(state.Skills))
	}
}

func TestReferencesFields(t *testing.T) {
	skills := []LockedSkill{{
		Name:           "my-skill",
		Version:        "1.2.3",
		Source:         "registry",
		Checksum:       "sha256:abc",
		CompatibleWith: []string{"codex"},
		InstalledPaths: []string{".agents/skills/my-skill/SKILL.md"},
	}}
	refs := References(skills)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if refs[0]["name"] != "my-skill" {
		t.Errorf("name mismatch: %v", refs[0]["name"])
	}
	if refs[0]["version"] != "1.2.3" {
		t.Errorf("version mismatch: %v", refs[0]["version"])
	}
	if refs[0]["path"] != ".agents/skills/my-skill/SKILL.md" {
		t.Errorf("path mismatch: %v", refs[0]["path"])
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

func TestNormalizeStringList(t *testing.T) {
	// scalar string.
	got := normalizeStringList("codex")
	if len(got) != 1 || got[0] != "codex" {
		t.Errorf("scalar string: got %v", got)
	}

	// empty string → empty slice.
	if got := normalizeStringList(""); len(got) != 0 {
		t.Errorf("empty string: got %v", got)
	}

	// []any slice.
	got = normalizeStringList([]any{"codex", "cursor", ""})
	if len(got) != 2 || got[0] != "codex" || got[1] != "cursor" {
		t.Errorf("[]any slice: got %v", got)
	}

	// []string slice.
	got = normalizeStringList([]string{"codex"})
	if len(got) != 1 || got[0] != "codex" {
		t.Errorf("[]string: got %v", got)
	}

	// nil / unknown type → empty.
	if got := normalizeStringList(nil); len(got) != 0 {
		t.Errorf("nil: got %v", got)
	}
	if got := normalizeStringList(42); len(got) != 0 {
		t.Errorf("int: got %v", got)
	}
}

func TestPlatformSkillDestination_KnownPlatforms(t *testing.T) {
	cases := map[string]string{
		"codex":         ".agents/skills/my-skill/SKILL.md",
		"claude-code":   ".claude/skills/my-skill/SKILL.md",
		"gitlab-duo":    "skills/my-skill/SKILL.md",
		"github-copilot": ".github/prompts/my-skill.prompt.md",
		"cursor":        ".agentic/skills/my-skill/SKILL.md",
		"windsurf":      ".agentic/skills/my-skill/SKILL.md",
		"generic":       ".agentic/skills/my-skill/SKILL.md",
	}
	for platform, want := range cases {
		got := PlatformSkillDestination(platform, "my-skill")
		if got != want {
			t.Errorf("PlatformSkillDestination(%q, my-skill) = %q, want %q", platform, got, want)
		}
	}
}

func TestPlatformSkillDestination_Unknown(t *testing.T) {
	// Unknown platform with no capability registration → empty string.
	got := PlatformSkillDestination("totally-unknown-platform-xyz", "my-skill")
	// Either empty or a reasonable path — just must not panic.
	_ = got
}

func TestParseSkillsMapFormat(t *testing.T) {
	// Map-format skills block (named entries).
	raw := map[string]any{
		"skills": map[string]any{
			"my-skill": map[string]any{
				"version":        "1.0.0",
				"compatible_with": []any{"codex"},
				"installed_paths": []any{".agents/skills/my-skill"},
			},
		},
	}
	skills, err := parseSkills(raw)
	if err != nil {
		t.Fatalf("parseSkills map format: %v", err)
	}
	if len(skills) != 1 || skills[0].Name != "my-skill" {
		t.Errorf("unexpected skills: %+v", skills)
	}
}

func TestParseSkillsInvalidFormat(t *testing.T) {
	// skills value is neither list nor map.
	raw := map[string]any{"skills": "bad-value"}
	if _, err := parseSkills(raw); err == nil {
		t.Fatal("expected error for invalid skills format")
	}
}

func TestParseSkillsMissingName(t *testing.T) {
	// List format with no name field and no fallback → error.
	raw := map[string]any{
		"skills": []any{
			map[string]any{"version": "1.0.0"},
		},
	}
	if _, err := parseSkills(raw); err == nil {
		t.Fatal("expected error for skill with missing name")
	}
}
