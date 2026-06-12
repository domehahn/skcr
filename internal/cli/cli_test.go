package cli

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/scaffold"
	"gopkg.in/yaml.v3"
)

func runRoot(args ...string) error {
	root := NewRootCommand()
	root.SetArgs(args)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	return root.Execute()
}

func TestRootAndVersionCommand(t *testing.T) {
	root := NewRootCommand()
	if root.Use != "skcr" {
		t.Fatalf("unexpected root use: %s", root.Use)
	}

	Version, Commit, Date = "1.2.3", "abcdef1", "2026-06-04"
	cmd := newVersionCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1.2.3") || !strings.Contains(out, "abcdef1") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

func TestSkillVersionLifecycleCommands(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "secure-code-review", "--output-dir", dir, "--version", "1.0.0", "--platform", "codex"); err != nil {
		t.Fatalf("scaffold skill failed: %v", err)
	}
	skillDir := filepath.Join(dir, "secure-code-review")
	if err := runRoot("version", "check", skillDir); err != nil {
		t.Fatalf("version check failed: %v", err)
	}
	if err := runRoot("version", "check", skillDir, "--json"); err != nil {
		t.Fatalf("version check json failed: %v", err)
	}
	if err := runRoot("version", "bump", skillDir, "--kind", "patch", "--date", "2026-06-12", "--change", "Preview release lifecycle automation", "--dry-run", "--json"); err != nil {
		t.Fatalf("version bump dry-run json failed: %v", err)
	}
	if err := runRoot("version", "bump", skillDir, "--kind", "patch", "--date", "2026-06-12", "--change", "Add release lifecycle automation"); err != nil {
		t.Fatalf("version bump failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `version: 1.0.1`) && !strings.Contains(string(content), `version: "1.0.1"`) {
		t.Fatalf("SKILL.md was not bumped: %s", content)
	}
	if err := runRoot("version", "changelog", skillDir); err != nil {
		t.Fatalf("version changelog failed: %v", err)
	}
	if err := runRoot("version", "release-notes", skillDir, "--since", "2026-06-12"); err != nil {
		t.Fatalf("version release-notes failed: %v", err)
	}
	if err := runRoot("version", "release-bundle", skillDir, "--since", "2026-06-12", "--json"); err != nil {
		t.Fatalf("version release-bundle failed: %v", err)
	}
}

func TestApplyBuildInfo(t *testing.T) {
	origV, origC, origD := Version, Commit, Date
	t.Cleanup(func() {
		Version, Commit, Date = origV, origC, origD
	})

	Version, Commit, Date = "fixed", "none", "unknown"
	applyBuildInfo(&debug.BuildInfo{}, true)
	if Version != "fixed" {
		t.Fatal("expected early return when version not dev")
	}

	Version, Commit, Date = "dev", "none", "unknown"
	applyBuildInfo(nil, false)
	if Version != "dev" {
		t.Fatal("expected no change when build info unavailable")
	}

	Version, Commit, Date = "dev", "none", "unknown"
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.0.0"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890"},
			{Key: "vcs.time", Value: "2026-06-04T00:00:00Z"},
		},
	}
	applyBuildInfo(info, true)
	if Version != "v1.0.0" || Commit != "1234567" || Date != "2026-06-04T00:00:00Z" {
		t.Fatalf("unexpected version info: %s %s %s", Version, Commit, Date)
	}

	Version, Commit, Date = "dev", "none", "unknown"
	info2 := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "123"},
		},
	}
	applyBuildInfo(info2, true)
	if Version != "dev" || Commit != "none" {
		t.Fatalf("expected unchanged short revision/devel, got %s %s", Version, Commit)
	}
}

func TestInitListBakeValidateFlow(t *testing.T) {
	dir := t.TempDir()

	if err := runRoot("init", "--target", dir, "--platform", "codex,gitlab-duo", "--project-name", "Demo"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "agentic.bake.yaml")); err != nil {
		t.Fatalf("missing bake file: %v", err)
	}
	if err := runRoot("init", "--target", dir); err == nil {
		t.Fatal("expected init error when bake file exists")
	}
	if err := runRoot("init", "--target", dir, "--force", "--preset", "minimal"); err != nil {
		t.Fatalf("force init failed: %v", err)
	}
	if err := runRoot("list-targets", "--target", dir); err != nil {
		t.Fatalf("list-targets failed: %v", err)
	}

	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake write failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agentic-template.lock")); err != nil {
		t.Fatalf("missing lockfile: %v", err)
	}

	agentsPath := filepath.Join(dir, ".agentic", "codex", "AGENTS.md")
	orig, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentsPath, append(orig, []byte("\nchanged\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "default", "--target", dir, "--plan", "--detailed-diff"); err != nil {
		t.Fatalf("bake plan failed: %v", err)
	}

	if err := runRoot("validate", "--target", dir); err == nil {
		t.Fatal("expected validate to detect changed generated file")
	}
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake rewrite failed: %v", err)
	}
	if err := runRoot("validate", "--target", dir); err != nil {
		t.Fatalf("validate failed: %v", err)
	}

	// Break validation deliberately.
	skill := filepath.Join(dir, ".agents", "skills", "security-reviewer", "SKILL.md")
	if err := os.WriteFile(skill, []byte("name: \"\"\ndescription: \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("validate", "--target", dir); err == nil {
		t.Fatal("expected validate failure")
	}
}

func TestCommandErrorPaths(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("list-targets", "--target", dir); err == nil {
		t.Fatal("expected list-targets error without bakefile")
	}
	if err := runRoot("bake", "default", "--target", dir); err == nil {
		t.Fatal("expected bake error without bakefile")
	}
	if err := runRoot("validate", "--target", filepath.Join(dir, "does-not-exist", string([]byte{0}))); err == nil {
		t.Fatal("expected validate error for invalid path")
	}
	if err := runRoot("init", "--target", dir, "--platform", "invalid"); err == nil {
		t.Fatal("expected init parse error")
	}
	if err := runRoot("scaffold", "skill", "Invalid_Name", "--output-dir", dir); err == nil {
		t.Fatal("expected scaffold invalid name error")
	}
}

func TestScaffoldSkillCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "secure-code-review",
		"--output-dir", dir,
		"--version", "0.1.0",
		"--description", "Security-focused code review skill",
		"--owner", "platform-engineering",
		"--platform", "codex",
		"--platform", "claude-code",
		"--platform", "gitlab-duo",
	); err != nil {
		t.Fatalf("scaffold skill failed: %v", err)
	}
	for _, rel := range []string{"SKILL.md", "skill.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("tests", "README.md")} {
		if _, err := os.Stat(filepath.Join(dir, "secure-code-review", rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	if err := runRoot("scaffold", "skill", "secure-code-review", "--output-dir", dir); err == nil {
		t.Fatal("expected existing scaffold error")
	}
	if err := runRoot("scaffold", "skill", "secure-code-review", "--output-dir", dir, "--force"); err != nil {
		t.Fatalf("expected force scaffold to succeed: %v", err)
	}

	dryRunDir := filepath.Join(dir, "dry-run")
	if err := runRoot("scaffold", "skill", "test-generator", "--output-dir", dryRunDir, "--dry-run"); err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dryRunDir, "test-generator")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write files, err=%v", err)
	}
}

func TestInjectedErrorPaths(t *testing.T) {
	// init abs error
	origAbs := cliAbsPath
	cliAbsPath = func(string) (string, error) { return "", errors.New("abs fail") }
	if err := runRoot("init", "--target", t.TempDir()); err == nil {
		t.Fatal("expected init abs error")
	}
	cliAbsPath = origAbs

	// init mkdir error
	origMkdir := cliMkdirAll
	cliMkdirAll = func(string, os.FileMode) error { return errors.New("mkdir fail") }
	if err := runRoot("init", "--target", t.TempDir()); err == nil {
		t.Fatal("expected init mkdir error")
	}
	cliMkdirAll = origMkdir

	// init build and dump errors
	origBuild := cliBuildInitialConfig
	cliBuildInitialConfig = func([]string, string, string, string, string, string) (*models.BakeConfig, error) {
		return nil, errors.New("build fail")
	}
	if err := runRoot("init", "--target", t.TempDir()); err == nil {
		t.Fatal("expected init build error")
	}
	cliBuildInitialConfig = origBuild

	origDump := cliDumpBakeFile
	cliDumpBakeFile = func(*models.BakeConfig, string) error { return errors.New("dump fail") }
	if err := runRoot("init", "--target", t.TempDir(), "--force"); err == nil {
		t.Fatal("expected init dump error")
	}
	cliDumpBakeFile = origDump

	// validate abs error
	origAbsValidate := cliAbsPathValidate
	cliAbsPathValidate = func(string) (string, error) { return "", errors.New("abs fail") }
	if err := runRoot("validate", "--target", "."); err == nil {
		t.Fatal("expected validate abs error")
	}
	cliAbsPathValidate = origAbsValidate

	// scaffold abs and write errors
	origAbsScaffold := cliAbsPathScaffold
	cliAbsPathScaffold = func(string) (string, error) { return "", errors.New("abs fail") }
	if err := runRoot("scaffold", "skill", "secure-code-review"); err == nil {
		t.Fatal("expected scaffold abs error")
	}
	cliAbsPathScaffold = origAbsScaffold

	origWriteSkill := cliWriteSkill
	cliWriteSkill = func(scaffold.SkillOptions) ([]scaffold.PlannedFile, error) {
		return nil, errors.New("scaffold fail")
	}
	if err := runRoot("scaffold", "skill", "secure-code-review", "--output-dir", t.TempDir()); err == nil {
		t.Fatal("expected scaffold write error")
	}
	cliWriteSkill = origWriteSkill
}

func TestBakeInjectedErrorPaths(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "Demo"); err != nil {
		t.Fatal(err)
	}

	origAbs := cliAbsPathBake
	cliAbsPathBake = func(string) (string, error) { return "", errors.New("abs fail") }
	if err := runRoot("bake", "default", "--target", dir); err == nil {
		t.Fatal("expected bake abs error")
	}
	cliAbsPathBake = origAbs

	origResolve := cliResolveTarget
	cliResolveTarget = func(*models.BakeConfig, string) (*models.TargetConfig, error) { return nil, errors.New("resolve fail") }
	if err := runRoot("bake", "default", "--target", dir); err == nil {
		t.Fatal("expected bake resolve error")
	}
	cliResolveTarget = origResolve

	origRender := cliRenderFiles
	cliRenderFiles = func(*models.BakeConfig, *models.TargetConfig) ([]models.RenderedFile, error) {
		return nil, errors.New("render fail")
	}
	if err := runRoot("bake", "default", "--target", dir); err == nil {
		t.Fatal("expected bake render error")
	}
	cliRenderFiles = origRender

	origLoadLock := cliLoadLockfile
	cliLoadLockfile = func(string) (map[string]any, error) { return nil, errors.New("lock fail") }
	if err := runRoot("bake", "default", "--target", dir); err == nil {
		t.Fatal("expected bake lock load error")
	}
	cliLoadLockfile = origLoadLock

	origRead := cliReadFile
	cliReadFile = func(string) ([]byte, error) { return nil, errors.New("read fail") }
	if err := runRoot("bake", "default", "--target", dir, "--plan"); err != nil {
		t.Fatalf("plan should continue on read errors, got %v", err)
	}
	cliReadFile = origRead

	// Write path errors
	origMkdir := cliMkdirAllBake
	cliMkdirAllBake = func(string, os.FileMode) error { return errors.New("mkdir fail") }
	if err := runRoot("bake", "default", "--target", dir, "--write"); err == nil {
		t.Fatal("expected bake mkdir error")
	}
	cliMkdirAllBake = origMkdir

	origWriteFile := cliWriteFile
	cliWriteFile = func(string, []byte, os.FileMode) error { return errors.New("write fail") }
	if err := runRoot("bake", "default", "--target", dir, "--write"); err == nil {
		t.Fatal("expected bake write file error")
	}
	cliWriteFile = origWriteFile

	origWriteLock := cliWriteLockfile
	cliWriteLockfile = func(string, []models.RenderedFile, string) error { return errors.New("write lock fail") }
	if err := runRoot("bake", "default", "--target", dir, "--write"); err == nil {
		t.Fatal("expected bake write lock error")
	}
	cliWriteLockfile = origWriteLock
}

func TestBakePlanCoversStateAndDiffBranches(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "Demo"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatal(err)
	}

	// Add stale state entries to trigger delete and platform fallback branches.
	lockPath := filepath.Join(dir, ".agentic-template.lock")
	lockData, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	var lock map[string]any
	if err := yaml.Unmarshal(lockData, &lock); err != nil {
		t.Fatal(err)
	}
	managed, _ := lock["managed_files"].([]any)
	if len(managed) == 0 {
		t.Fatal("expected managed files in lock")
	}
	first, _ := managed[0].(map[string]any)
	first["checksum"] = "sha256:deadbeef"
	lock["managed_files"] = append(managed,
		map[string]any{"path": "stale1", "platform": "codex", "source": "x", "checksum": "sha256:1"},
		map[string]any{"path": "stale2", "source": "x", "checksum": "sha256:2"},
	)
	encoded, err := yaml.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	agentsPath := filepath.Join(dir, ".agentic", "codex", "AGENTS.md")
	original, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(dir, ".agents", "skills", "security-reviewer", "SKILL.md")
	originalSkill, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	origRead := cliReadFile
	counts := map[string]int{}
	cliReadFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, filepath.Join(".agentic", "codex", "AGENTS.md")) {
			counts["agents"]++
			if counts["agents"] == 2 {
				return []byte("changed-content"), nil
			}
			if counts["agents"] == 3 {
				return nil, errors.New("second read error")
			}
			return original, nil
		}
		if strings.HasSuffix(path, filepath.Join(".agents", "skills", "security-reviewer", "SKILL.md")) {
			counts["skill"]++
			if counts["skill"] == 2 {
				return []byte("changed-content"), nil
			}
			return originalSkill, nil
		}
		return origRead(path)
	}
	defer func() { cliReadFile = origRead }()

	if err := runRoot("bake", "default", "--target", dir, "--plan"); err != nil {
		t.Fatalf("expected plan to succeed, got %v", err)
	}
}

func TestSkillIntegrationValidateAndCleanFlow(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "Demo"); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, ".skpm", "skills", "secure-code-review")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: secure-code-review\ndescription: ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lock := `skills:
  - name: secure-code-review
    version: v1.2.3
    source: registry
    compatible_with:
      - codex
    installed_paths:
      - .skpm/skills/secure-code-review
`
	if err := os.WriteFile(filepath.Join(dir, "agent-skills.lock"), []byte(lock), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "default", "--target", dir, "--skills-from", "agent-skills.lock", "--write"); err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(dir, ".agentic", "codex", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(agents), "secure-code-review v1.2.3") {
		t.Fatalf("expected skpm skill reference, got:\n%s", agents)
	}
	if err := runRoot("validate", "--target", dir, "--against-lock", "agent-skills.lock"); err != nil {
		t.Fatalf("validate against lock failed: %v", err)
	}
	if err := runRoot("list-targets", "--target", dir, "--with-skills"); err != nil {
		t.Fatalf("list-targets with skills failed: %v", err)
	}
	if err := runRoot("clean", "--target", dir, "--plan"); err != nil {
		t.Fatalf("clean plan failed: %v", err)
	}
	if err := runRoot("clean", "--target", dir, "--write"); err != nil {
		t.Fatalf("clean write failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agentic", "codex", "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("expected generated AGENTS.md removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
		t.Fatalf("clean should not remove skpm-managed skill file: %v", err)
	}
}

func TestHelpers(t *testing.T) {
	rendered := sortedRendered([]models.RenderedFile{
		{Destination: "b"},
		{Destination: "a"},
	})
	if len(rendered) != 2 || rendered[0].Destination != "a" {
		t.Fatalf("sortedRendered bad result: %#v", rendered)
	}

	// Cover helper branches directly.
	if got := checksumValue(map[string]any{"checksum": "x"}); got != "x" {
		t.Fatalf("checksumValue mismatch: %q", got)
	}
	if got := checksumValue(map[string]any{"checksum": 1}); got != "" {
		t.Fatalf("checksumValue expected empty, got %q", got)
	}

	sm := sortedMapKeys(map[string]map[string]any{"b": {}, "a": {}})
	if len(sm) != 2 || sm[0] != "a" {
		t.Fatalf("sortedMapKeys bad result: %#v", sm)
	}

	sk := sortedKeys(map[string]models.RenderedFile{"b": {Destination: "b"}, "a": {Destination: "a"}})
	if len(sk) != 2 || sk[0] != "a" {
		t.Fatalf("sortedKeys bad result, got %#v", sk)
	}

	short := unifiedDiff("a\n", "b\n", "x", true)
	if !strings.Contains(short, "a/x") || !strings.Contains(short, "b/x") {
		t.Fatalf("unexpected diff header: %q", short)
	}
	longA := strings.Repeat("a\n", 300)
	longB := strings.Repeat("b\n", 300)
	truncated := unifiedDiff(longA, longB, "x", true)
	if !strings.Contains(truncated, "... (diff truncated)") {
		t.Fatal("expected truncated diff marker")
	}
	full := unifiedDiff(longA, longB, "x", false)
	if strings.Contains(full, "... (diff truncated)") {
		t.Fatal("did not expect truncation")
	}
}

func TestInitWithSkillFlag(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex",
		"--project-name", "Demo",
		"--skill", "my-skill",
		"--skill", "another-skill",
	); err != nil {
		t.Fatalf("init with --skill failed: %v", err)
	}
	bakeBytes, err := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	bakeContent := string(bakeBytes)
	if !strings.Contains(bakeContent, "my-skill") {
		t.Fatalf("expected my-skill in bakefile:\n%s", bakeContent)
	}
	if !strings.Contains(bakeContent, "another-skill") {
		t.Fatalf("expected another-skill in bakefile:\n%s", bakeContent)
	}
	if !strings.Contains(bakeContent, "skill_sources") {
		t.Fatalf("expected skill_sources block in bakefile:\n%s", bakeContent)
	}
}

func TestInitGeneratesSkillSourcesBlock(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex,gitlab-duo", "--project-name", "Demo"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	bakeBytes, err := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	bakeContent := string(bakeBytes)
	if !strings.Contains(bakeContent, "skill_sources") {
		t.Fatalf("expected skill_sources block in generated bakefile:\n%s", bakeContent)
	}
}

func TestScaffoldSkillsCommand(t *testing.T) {
	dir := t.TempDir()
	// Create a bakefile with skill_sources configured.
	bakeContent := `version: "1"
skill_sources:
  output_dir: skills
  defaults:
    version: 0.1.0
    owner: platform-engineering
    license: MIT
    compatible_with:
      - codex
  skills:
    - name: my-skill
      description: My great skill.
    - name: another-skill
      compatible_with:
        - claude-code
targets:
  default:
    platforms:
      - codex
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("scaffold", "skills", "--target", dir); err != nil {
		t.Fatalf("scaffold skills failed: %v", err)
	}

	for _, skillName := range []string{"my-skill", "another-skill"} {
		for _, rel := range []string{"SKILL.md", "skill.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("tests", "README.md")} {
			if _, err := os.Stat(filepath.Join(dir, "skills", skillName, rel)); err != nil {
				t.Fatalf("missing %s/%s/%s: %v", skillName, rel, skillName, err)
			}
		}
	}

	// Second run: all files exist, no error.
	if err := runRoot("scaffold", "skills", "--target", dir); err != nil {
		t.Fatalf("scaffold skills second run (skip-existing) failed: %v", err)
	}

	// Force run: succeeds with --force.
	if err := runRoot("scaffold", "skills", "--target", dir, "--force"); err != nil {
		t.Fatalf("scaffold skills --force failed: %v", err)
	}
}

func TestScaffoldSkillsDryRun(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `version: "1"
skill_sources:
  output_dir: skills
  skills:
    - name: dry-skill
targets:
  default:
    platforms:
      - codex
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("scaffold", "skills", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("scaffold skills --dry-run failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "skills", "dry-skill")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not write files, err=%v", err)
	}
}

func TestScaffoldSkillsNoBakefile(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skills", "--target", dir); err == nil {
		t.Fatal("expected error when no bakefile")
	}
}

func TestScaffoldSkillsEmpty(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `version: "1"
skill_sources:
  output_dir: skills
  skills: []
targets:
  default:
    platforms:
      - codex
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	// Should succeed gracefully with a "nothing to scaffold" message.
	if err := runRoot("scaffold", "skills", "--target", dir); err != nil {
		t.Fatalf("scaffold skills with empty skills list: %v", err)
	}
}

func TestScaffoldSkillsDuplicateNames(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `version: "1"
skill_sources:
  output_dir: skills
  skills:
    - name: my-skill
    - name: my-skill
targets:
  default:
    platforms:
      - codex
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("scaffold", "skills", "--target", dir); err == nil {
		t.Fatal("expected error for duplicate skill name")
	}
}

func TestBakeWithSkillSources(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `version: "1"
skill_sources:
  defaults:
    version: 0.1.0
    owner: platform-engineering
    license: MIT
    compatible_with:
      - codex
targets:
  default:
    platforms:
      - codex
    skills:
      - test-skill
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// bake --write should scaffold the skill and render the platform skill file.
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake write with skill_sources failed: %v", err)
	}

	// Canonical skill source created directly in .agents/skills/.
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "test-skill", "SKILL.md")); err != nil {
		t.Fatalf("skill source SKILL.md not created: %v", err)
	}

	// bake --plan should also succeed.
	if err := runRoot("bake", "default", "--target", dir, "--plan"); err != nil {
		t.Fatalf("bake plan with skill_sources failed: %v", err)
	}
}

func TestBakeCanonicalSkillPlatformOutputs(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `version: "1"
targets:
  default:
    platforms:
      - codex
      - claude-code
      - gitlab-duo
    skills:
      - multi-platform-skill
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake write failed: %v", err)
	}

	// Canonical source and codex/gitlab-duo: .agents/skills/<name>/SKILL.md
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "multi-platform-skill", "SKILL.md")); err != nil {
		t.Fatalf("canonical skill source missing: %v", err)
	}
	// Claude Code gets its own copy: .claude/skills/<name>/SKILL.md
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", "multi-platform-skill", "SKILL.md")); err != nil {
		t.Fatalf("claude-code platform skill missing: %v", err)
	}
}

func TestScaffoldSkillCommandNextSteps(t *testing.T) {
	dir := t.TempDir()
	// Just verify the command succeeds and produces the skill directory.
	if err := runRoot("scaffold", "skill", "policy-reviewer",
		"--output-dir", dir,
		"--platform", "codex",
	); err != nil {
		t.Fatalf("scaffold skill failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "policy-reviewer", "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md not created: %v", err)
	}
}

func TestScaffoldSkillsAbsError(t *testing.T) {
	origAbs := cliAbsPathScaffold
	cliAbsPathScaffold = func(string) (string, error) { return "", errors.New("abs fail") }
	t.Cleanup(func() { cliAbsPathScaffold = origAbs })
	if err := runRoot("scaffold", "skills", "--target", "."); err == nil {
		t.Fatal("expected abs error")
	}
}

func TestScaffoldSkillsWriteSkillSafeError(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `version: "1"
skill_sources:
  output_dir: skills
  skills:
    - name: fail-skill
targets:
  default:
    platforms:
      - codex
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := cliWriteSkillSafe
	cliWriteSkillSafe = func(scaffold.SkillOptions) (*scaffold.SkillWriteResult, error) {
		return nil, errors.New("safe write fail")
	}
	t.Cleanup(func() { cliWriteSkillSafe = orig })
	if err := runRoot("scaffold", "skills", "--target", dir); err == nil {
		t.Fatal("expected safe write error")
	}
}

func TestExecuteErrorSubprocess(t *testing.T) {
	if os.Getenv("SKCR_EXECUTE_ERR") == "1" {
		orig := os.Args
		defer func() { os.Args = orig }()
		os.Args = []string{"skcr", "does-not-exist"}
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run", "TestExecuteErrorSubprocess")
	cmd.Env = append(os.Environ(), "SKCR_EXECUTE_ERR=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected subprocess to fail via os.Exit")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got err=%v", err)
	}
}
