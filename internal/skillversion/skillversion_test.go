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

func TestSyncArtifacts(t *testing.T) {
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

	// Diverge VERSION and skill.yaml from SKILL.md.
	if err := os.WriteFile(filepath.Join(skillDir, "VERSION"), []byte("9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("name: secure-code-review\nversion: 9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := SyncArtifacts(skillDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "1.0.0" {
		t.Fatalf("SyncArtifacts returned wrong version: %q", info.Version)
	}

	version, err := os.ReadFile(filepath.Join(skillDir, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(version)) != "1.0.0" {
		t.Fatalf("VERSION not synced: %q", version)
	}

	got, ok, err := skillYAMLVersion(filepath.Join(skillDir, "skill.yaml"))
	if err != nil || !ok || got != "1.0.0" {
		t.Fatalf("skill.yaml not synced: %q ok=%v err=%v", got, ok, err)
	}

	changelog, err := os.ReadFile(filepath.Join(skillDir, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(changelog), "1.0.0") {
		t.Fatalf("CHANGELOG.md not synced: %s", changelog)
	}
}

func TestReleaseBundleFor(t *testing.T) {
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

	bundle, err := ReleaseBundleFor(skillDir, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Checks) == 0 {
		t.Fatal("expected at least one check entry")
	}
	if len(bundle.Changelog) == 0 {
		t.Fatal("expected at least one changelog entry")
	}
	if !strings.Contains(bundle.ReleaseNotes, "# Release Notes") {
		t.Fatalf("unexpected release notes: %q", bundle.ReleaseNotes)
	}
	if bundle.Changed != nil {
		t.Fatal("includeChanged=false should not populate Changed")
	}

	// With a since date that excludes entries: notes should still have header.
	bundle2, err := ReleaseBundleFor(skillDir, "9999-01-01", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(bundle2.ReleaseNotes, "# Release Notes") {
		t.Fatalf("expected header in filtered notes: %q", bundle2.ReleaseNotes)
	}
}

func TestEnsureChangelogEntry(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		version string
		date    string
		change  string
		check   string
	}{
		{
			name:    "empty file",
			input:   "",
			version: "1.0.0", date: "2026-06-12", change: "Initial release",
			check: "## 1.0.0 - 2026-06-12",
		},
		{
			name:    "existing entry same version",
			input:   "# Changelog\n\n## 1.0.0 - 2026-06-12\n\n- Initial release\n\n",
			version: "1.0.0", date: "2026-06-12", change: "Initial release",
			check: "# Changelog",
		},
		{
			name:    "prepends new version",
			input:   "# Changelog\n\n## 1.0.0 - 2026-06-10\n\n- Old entry\n\n",
			version: "1.1.0", date: "2026-06-12", change: "New feature",
			check: "## 1.1.0 - 2026-06-12",
		},
		{
			name:    "replaces same version with new date",
			input:   "# Changelog\n\n## 1.0.0 - 2026-06-10\n\n- Old entry\n\n",
			version: "1.0.0", date: "2026-06-12", change: "Updated entry",
			check: "## 1.0.0 - 2026-06-12",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ensureChangelogEntry(tc.input, tc.version, tc.date, tc.change)
			if !strings.Contains(out, tc.check) {
				t.Fatalf("expected %q in output:\n%s", tc.check, out)
			}
		})
	}
}

func TestVersionFileVersion(t *testing.T) {
	dir := t.TempDir()

	// Missing file → false, no error.
	got, ok, err := versionFileVersion(filepath.Join(dir, "VERSION"))
	if err != nil || ok {
		t.Fatalf("missing file: got=%q ok=%v err=%v", got, ok, err)
	}

	// Present file with trailing newline.
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.2.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok, err = versionFileVersion(filepath.Join(dir, "VERSION"))
	if err != nil || !ok || got != "1.2.3" {
		t.Fatalf("present file: got=%q ok=%v err=%v", got, ok, err)
	}
}

func TestSkillYAMLVersion(t *testing.T) {
	dir := t.TempDir()

	// Missing file → false, no error.
	got, ok, err := skillYAMLVersion(filepath.Join(dir, "skill.yaml"))
	if err != nil || ok {
		t.Fatalf("missing file: got=%q ok=%v err=%v", got, ok, err)
	}

	// Valid YAML with version field.
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("name: my-skill\nversion: 2.3.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok, err = skillYAMLVersion(filepath.Join(dir, "skill.yaml"))
	if err != nil || !ok || got != "2.3.4" {
		t.Fatalf("valid yaml: got=%q ok=%v err=%v", got, ok, err)
	}

	// YAML without version field returns empty string.
	if err := os.WriteFile(filepath.Join(dir, "skill.yaml"), []byte("name: my-skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok, err = skillYAMLVersion(filepath.Join(dir, "skill.yaml"))
	if err != nil || !ok || got != "" {
		t.Fatalf("no version field: got=%q ok=%v err=%v", got, ok, err)
	}
}

func TestSamePath(t *testing.T) {
	dir := t.TempDir()
	if !samePath(dir, dir) {
		t.Fatal("same dir should be equal")
	}
	if samePath(dir, filepath.Join(dir, "other")) {
		t.Fatal("different paths should not be equal")
	}
	// Trailing slash should still match.
	if !samePath(dir+"/", dir) {
		t.Fatal("trailing slash variant should be equal")
	}
}

func TestNearestSkillFile(t *testing.T) {
	root := t.TempDir()

	// Create .agents/skills/my-skill/SKILL.md
	skillDir := filepath.Join(root, ".agents", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: my-skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A file inside the skill dir resolves to the skill's SKILL.md.
	got := nearestSkillFile(root, ".agents/skills/my-skill/tests/README.md")
	want := ".agents/skills/my-skill/SKILL.md"
	if got != want {
		t.Fatalf("nearestSkillFile(.../tests/README.md) = %q, want %q", got, want)
	}

	// The SKILL.md itself.
	got = nearestSkillFile(root, ".agents/skills/my-skill/SKILL.md")
	if got != want {
		t.Fatalf("nearestSkillFile(SKILL.md itself) = %q, want %q", got, want)
	}

	// A path with no SKILL.md ancestor returns empty string.
	got = nearestSkillFile(root, "some/unrelated/file.txt")
	if got != "" {
		t.Fatalf("unrelated path: expected empty, got %q", got)
	}
}

func TestNormalizeRepoRelPath(t *testing.T) {
	root := t.TempDir()

	// Only the dot-prefixed directory exists — function should prepend the dot.
	dotDir := filepath.Join(root, ".agents", "skills", "my-skill")
	if err := os.MkdirAll(dotDir, 0o755); err != nil {
		t.Fatal(err)
	}
	got := normalizeRepoRelPath(root, "agents/skills/my-skill")
	if got != ".agents/skills/my-skill" {
		t.Fatalf("dot-prefix path: got %q", got)
	}

	// Path already has dot prefix → returned as-is regardless.
	got = normalizeRepoRelPath(root, ".agents/skills/my-skill")
	if got != ".agents/skills/my-skill" {
		t.Fatalf("already-dotted path: got %q", got)
	}

	// Non-existent path with no dot variant → returned as-is.
	got = normalizeRepoRelPath(root, "nonexistent/path")
	if got != "nonexistent/path" {
		t.Fatalf("nonexistent path: got %q", got)
	}

	// File exists at exact path (no dot) → returned as-is.
	regularDir := filepath.Join(root, "skills", "plain-skill")
	if err := os.MkdirAll(regularDir, 0o755); err != nil {
		t.Fatal(err)
	}
	got = normalizeRepoRelPath(root, "skills/plain-skill")
	if got != "skills/plain-skill" {
		t.Fatalf("existing non-dot path: got %q", got)
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
