package skillversion

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/scaffold"
)

func TestBumpSynchronizesSkillArtifacts(t *testing.T) {
	dir := t.TempDir()
	files, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name:      "secure-code-review",
		OutputDir: dir,
		Version:   "1.0.0",
		Owner:     "platform-engineering",
		Platforms: []string{"codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected scaffold files")
	}
	skillDir := filepath.Join(dir, "secure-code-review")
	info, err := Bump(skillDir, BumpMinor, "2026-06-12", "Add production-ready version lifecycle")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "1.1.0" || info.LastModified != "2026-06-12" {
		t.Fatalf("unexpected bumped info: %#v", info)
	}
	for _, path := range []string{"SKILL.md", "VERSION", "skill.yaml", "CHANGELOG.md"} {
		content, err := os.ReadFile(filepath.Join(skillDir, path))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "1.1.0") {
			t.Fatalf("%s was not updated: %s", path, content)
		}
	}
	infos, err := Check(skillDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 1 || len(infos[0].Errors) != 0 {
		t.Fatalf("expected clean check, got %#v", infos)
	}
	entries, err := Changelog(skillDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 || entries[0].Version != "1.1.0" {
		t.Fatalf("unexpected changelog: %#v", entries)
	}
	notes, err := ReleaseNotes(skillDir, "2026-06-12")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(notes, "secure-code-review 1.1.0") {
		t.Fatalf("unexpected release notes: %s", notes)
	}
}

func TestCheckDetectsDivergentSkillArtifacts(t *testing.T) {
	dir := t.TempDir()
	if _, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name:      "secure-code-review",
		OutputDir: dir,
		Version:   "1.0.0",
		Owner:     "platform-engineering",
		Platforms: []string{"codex"},
	}); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, "secure-code-review")
	if err := os.WriteFile(filepath.Join(skillDir, "VERSION"), []byte("9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("name: secure-code-review\nversion: 9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "CHANGELOG.md"), []byte("# Changelog\n\n## 9.9.9 - 2026-06-12\n\n- Drift.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	infos, err := Check(skillDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected one skill info, got %#v", infos)
	}
	errText := strings.Join(infos[0].Errors, "\n")
	for _, want := range []string{"VERSION", "skill.yaml", "CHANGELOG.md"} {
		if !strings.Contains(errText, want) {
			t.Fatalf("expected %s drift error, got %#v", want, infos[0].Errors)
		}
	}
}

func TestBumpRejectsInvalidInputs(t *testing.T) {
	if _, err := Bump(t.TempDir(), BumpPatch, "2026-06-12", ""); err == nil {
		t.Fatal("expected missing change error")
	}
	if _, err := Bump(t.TempDir(), BumpPatch, "2026/06/12", "change"); err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestChangedAndBumpAllChanged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	if _, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name:      "secure-code-review",
		OutputDir: filepath.Join(dir, ".agents", "skills"),
		Version:   "1.0.0",
		Owner:     "platform-engineering",
		Platforms: []string{"codex"},
	}); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial skill")

	skillDir := filepath.Join(dir, ".agents", "skills", "secure-code-review")
	skillMD := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillMD, append(content, []byte("\nExtra production guidance.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := Changed(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) == 0 {
		root, rootErr := gitRoot(filepath.Join(dir, ".agents", "skills"))
		files, filesErr := gitChangedFiles(filepath.Join(dir, ".agents", "skills"), ".")
		status := exec.Command("git", "-C", dir, "status", "--porcelain")
		statusOut, _ := status.CombinedOutput()
		t.Fatalf("expected changed skill without version bump, got none root=%q rootErr=%v files=%#v filesErr=%v status=%q", root, rootErr, files, filesErr, statusOut)
	}
	if len(changed) != 1 || len(changed[0].Errors) == 0 {
		t.Fatalf("expected changed skill without version bump, got %#v", changed)
	}
	dry, err := BumpAllChanged(filepath.Join(dir, ".agents", "skills"), BumpOptions{
		Kind:   BumpPatch,
		Date:   "2026-06-12",
		Change: "Update production guidance",
		DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dry) != 1 || !dry[0].DryRun || dry[0].NewVersion != "1.0.1" {
		t.Fatalf("unexpected dry run results: %#v", dry)
	}
	still, err := os.ReadFile(filepath.Join(skillDir, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(still)) != "1.0.0" {
		t.Fatalf("dry run modified VERSION: %s", still)
	}
	results, err := BumpAllChanged(filepath.Join(dir, ".agents", "skills"), BumpOptions{
		Kind:   BumpPatch,
		Date:   "2026-06-12",
		Change: "Update production guidance",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].NewVersion != "1.0.1" {
		t.Fatalf("unexpected bump results: %#v", results)
	}
	changed, err = Changed(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 1 || len(changed[0].Errors) != 0 || !changed[0].VersionChanged {
		t.Fatalf("expected changed skill with version bump, got %#v", changed)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
