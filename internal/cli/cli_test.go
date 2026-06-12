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
	"github.com/domehahn/skcr/internal/renderer"
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

func TestCompatibilityCommandsAndBakeUseEvidence(t *testing.T) {
	dir := t.TempDir()
	evidence := filepath.Join(dir, "docs", "compat", "codex-0.51.0.md")
	if err := os.MkdirAll(filepath.Dir(evidence), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidence, []byte("# Codex evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "Demo"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := runRoot("compatibility", "set", "codex", "--target", dir, "--min-version", "0.51.0", "--evidence", "docs/compat/codex-0.51.0.md", "--validated", "2026-06-12"); err != nil {
		t.Fatalf("compatibility set failed: %v", err)
	}
	if err := runRoot("compatibility", "check", "--target", dir); err != nil {
		t.Fatalf("compatibility check failed: %v", err)
	}
	if err := runRoot("compatibility", "matrix", "--target", dir, "--json"); err != nil {
		t.Fatalf("compatibility matrix failed: %v", err)
	}
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "security-reviewer", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `codex: "0.51.0"`) {
		t.Fatalf("expected baked skill to use verified codex min version: %s", content)
	}
	skillDir := filepath.Join(dir, ".agents", "skills", "security-reviewer")
	versionFile, err := os.ReadFile(filepath.Join(skillDir, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(versionFile)) != "1.0.0" {
		t.Fatalf("VERSION not synchronized with SKILL.md: %q", versionFile)
	}
	skillYAML, err := os.ReadFile(filepath.Join(skillDir, "skill.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(skillYAML), "version: 1.0.0") {
		t.Fatalf("skill.yaml not synchronized with SKILL.md: %s", skillYAML)
	}
	changelog, err := os.ReadFile(filepath.Join(skillDir, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(changelog), "## 1.0.0 - ") {
		t.Fatalf("CHANGELOG.md missing versioned entry with date: %s", changelog)
	}
	if err := runRoot("version", "check", skillDir); err != nil {
		t.Fatalf("version check should pass for baked skill artifacts: %v", err)
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
	for _, rel := range []string{"SKILL.md", "skill.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("scripts", "README.md"), filepath.Join("references", "README.md"), filepath.Join("assets", "README.md"), filepath.Join("tests", "README.md")} {
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

	origRenderWithOpts := cliRenderWithOpts
	cliRenderWithOpts = func(*models.BakeConfig, *models.TargetConfig, renderer.Options) ([]models.RenderedFile, error) {
		return nil, errors.New("render fail")
	}
	if err := runRoot("bake", "default", "--target", dir); err == nil {
		t.Fatal("expected bake render error")
	}
	cliRenderWithOpts = origRenderWithOpts

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
		for _, rel := range []string{"SKILL.md", "skill.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("scripts", "README.md"), filepath.Join("references", "README.md"), filepath.Join("assets", "README.md"), filepath.Join("tests", "README.md")} {
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

func TestCompatibilityCheckValidation(t *testing.T) {
	dir := t.TempDir()

	// No agentic.compatibility.yaml → no verified entries, no error.
	if err := runRoot("compatibility", "check", "--target", dir); err != nil {
		t.Fatalf("check with no evidence file should succeed: %v", err)
	}

	// Write invalid evidence: verified with unknown version → should fail.
	badEvidence := `platforms:
  - name: codex
    min_version: "unknown"
    status: "verified"
    source: "test"
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.compatibility.yaml"), []byte(badEvidence), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("compatibility", "check", "--target", dir); err == nil {
		t.Fatal("expected error for verified entry with unknown version")
	}

	// Valid evidence: verified with concrete version and existing evidence file.
	evFile := filepath.Join(dir, "evidence.md")
	if err := os.WriteFile(evFile, []byte("# evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	goodEvidence := `platforms:
  - name: codex
    min_version: "0.51.0"
    status: "verified"
    source: "test"
    evidence: "evidence.md"
    validated: "2026-06-10"
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.compatibility.yaml"), []byte(goodEvidence), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("compatibility", "check", "--target", dir); err != nil {
		t.Fatalf("expected no error for valid verified entry: %v", err)
	}
}

func TestDoctorHelpers(t *testing.T) {
	// doctorExitCode: no errors → nil.
	findings := []doctorFinding{
		{level: "ok", check: "bakefile", msg: "ok"},
		{level: "warn", check: "toolchain", msg: "skpm not found"},
	}
	if err := doctorExitCode(findings); err != nil {
		t.Fatalf("expected nil for ok+warn findings, got %v", err)
	}

	// doctorExitCode: with error → non-nil.
	findings = append(findings, doctorFinding{level: "error", check: "targets", msg: "no targets"})
	if err := doctorExitCode(findings); err == nil {
		t.Fatal("expected error when findings contain level=error")
	}

	// checkSkillMDFrontmatter: valid content from a real scaffolded skill → empty string.
	{
		tmpDir := t.TempDir()
		if _, err := scaffold.WriteSkill(scaffold.SkillOptions{
			Name:      "security-reviewer",
			OutputDir: tmpDir,
			Platforms: []string{"codex"},
		}); err != nil {
			t.Fatal(err)
		}
		skillMDBytes, err := os.ReadFile(filepath.Join(tmpDir, "security-reviewer", "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if msg := checkSkillMDFrontmatter(string(skillMDBytes)); msg != "" {
			t.Fatalf("valid scaffolded content: expected empty, got %q", msg)
		}
	}

	// checkSkillMDFrontmatter: missing required fields → non-empty error.
	if msg := checkSkillMDFrontmatter("# No frontmatter\n"); msg == "" {
		t.Fatal("invalid content: expected error message")
	}
}

func TestStripFrontmatter(t *testing.T) {
	// No frontmatter: returned as-is.
	plain := "# Just markdown\nsome text"
	if got := stripFrontmatter(plain); got != plain {
		t.Fatalf("no frontmatter: got %q", got)
	}

	// With frontmatter: strips it.
	withFM := "---\nname: test\n---\n# Body\nsome text"
	got := stripFrontmatter(withFM)
	if strings.Contains(got, "name: test") {
		t.Fatalf("frontmatter not stripped: %q", got)
	}
	if !strings.Contains(got, "# Body") {
		t.Fatalf("body missing after strip: %q", got)
	}

	// Unclosed frontmatter: returned as-is.
	unclosed := "---\nname: test\n# no closing"
	if got := stripFrontmatter(unclosed); got != unclosed {
		t.Fatalf("unclosed frontmatter: expected passthrough, got %q", got)
	}
}

func TestShortDirLabel(t *testing.T) {
	cases := []struct{ dir, want string }{
		{".agents/skills", "agents"},
		{"agents/skills", "agents"},
		{".vscode/skills", "vscode"},
		{"src/skills", "src"},
	}
	for _, tc := range cases {
		if got := shortDirLabel(tc.dir); got != tc.want {
			t.Errorf("shortDirLabel(%q) = %q, want %q", tc.dir, got, tc.want)
		}
	}
}

func TestAllPlatformBaseDirs(t *testing.T) {
	cfg := &models.BakeConfig{
		Targets: map[string]*models.TargetConfig{
			"codex": {Platforms: []string{"codex", "claude-code"}},
		},
	}
	dirs := allPlatformBaseDirs(cfg)
	if len(dirs) == 0 {
		t.Fatal("expected at least one dir")
	}
	// .agents/skills must always be present.
	found := false
	for _, d := range dirs {
		if d == ".agents/skills" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf(".agents/skills not in dirs: %v", dirs)
	}
	// No duplicates.
	seen := map[string]struct{}{}
	for _, d := range dirs {
		if _, dup := seen[d]; dup {
			t.Fatalf("duplicate dir %q in allPlatformBaseDirs result: %v", d, dirs)
		}
		seen[d] = struct{}{}
	}
}

func TestAddSkillCommand(t *testing.T) {
	dir := t.TempDir()
	// Init a project first.
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}

	// Add a new skill — should succeed and scaffold files.
	if err := runRoot("add", "skill", "my-new-skill", "--target", dir); err != nil {
		t.Fatalf("add skill failed: %v", err)
	}

	// Check the skill was added to agentic.bake.yaml.
	bakeData, err := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bakeData), "my-new-skill") {
		t.Fatalf("skill not added to bakefile: %s", bakeData)
	}

	// Check scaffold files were created.
	skillDir := filepath.Join(dir, ".agents", "skills", "my-new-skill")
	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md not scaffolded: %v", err)
	}

	// Add the same skill again: should succeed (already present message, no error).
	if err := runRoot("add", "skill", "my-new-skill", "--target", dir); err != nil {
		t.Fatalf("add skill twice should not error: %v", err)
	}

	// Add with --no-scaffold: bakefile updated but no new directory.
	if err := runRoot("add", "skill", "another-skill", "--target", dir, "--no-scaffold"); err != nil {
		t.Fatalf("add skill --no-scaffold failed: %v", err)
	}
	bakeData2, err := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bakeData2), "another-skill") {
		t.Fatalf("skill not added to bakefile (--no-scaffold): %s", bakeData2)
	}

	// Invalid skill name → error.
	if err := runRoot("add", "skill", "Invalid_Name", "--target", dir); err == nil {
		t.Fatal("expected error for invalid skill name")
	}
}

func TestRenameSkillCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "old-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Dry-run: bakefile unchanged, directories unchanged.
	if err := runRoot("rename", "skill", "old-skill", "new-skill", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("dry-run rename failed: %v", err)
	}
	bakeData, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bakeData), "old-skill") {
		t.Fatal("dry-run should not modify bakefile")
	}

	// Real rename.
	if err := runRoot("rename", "skill", "old-skill", "new-skill", "--target", dir); err != nil {
		t.Fatalf("rename failed: %v", err)
	}
	bakeData, _ = os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if strings.Contains(string(bakeData), "old-skill") {
		t.Fatal("old skill name still in bakefile after rename")
	}
	if !strings.Contains(string(bakeData), "new-skill") {
		t.Fatal("new skill name not in bakefile after rename")
	}

	// Directory moved.
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "new-skill", "SKILL.md")); err != nil {
		t.Fatalf("new skill directory not found after rename: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "old-skill")); !os.IsNotExist(err) {
		t.Fatal("old skill directory still exists after rename")
	}

	// Rename non-existent skill → error.
	if err := runRoot("rename", "skill", "old-skill", "other-skill", "--target", dir); err == nil {
		t.Fatal("expected error when renaming non-existent skill")
	}

	// Rename to identical name → error.
	if err := runRoot("rename", "skill", "new-skill", "new-skill", "--target", dir); err == nil {
		t.Fatal("expected error when old and new name are the same")
	}

	// Rename to invalid name → error.
	if err := runRoot("rename", "skill", "new-skill", "Invalid_Name", "--target", dir); err == nil {
		t.Fatal("expected error for invalid new name")
	}
}

func TestRenameSkillConflict(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "skill-a", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "skill-b", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Rename skill-a to skill-b — skill-b already exists in bakefile → error.
	if err := runRoot("rename", "skill", "skill-a", "skill-b", "--target", dir); err == nil {
		t.Fatal("expected error when new name already exists in a target")
	}

	// Directory conflict path: add skill-c (not in bakefile), but create its destination
	// directory manually so the os.Rename would conflict. We do this by scaffolding
	// skill-c, then adding skill-d to the bakefile without scaffolding, then manually
	// creating skill-d's directory to produce the conflict.
	if err := runRoot("add", "skill", "skill-c", "--target", dir); err != nil {
		t.Fatal(err)
	}
	// Manually pre-create skill-d directory (destination of future rename).
	if err := os.MkdirAll(filepath.Join(dir, ".agents", "skills", "skill-d"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Rename skill-c → skill-d: bakefile rename succeeds, but directory move is
	// skipped because skill-d already exists on disk. Command must not return error.
	if err := runRoot("rename", "skill", "skill-c", "skill-d", "--target", dir); err != nil {
		t.Fatalf("rename with dir conflict should not return error (conflict skipped): %v", err)
	}
	// skill-c directory must still exist (was not moved).
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "skill-c")); err != nil {
		t.Fatal("skill-c directory should be preserved when rename destination conflicts")
	}
}

func TestRemoveSkillCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "to-remove", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Dry-run with --delete-dirs: nothing actually deleted.
	if err := runRoot("remove", "skill", "to-remove", "--target", dir, "--dry-run", "--delete-dirs"); err != nil {
		t.Fatalf("dry-run remove failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "to-remove")); err != nil {
		t.Fatal("dry-run should not delete directories")
	}
	bakeData, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bakeData), "to-remove") {
		t.Fatal("dry-run should not modify bakefile")
	}

	// Real remove without --delete-dirs: bakefile updated, directories kept.
	if err := runRoot("remove", "skill", "to-remove", "--target", dir); err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	bakeData, _ = os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if strings.Contains(string(bakeData), "to-remove") {
		t.Fatal("skill still in bakefile after remove")
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "to-remove")); err != nil {
		t.Fatal("directory should be preserved when --delete-dirs not given")
	}

	// Remove with --delete-dirs on already-removed-from-bakefile skill: directories deleted.
	if err := runRoot("add", "skill", "to-delete", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("remove", "skill", "to-delete", "--target", dir, "--delete-dirs"); err != nil {
		t.Fatalf("remove --delete-dirs failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "to-delete")); !os.IsNotExist(err) {
		t.Fatal("directory should be deleted when --delete-dirs given")
	}

	// Remove skill not in any target: no error, message only.
	if err := runRoot("remove", "skill", "nonexistent", "--target", dir); err != nil {
		t.Fatalf("remove of absent skill should not error: %v", err)
	}
}

func TestRemoveSkillInTarget(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "targeted-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Load bakefile to discover target names.
	bakeData, err := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// The init command creates at least a "codex" target.
	if !strings.Contains(string(bakeData), "targeted-skill") {
		t.Fatal("skill not in bakefile before targeted remove")
	}

	// Remove only from the "codex" target; other targets should keep it.
	if err := runRoot("remove", "skill", "targeted-skill", "--target", dir, "--in-target", "codex"); err != nil {
		t.Fatalf("remove --in-target codex failed: %v", err)
	}
	bakeData2, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	// skill should be gone from codex section — but may remain in other targets.
	// Verify bakefile changed at all.
	if string(bakeData) == string(bakeData2) {
		t.Fatal("bakefile unchanged after remove --in-target")
	}

	// Remove from a non-existent target → error.
	if err := runRoot("remove", "skill", "targeted-skill", "--target", dir, "--in-target", "no-such-target"); err == nil {
		t.Fatal("expected error for non-existent target")
	}
}

func TestExportCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "export-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Export to file.
	outFile := filepath.Join(dir, "export.md")
	if err := runRoot("export", "--target", dir, "--out", outFile); err != nil {
		t.Fatalf("export failed: %v", err)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("export file not created: %v", err)
	}
	if !strings.Contains(string(data), "# Agent Skills") {
		t.Fatalf("export missing header: %s", data[:min(200, len(data))])
	}

	// --keep-frontmatter includes YAML block.
	outFM := filepath.Join(dir, "export-fm.md")
	if err := runRoot("export", "--target", dir, "--out", outFM, "--keep-frontmatter"); err != nil {
		t.Fatalf("export --keep-frontmatter failed: %v", err)
	}
	dataFM, _ := os.ReadFile(outFM)
	if !strings.Contains(string(dataFM), "name:") {
		t.Fatalf("--keep-frontmatter export should contain YAML name: field")
	}

	// --skill filter: only that skill exported.
	outSkill := filepath.Join(dir, "export-skill.md")
	if err := runRoot("export", "--target", dir, "--out", outSkill, "--skill", "export-skill"); err != nil {
		t.Fatalf("export --skill failed: %v", err)
	}

	// No skills → error.
	emptyDir := t.TempDir()
	if err := runRoot("init", "--target", emptyDir, "--platform", "codex", "--project-name", "Empty"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("export", "--target", emptyDir); err == nil {
		t.Fatal("expected error when no skill SKILL.md files found")
	}
}

func TestStatusCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "status-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Status with a scaffolded skill.
	if err := runRoot("status", "--target", dir); err != nil {
		t.Fatalf("status failed: %v", err)
	}

	// Status on empty project (no skills).
	emptyDir := t.TempDir()
	if err := runRoot("init", "--target", emptyDir, "--platform", "codex", "--project-name", "Empty"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("status", "--target", emptyDir); err != nil {
		t.Fatalf("status on empty project failed: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestDoctorEndToEnd(t *testing.T) {
	dir := t.TempDir()

	// No bakefile → error.
	if err := runRoot("doctor", "--target", dir); err == nil {
		t.Fatal("expected error when bakefile is missing")
	}

	// Init + scaffold skills → doctor should pass cleanly.
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "DoctorTest"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "my-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("doctor", "--target", dir); err != nil {
		t.Fatalf("doctor should pass on clean project: %v", err)
	}

	// Corrupt VERSION → doctor should report an error.
	versionPath := filepath.Join(dir, ".agents", "skills", "my-skill", "VERSION")
	if err := os.WriteFile(versionPath, []byte("not-semver\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("doctor", "--target", dir); err == nil {
		t.Fatal("expected doctor error for invalid VERSION")
	}

	// Restore valid VERSION.
	if err := os.WriteFile(versionPath, []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Inject a since > last_modified violation by replacing the since field inline.
	skillMDPath := filepath.Join(dir, ".agents", "skills", "my-skill", "SKILL.md")
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		t.Fatal(err)
	}
	patched := strings.ReplaceAll(string(data), "since: \"2026-06-12\"", "since: \"9999-01-01\"")
	if patched == string(data) {
		// since field uses today's actual date — replace whatever value is there
		lines := strings.Split(string(data), "\n")
		for i, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "since:") {
				lines[i] = "since: \"9999-01-01\""
				break
			}
		}
		patched = strings.Join(lines, "\n")
	}
	if err := os.WriteFile(skillMDPath, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("doctor", "--target", dir); err == nil {
		t.Fatal("expected doctor error for since > last_modified")
	}
}

func TestValidateRenderedSkillDates(t *testing.T) {
	// No violations → empty slice.
	clean := []models.RenderedFile{
		{Destination: ".agents/skills/my-skill/SKILL.md", Content: "---\nsince: \"2026-01-01\"\nlast_modified: \"2026-06-12\"\n---\n"},
		{Destination: ".agents/skills/my-skill/skill.yaml", Content: "name: my-skill\n"},
	}
	if errs := validateRenderedSkillDates(clean); len(errs) != 0 {
		t.Fatalf("expected no errors for clean files, got: %v", errs)
	}

	// since > last_modified in a SKILL.md → one error.
	bad := []models.RenderedFile{
		{Destination: ".agents/skills/bad-skill/SKILL.md", Content: "---\nsince: \"9999-01-01\"\nlast_modified: \"2026-06-12\"\n---\n"},
	}
	if errs := validateRenderedSkillDates(bad); len(errs) != 1 {
		t.Fatalf("expected one error for violated invariant, got: %v", errs)
	}

	// Non-SKILL.md files are skipped even if they have the right field names.
	nonSkill := []models.RenderedFile{
		{Destination: ".agents/skills/my-skill/README.md", Content: "---\nsince: \"9999-01-01\"\nlast_modified: \"2026-06-12\"\n---\n"},
	}
	if errs := validateRenderedSkillDates(nonSkill); len(errs) != 0 {
		t.Fatalf("expected no errors for non-SKILL.md file, got: %v", errs)
	}

	// Empty content skipped gracefully.
	empty := []models.RenderedFile{
		{Destination: ".agents/skills/my-skill/SKILL.md", Content: ""},
	}
	if errs := validateRenderedSkillDates(empty); len(errs) != 0 {
		t.Fatalf("expected no errors for empty content, got: %v", errs)
	}
}

func TestDoctorSkillYAMLVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "version-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Read the actual version from skill.yaml, then tamper it.
	yamlPath := filepath.Join(dir, ".agents", "skills", "version-skill", "skill.yaml")
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	// Replace whichever version is present with a clearly divergent one.
	patched := yamlData
	for _, old := range []string{"version: 0.1.0", `version: "0.1.0"`, "version: 1.0.0", `version: "1.0.0"`} {
		replaced := strings.ReplaceAll(string(yamlData), old, "version: 9.9.9")
		if replaced != string(yamlData) {
			patched = []byte(replaced)
			break
		}
	}
	if string(patched) == string(yamlData) {
		t.Skip("could not locate version field in skill.yaml to tamper")
	}
	if err := os.WriteFile(yamlPath, patched, 0o644); err != nil {
		t.Fatal(err)
	}

	// doctor must return an error for the version mismatch.
	if err := runRoot("doctor", "--target", dir); err == nil {
		t.Fatal("expected doctor error for skill.yaml version mismatch")
	}
}

func TestSyncCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "sync-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Sync with no extra platform dirs: nothing to sync, should succeed.
	if err := runRoot("sync", "--target", dir); err != nil {
		t.Fatalf("sync on single-platform project failed: %v", err)
	}

	// Dry-run.
	if err := runRoot("sync", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("sync --dry-run failed: %v", err)
	}

	// Skill filter on non-existent skill: should succeed (skip + summary).
	if err := runRoot("sync", "--target", dir, "--skill", "nonexistent"); err != nil {
		t.Fatalf("sync --skill nonexistent failed: %v", err)
	}
}

func TestAddSkillInTarget(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}

	// Add only to the "codex" target.
	if err := runRoot("add", "skill", "targeted-new-skill", "--target", dir, "--in-target", "codex"); err != nil {
		t.Fatalf("add skill --in-target codex failed: %v", err)
	}
	bakeData, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bakeData), "targeted-new-skill") {
		t.Fatal("skill not added to bakefile")
	}

	// Add to a non-existent target → error.
	if err := runRoot("add", "skill", "another-skill", "--target", dir, "--in-target", "no-such-target"); err == nil {
		t.Fatal("expected error for non-existent target in --in-target")
	}
}

func TestListSkillsCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "list-skill-a", "--target", dir, "--no-scaffold"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "list-skill-b", "--target", dir, "--no-scaffold"); err != nil {
		t.Fatal(err)
	}

	// Default: lists all skills.
	if err := runRoot("list", "skills", "--target", dir); err != nil {
		t.Fatalf("list skills failed: %v", err)
	}

	// --with-targets.
	if err := runRoot("list", "skills", "--target", dir, "--with-targets"); err != nil {
		t.Fatalf("list skills --with-targets failed: %v", err)
	}

	// --in-target filter.
	if err := runRoot("list", "skills", "--target", dir, "--in-target", "codex"); err != nil {
		t.Fatalf("list skills --in-target codex failed: %v", err)
	}

	// --in-target with no matching skills prints a message but no error.
	emptyDir := t.TempDir()
	if err := runRoot("init", "--target", emptyDir, "--platform", "codex", "--project-name", "Empty"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("list", "skills", "--target", emptyDir); err != nil {
		t.Fatalf("list skills on empty project failed: %v", err)
	}
}

func TestListTargetsCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("list-targets", "--target", dir); err != nil {
		t.Fatalf("list-targets failed: %v", err)
	}

	// Missing bakefile → error.
	if err := runRoot("list-targets", "--target", t.TempDir()); err == nil {
		t.Fatal("expected error for missing bakefile")
	}
}

func TestCleanCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "--write", "--target", dir); err != nil {
		t.Fatalf("bake --write failed: %v", err)
	}

	// Plan mode (default when neither --plan nor --write given): lists files, no deletion.
	if err := runRoot("clean", "--target", dir); err != nil {
		t.Fatalf("clean (default plan) failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatal("clean default plan should not delete files")
	}

	// Explicit --plan: same behaviour.
	if err := runRoot("clean", "--target", dir, "--plan"); err != nil {
		t.Fatalf("clean --plan failed: %v", err)
	}

	// --write: removes managed files.
	if err := runRoot("clean", "--write", "--target", dir); err != nil {
		t.Fatalf("clean --write failed: %v", err)
	}
	// At least AGENTS.md should be gone.
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatal("AGENTS.md should be removed after clean --write")
	}

	// Second --write: nothing to remove, should succeed with empty-lock message.
	if err := runRoot("clean", "--write", "--target", dir); err != nil {
		t.Fatalf("clean --write on already-clean dir failed: %v", err)
	}
}

func TestValidateCommand(t *testing.T) {
	dir := t.TempDir()

	// Missing bakefile → error.
	if err := runRoot("validate", "--target", dir); err == nil {
		t.Fatal("expected error for missing bakefile")
	}

	// Init + bake: validate should pass.
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "--write", "--target", dir); err != nil {
		t.Fatalf("bake --write failed: %v", err)
	}
	if err := runRoot("validate", "--target", dir); err != nil {
		t.Fatalf("validate on baked project failed: %v", err)
	}

	// --ci flag: should also succeed.
	if err := runRoot("validate", "--target", dir, "--ci"); err != nil {
		t.Fatalf("validate --ci failed: %v", err)
	}
}

func TestExportInTarget(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "skill-alpha", "--target", dir); err != nil {
		t.Fatal(err)
	}

	// Export with --in-target codex: should succeed and contain the skill.
	outFile := filepath.Join(dir, "export-in-target.md")
	if err := runRoot("export", "--target", dir, "--out", outFile, "--in-target", "codex"); err != nil {
		t.Fatalf("export --in-target codex failed: %v", err)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("export file not created: %v", err)
	}
	if !strings.Contains(string(data), "# Agent Skills") {
		t.Fatalf("export output missing header: %s", string(data)[:min(200, len(data))])
	}

	// --in-target on a target that has no skills → error (no skills to export).
	emptyDir := t.TempDir()
	if err := runRoot("init", "--target", emptyDir, "--platform", "codex", "--project-name", "Empty"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("export", "--target", emptyDir, "--in-target", "codex"); err == nil {
		t.Fatal("expected error when target has no skills")
	}
}

func TestIsSDLCSkill(t *testing.T) {
	// All 18 canonical SDLC skill names must return true.
	for _, name := range scaffold.SDLCSkillNames {
		if !isSDLCSkill(name) {
			t.Errorf("isSDLCSkill(%q) = false, want true", name)
		}
	}

	// Custom / non-SDLC names must return false.
	for _, name := range []string{"my-custom-skill", "unknown", "", "security_reviewer"} {
		if isSDLCSkill(name) {
			t.Errorf("isSDLCSkill(%q) = true, want false", name)
		}
	}
}
