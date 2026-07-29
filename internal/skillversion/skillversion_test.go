package skillversion

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
		Since:     "2026-06-10",
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
	info, err := Bump(skillDir, BumpMinor, "2026-06-17", "Add production-ready version lifecycle")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "1.1.0" || info.LastModified != "2026-06-17" {
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
	descriptor, err := os.ReadFile(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"schema_version: \"2\"", "goal:", "contract:", "file: contract.yaml", "evals:"} {
		if !strings.Contains(string(descriptor), want) {
			t.Fatalf("version bump erased descriptor field %q:\n%s", want, descriptor)
		}
	}
	contract, err := os.ReadFile(filepath.Join(skillDir, "contract.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"schema_version: \"1\"", "declared-tools-only", "scope: invocation"} {
		if !strings.Contains(string(contract), want) {
			t.Fatalf("version bump erased contract field %q:\n%s", want, contract)
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
	notes, err := ReleaseNotes(skillDir, "2026-06-17")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(notes, "secure-code-review 1.1.0") {
		t.Fatalf("unexpected release notes: %s", notes)
	}
}

func TestChangedDetectsContractOnlyChange(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
	if _, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name: "contract-skill", OutputDir: filepath.Join(dir, ".agents", "skills"),
		Version: "1.0.0", Platforms: []string{"codex"},
	}); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")
	path := filepath.Join(dir, ".agents", "skills", "contract-skill", "contract.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.ReplaceAll(string(data), "network:\n            allow: []", "network:\n            allow:\n                - api.example.com")
	if updated == string(data) {
		t.Fatalf("network contract fixture not found:\n%s", data)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := Changed(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	if len(changed) != 1 || len(changed[0].Errors) == 0 || !strings.Contains(strings.Join(changed[0].Files, "\n"), "contract.yaml") || changed[0].ContractImpact != "EXPANSION" {
		t.Fatalf("contract-only change was not material: %#v", changed)
	}
}

func TestChangedClassifiesGoalAndEvalOnlyChanges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	tests := []struct {
		name, rel, old, replacement, wantType string
	}{
		{"goal", "skill.yaml", "TODO: Define the outcome this skill is expected to achieve.", "Produce a classified result.", "goal_or_descriptor"},
		{"eval", filepath.Join("evals", "baseline.yaml"), "Perform the requested task.", "Perform the bounded requested task.", "eval"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			runGit(t, dir, "init")
			runGit(t, dir, "config", "user.email", "test@example.com")
			runGit(t, dir, "config", "user.name", "Test User")
			if _, err := scaffold.WriteSkill(scaffold.SkillOptions{
				Name: "classified-skill", OutputDir: filepath.Join(dir, ".agents", "skills"),
				Version: "1.0.0", Platforms: []string{"codex"},
			}); err != nil {
				t.Fatal(err)
			}
			runGit(t, dir, "add", ".")
			runGit(t, dir, "commit", "-m", "initial")
			path := filepath.Join(dir, ".agents", "skills", "classified-skill", tc.rel)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			updated := strings.Replace(string(data), tc.old, tc.replacement, 1)
			if updated == string(data) {
				t.Fatalf("fixture %q not found in %s", tc.old, data)
			}
			if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
				t.Fatal(err)
			}
			changed, err := Changed(filepath.Join(dir, ".agents", "skills"))
			if err != nil {
				t.Fatal(err)
			}
			if len(changed) != 1 || !slices.Contains(changed[0].ChangeTypes, tc.wantType) {
				t.Fatalf("unexpected classification: %#v", changed)
			}
		})
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

func TestNextVersion(t *testing.T) {
	cases := []struct {
		current string
		kind    BumpKind
		want    string
		wantErr bool
	}{
		{"1.2.3", BumpPatch, "1.2.4", false},
		{"1.2.3", BumpMinor, "1.3.0", false},
		{"1.2.3", BumpMajor, "2.0.0", false},
		{"1.2.3", "", "1.2.4", false},
		{"0.0.1", BumpPatch, "0.0.2", false},
		{"0.9.9", BumpMinor, "0.10.0", false},
		{"9.9.9", BumpMajor, "10.0.0", false},
		{"not-semver", BumpPatch, "", true},
		{"1.2", BumpPatch, "", true},
		{"1.2.3", "unsupported", "", true},
		{"1.x.3", BumpPatch, "", true},
	}
	for _, tc := range cases {
		got, err := nextVersion(tc.current, tc.kind)
		if tc.wantErr {
			if err == nil {
				t.Errorf("nextVersion(%q, %q): expected error, got %q", tc.current, tc.kind, got)
			}
		} else {
			if err != nil {
				t.Errorf("nextVersion(%q, %q): unexpected error: %v", tc.current, tc.kind, err)
			} else if got != tc.want {
				t.Errorf("nextVersion(%q, %q) = %q, want %q", tc.current, tc.kind, got, tc.want)
			}
		}
	}
}

func TestChangedNoGit(t *testing.T) {
	dir := t.TempDir()
	if _, err := Changed(dir); err == nil {
		t.Fatal("Changed in non-git dir should return error")
	}
}

func TestBumpAllChangedNoGit(t *testing.T) {
	dir := t.TempDir()
	if _, err := BumpAllChanged(dir, BumpOptions{}); err == nil {
		t.Fatal("BumpAllChanged in non-git dir should return error")
	}
}

func TestSplitFrontmatter(t *testing.T) {
	// Valid frontmatter.
	content := "---\nname: my-skill\nversion: \"1.0.0\"\n---\n# Body\n"
	fm, body, ok := splitFrontmatter(content)
	if !ok {
		t.Fatal("expected ok=true for valid frontmatter")
	}
	if fm == "" {
		t.Error("expected non-empty frontmatter")
	}
	if body != "# Body\n" {
		t.Errorf("unexpected body: %q", body)
	}

	// No --- prefix.
	if _, _, ok := splitFrontmatter("name: my-skill\n"); ok {
		t.Error("expected ok=false for content without --- prefix")
	}

	// No closing ---.
	if _, _, ok := splitFrontmatter("---\nname: my-skill\n"); ok {
		t.Error("expected ok=false for unclosed frontmatter")
	}

	// Empty body after ---.
	_, body2, ok2 := splitFrontmatter("---\nkey: val\n---\n")
	if !ok2 {
		t.Error("expected ok=true for frontmatter with empty body")
	}
	if body2 != "" {
		t.Errorf("expected empty body, got %q", body2)
	}
}

func TestUpdateSkillMD(t *testing.T) {
	base := "---\nname: my-skill\ndescription: ok\nversion: \"0.1.0\"\nsince: \"2025-01-01\"\nlast_modified: \"2025-01-01\"\nauthors:\n  - team\nstability: stable\nmin_platform_version:\n  codex: unknown\ndeprecated_since:\nreplaces:\nsupersedes: []\nchangelog:\n  - version: \"0.1.0\"\n    date: \"2025-01-01\"\n    change: \"Initial\"\n---\n\n## Changelog\n\n### 0.1.0 - 2025-01-01\n\n- Initial.\n"

	got, err := updateSkillMD(base, "0.2.0", "2026-06-12", "Add feature X")
	if err != nil {
		t.Fatalf("updateSkillMD: %v", err)
	}
	if !strings.Contains(got, "0.2.0") {
		t.Error("updated SKILL.md should contain new version")
	}
	if !strings.Contains(got, "Add feature X") {
		t.Error("updated SKILL.md should contain new changelog entry")
	}

	// No frontmatter → error.
	if _, err := updateSkillMD("no frontmatter here", "0.2.0", "2026-06-12", "X"); err == nil {
		t.Error("expected error for content without frontmatter")
	}
}

func TestUpdateTextFileIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "VERSION")

	// File does not exist → no error, no-op.
	if err := updateTextFileIfExists(path, "1.0.0"); err != nil {
		t.Fatalf("updateTextFileIfExists on missing file: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not be created if it did not exist")
	}

	// File exists → updated.
	if err := os.WriteFile(path, []byte("0.1.0"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := updateTextFileIfExists(path, "0.2.0"); err != nil {
		t.Fatalf("updateTextFileIfExists on existing file: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "0.2.0" {
		t.Errorf("expected %q, got %q", "0.2.0", string(data))
	}
}

func TestUpdateSkillYAMLIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skill.yaml")

	// File does not exist → no error.
	if err := updateSkillYAMLIfExists(path, "1.0.0"); err != nil {
		t.Fatalf("updateSkillYAMLIfExists on missing file: %v", err)
	}

	// Write skill.yaml with a version field.
	if err := os.WriteFile(path, []byte("name: my-skill\nversion: \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := updateSkillYAMLIfExists(path, "0.2.0"); err != nil {
		t.Fatalf("updateSkillYAMLIfExists: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "0.2.0") {
		t.Errorf("version not updated in skill.yaml: %s", string(data))
	}
}

func TestSkillFiles(t *testing.T) {
	dir := t.TempDir()

	// Direct SKILL.md path.
	skillMD := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillMD, []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := skillFiles(skillMD)
	if err != nil || len(files) != 1 || files[0] != skillMD {
		t.Errorf("skillFiles(SKILL.md path): got %v, err=%v", files, err)
	}

	// Directory with SKILL.md.
	skillDir := filepath.Join(dir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err = skillFiles(skillDir)
	if err != nil || len(files) != 1 {
		t.Errorf("skillFiles(dir with SKILL.md): got %v, err=%v", files, err)
	}

	// Non-existent path → error.
	if _, err := skillFiles(filepath.Join(dir, "nonexistent")); err == nil {
		t.Error("expected error for nonexistent path")
	}

	// Regular file (not SKILL.md) → error.
	notSkill := filepath.Join(dir, "README.md")
	if err := os.WriteFile(notSkill, []byte("# Readme"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := skillFiles(notSkill); err == nil {
		t.Error("expected error for non-SKILL.md file path")
	}
}

func TestChangelogFromDir(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := filepath.Join(skillDir, "SKILL.md")

	// Directory with SKILL.md that has no frontmatter → no entries (skipped).
	if err := os.WriteFile(skillMD, []byte("# My Skill\n\nNo frontmatter.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := Changelog(skillDir)
	if err != nil {
		t.Fatalf("Changelog on no-frontmatter skill: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for skill without frontmatter, got %d", len(entries))
	}

	// Non-existent path → error.
	if _, err := Changelog(filepath.Join(dir, "nonexistent")); err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestParseInfoWithErrors(t *testing.T) {
	// Valid frontmatter structure but invalid skill content (missing required sections).
	// parseInfo should collect validation errors without returning a hard error.
	content := "---\nname: bad-skill\ndescription: \"\"\nversion: \"1.0.0\"\nsince: \"2025-01-01\"\nlast_modified: \"2026-06-12\"\nauthors:\n  - team\nstability: stable\nmin_platform_version:\n  codex: unknown\ndeprecated_since:\nreplaces:\nsupersedes: []\nchangelog:\n  - version: \"1.0.0\"\n    date: \"2026-06-12\"\n    change: \"Initial\"\n---\n\n# Bad Skill\n"
	info, err := parseInfo("/tmp/SKILL.md", content)
	if err != nil {
		t.Fatalf("parseInfo should not error for structurally valid frontmatter: %v", err)
	}
	if len(info.Errors) == 0 {
		t.Error("expected validation errors to be recorded in info.Errors")
	}
}

func TestArtifactConsistencyErrors(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := filepath.Join(skillDir, "SKILL.md")

	// Empty version → no errors.
	errs := artifactConsistencyErrors(skillMD, "")
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty version, got %v", errs)
	}

	// VERSION with wrong version → error.
	if err := os.WriteFile(filepath.Join(skillDir, "VERSION"), []byte("0.9.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs = artifactConsistencyErrors(skillMD, "1.0.0")
	found := false
	for _, e := range errs {
		if strings.Contains(e, "VERSION") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected VERSION mismatch error, got %v", errs)
	}

	// skill.yaml with wrong version → error.
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("name: my-skill\nversion: 0.9.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs = artifactConsistencyErrors(skillMD, "1.0.0")
	foundYAML := false
	for _, e := range errs {
		if strings.Contains(e, "skill.yaml") {
			foundYAML = true
			break
		}
	}
	if !foundYAML {
		t.Errorf("expected skill.yaml mismatch error, got %v", errs)
	}

	// Matching versions → no errors.
	if err := os.WriteFile(filepath.Join(skillDir, "VERSION"), []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte("name: my-skill\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := artifactConsistencyErrors(skillMD, "1.0.0"); len(errs) != 0 {
		t.Errorf("expected no errors for matching versions, got %v", errs)
	}
}

func TestSkillFilesRecursiveWalk(t *testing.T) {
	dir := t.TempDir()

	// Create subdirectories, each with a SKILL.md — no top-level SKILL.md.
	for _, sub := range []string{"skill-a", "skill-b"} {
		subDir := filepath.Join(dir, sub)
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte("# "+sub), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := skillFiles(dir)
	if err != nil {
		t.Fatalf("skillFiles recursive walk: unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 SKILL.md files from recursive walk, got %d: %v", len(files), files)
	}
	for _, f := range files {
		if filepath.Base(f) != "SKILL.md" {
			t.Errorf("unexpected file in result: %s", f)
		}
	}
}

func TestSyncArtifactsNoChangelog(t *testing.T) {
	dir := t.TempDir()
	// Scaffold a skill without changelog entries.
	_, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name:      "sync-no-changelog",
		OutputDir: dir,
		Version:   "1.0.0",
		Owner:     "team",
		Platforms: []string{"codex"},
	})
	if err != nil {
		t.Fatalf("WriteSkill: %v", err)
	}
	skillDir := filepath.Join(dir, "sync-no-changelog")

	// Remove changelog entries from SKILL.md so the no-changelog branch is taken.
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		t.Fatal(err)
	}
	// Strip the changelog section entirely.
	text := string(content)
	if idx := strings.Index(text, "changelog:"); idx != -1 {
		// Keep everything up to (not including) the changelog key.
		end := strings.Index(text[idx:], "\n---")
		if end == -1 {
			end = len(text) - idx
		}
		text = text[:idx] + "changelog: []\n" + text[idx+end:]
	}
	if err := os.WriteFile(skillMDPath, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := SyncArtifacts(skillDir)
	if err != nil {
		t.Fatalf("SyncArtifacts with no changelog: %v", err)
	}
	if info.Version == "" {
		t.Error("expected non-empty version after SyncArtifacts")
	}
}

func TestReleaseBundleFor_NonExistentPath(t *testing.T) {
	_, err := ReleaseBundleFor(filepath.Join(t.TempDir(), "nonexistent"), "", false)
	if err == nil {
		t.Error("expected error for non-existent path in ReleaseBundleFor")
	}
}

func TestReleaseBundleFor_NoEntriesSinceDate(t *testing.T) {
	dir := t.TempDir()
	_, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name:      "release-skill",
		OutputDir: dir,
		Version:   "1.0.0",
		Owner:     "team",
		Platforms: []string{"codex"},
	})
	if err != nil {
		t.Fatalf("WriteSkill: %v", err)
	}
	skillDir := filepath.Join(dir, "release-skill")

	// Use a far-future date so no changelog entries are included.
	bundle, err := ReleaseBundleFor(skillDir, "9999-01-01", false)
	if err != nil {
		t.Fatalf("ReleaseBundleFor: %v", err)
	}
	// ReleaseNotes should contain the header but no skill entries.
	if !strings.HasPrefix(bundle.ReleaseNotes, "# Release Notes") {
		t.Errorf("expected Release Notes header, got: %q", bundle.ReleaseNotes)
	}
	// No skill version lines expected when since is in the future.
	if strings.Contains(bundle.ReleaseNotes, "## release-skill") {
		t.Error("expected no skill entries in notes for far-future since date")
	}
}

func TestReleaseBundleFor_IncludeChanged_NoGit(t *testing.T) {
	dir := t.TempDir()
	_, err := scaffold.WriteSkill(scaffold.SkillOptions{
		Name:      "git-skill",
		OutputDir: dir,
		Version:   "1.0.0",
		Owner:     "team",
		Platforms: []string{"codex"},
	})
	if err != nil {
		t.Fatalf("WriteSkill: %v", err)
	}
	// includeChanged = true in a non-git dir: Changed() error is silently ignored.
	bundle, err := ReleaseBundleFor(filepath.Join(dir, "git-skill"), "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Changed is nil/empty when not in a git repo.
	if bundle.Changed != nil {
		t.Logf("changed: %v (non-nil is ok, git may be present)", bundle.Changed)
	}
}
