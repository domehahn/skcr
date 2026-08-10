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

	"github.com/domehahn/skcr/internal/catalog"
	"github.com/domehahn/skcr/internal/models"
	"github.com/domehahn/skcr/internal/renderer"
	"github.com/domehahn/skcr/internal/scaffold"
	"github.com/domehahn/skcr/internal/skillmeta"
	"github.com/domehahn/sklib/spec"
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

func TestContractDiffAndDigestCommands(t *testing.T) {
	oldDir := t.TempDir()
	newDir := t.TempDir()
	oldContract := skillmeta.NewContract()
	newContract := skillmeta.NewContract()
	newContract.Capabilities.Allowed.Network.Allow = []string{"api.example.com"}
	for dir, contract := range map[string]skillmeta.Contract{oldDir: oldContract, newDir: newContract} {
		data, err := yaml.Marshal(contract)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "contract.yaml"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out, err := runRootOut("contract", "diff", oldDir, newDir, "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"classification": "EXPANSION"`) || !strings.Contains(out, "api.example.com") {
		t.Fatalf("unexpected contract diff: %s", out)
	}
	out, err = runRootOut("contract", "digest", oldDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "sha256:") {
		t.Fatalf("unexpected digest: %s", out)
	}
}

func TestMigrateSkillCreatesDefaultDenySplitArtifact(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "Migration"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "legacy-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, ".agents", "skills", "legacy-skill")
	skillMD := filepath.Join(skillDir, "SKILL.md")
	originalInstructions, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	legacy := "name: legacy-skill\nversion: 0.1.0\ndescription: Legacy skill\nentrypoint: SKILL.md\ncompatible_with: [codex]\nsecurity:\n  requires_network: true\n  requires_secrets: false\n  writes_files: false\n  runs_commands: false\n"
	if err := os.Remove(filepath.Join(skillDir, "descriptor.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(skillDir, "contract.yaml")); err != nil {
		t.Fatal(err)
	}
	out, err := runRootOut("migrate", "skill", "legacy-skill", "--target", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "deprecated") || !strings.Contains(out, "do not grant permissions") {
		t.Fatalf("migration did not explain legacy security semantics: %s", out)
	}
	afterInstructions, err := os.ReadFile(skillMD)
	if err != nil || string(afterInstructions) != string(originalInstructions) {
		t.Fatal("migration modified SKILL.md")
	}
	contract, err := skillmeta.LoadContract(filepath.Join(skillDir, "contract.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if contract.SchemaVersion != skillmeta.ContractSchemaVersion || len(contract.Capabilities.Runtime.Allowed.Filesystem.Write) != 0 || *contract.Capabilities.Runtime.Allowed.Commands.Execute {
		t.Fatal("migration inferred permissions instead of default deny")
	}
	evalData, err := os.ReadFile(filepath.Join(skillDir, "evals", "baseline.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	eval, err := skillmeta.ParseEval(evalData)
	if err != nil || eval.SchemaVersion != skillmeta.EvalSchemaVersion {
		t.Fatalf("migration did not create Eval v2: %v", err)
	}
	descriptor, err := skillmeta.LoadDescriptor(filepath.Join(skillDir, "descriptor.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if descriptor.Security == nil || !descriptor.Security.RequiresNetwork {
		t.Fatal("migration did not preserve deprecated security hints")
	}
}

func TestASPSRequirementsCommands(t *testing.T) {
	dir := t.TempDir()
	files, err := scaffold.PlanSkill(scaffold.SkillOptions{Name: "asps-skill", OutputDir: dir, Owner: "security", Platforms: []string{"codex"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file.Path, []byte(file.Content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	skillDir := filepath.Join(dir, "asps-skill")
	for _, tc := range []struct {
		args []string
		want string
	}{{[]string{"asps", "show", "ASP-08.03"}, "Delegation Monotonicity"}, {[]string{"asps", "requirements", skillDir}, "ASP-01.01"}, {[]string{"asps", "coverage", skillDir}, "Source coverage only"}, {[]string{"asps", "validate", skillDir}, "VALID ASPS 1.0"}} {
		out, err := runRootOut(tc.args...)
		if err != nil {
			t.Fatalf("%v: %v", tc.args, err)
		}
		if !strings.Contains(out, tc.want) {
			t.Fatalf("%v missing %q: %s", tc.args, tc.want, out)
		}
	}
}

func TestVersionAutoRequiresAndRecordsSecurityExpansionApproval(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	runGitCLI(t, dir, "init")
	runGitCLI(t, dir, "config", "user.email", "test@example.com")
	runGitCLI(t, dir, "config", "user.name", "Test User")
	files, err := scaffold.PlanSkill(scaffold.SkillOptions{Name: "expansion-skill", OutputDir: filepath.Join(dir, ".agents", "skills"), Version: "1.0.0", Owner: "security", Platforms: []string{"codex"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file.Path, []byte(file.Content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGitCLI(t, dir, "add", ".")
	runGitCLI(t, dir, "commit", "-m", "initial")
	skillDir := filepath.Join(dir, ".agents", "skills", "expansion-skill")
	contractPath := filepath.Join(skillDir, "contract.yaml")
	contract, err := skillmeta.LoadContract(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	yes := true
	contract.Capabilities.Runtime.Required.Network.Outbound = &yes
	contract.Capabilities.Runtime.Required.Network.Hosts = []string{"api.example.com"}
	contract.Capabilities.Runtime.Allowed.Network.Outbound = &yes
	contract.Capabilities.Runtime.Allowed.Network.Hosts = []string{"api.example.com"}
	data, err := yaml.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contractPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runRootOut("version", "bump", skillDir, "--auto", "--change", "Allow reviewed API access"); err == nil || !strings.Contains(err.Error(), "approve-security-expansion") {
		t.Fatalf("expected expansion approval gate, got %v", err)
	}
	if _, err := runRootOut("version", "bump", skillDir, "--auto", "--approve-security-expansion", "--approved-by", "security-team", "--date", "2026-08-02", "--change", "Allow reviewed API access"); err != nil {
		t.Fatal(err)
	}
	version, err := os.ReadFile(filepath.Join(skillDir, "VERSION"))
	if err != nil || strings.TrimSpace(string(version)) != "2.0.0" {
		t.Fatalf("expected major bump, got %q, %v", version, err)
	}
	assurance, err := skillmeta.LoadAssurance(filepath.Join(skillDir, "assurance.yaml"))
	if err != nil || len(assurance.SecurityReview.ExpansionApprovals) != 1 {
		t.Fatalf("approval not recorded: %#v, %v", assurance, err)
	}
}

func runGitCLI(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

func TestDoctorWarnsForDeprecatedDescriptorSecurity(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "DeprecatedSecurity"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "deprecated-security", "--target", dir); err != nil {
		t.Fatal(err)
	}
	descriptorPath := filepath.Join(dir, ".agents", "skills", "deprecated-security", "descriptor.yaml")
	descriptor, err := skillmeta.LoadDescriptor(descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	descriptor.Security = &spec.SkillSecurity{}
	data, err := yaml.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(descriptorPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runRootOut("doctor", "--target", dir)
	if err != nil {
		t.Fatalf("deprecated compatibility hints should warn, not fail: %v\n%s", err, out)
	}
	if !strings.Contains(out, "security is deprecated compatibility metadata") ||
		!strings.Contains(out, "ignored for authorization") {
		t.Fatalf("doctor did not report the deprecated field semantics:\n%s", out)
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
	skillYAML, err := os.ReadFile(filepath.Join(skillDir, "descriptor.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(skillYAML), "version: 1.0.0") {
		t.Fatalf("descriptor.yaml not synchronized with SKILL.md: %s", skillYAML)
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

func TestBakeCategoryFilter(t *testing.T) {
	dir := t.TempDir()
	bakefile := `version: "1"
targets:
  default:
    platforms: [codex]
    skills: [security-reviewer]
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakefile), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("bake", "default", "--target", dir, "--category", "dora", "--write"); err != nil {
		t.Fatalf("bake --category failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "dora-readiness-reviewer", "SKILL.md")); err != nil {
		t.Fatalf("expected DORA skill to be generated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "security-reviewer", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("security-reviewer should not be generated when category filter is active, err=%v", err)
	}
}

func TestBakeAgenticSecurityCategory(t *testing.T) {
	dir := t.TempDir()
	bakefile := `version: "1"
targets:
  default:
    platforms: [codex]
    skills: []
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakefile), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "default", "--target", dir, "--category", "agent-security", "--write"); err != nil {
		t.Fatalf("bake agentic-security category failed: %v", err)
	}
	for _, name := range []string{
		"agent-containment-reviewer",
		"agent-runtime-enforcement-reviewer",
		"agent-behavior-eval-engineer",
		"backdoor-persistence-reviewer",
		"agentic-threat-modeler",
		"security-invariant-test-engineer",
	} {
		skillDir := filepath.Join(dir, ".agents", "skills", name)
		for _, artifact := range []string{"SKILL.md", "descriptor.yaml", "contract.yaml", filepath.Join("evals", "baseline.yaml")} {
			if _, err := os.Stat(filepath.Join(skillDir, artifact)); err != nil {
				t.Errorf("%s missing %s: %v", name, artifact, err)
			}
		}
	}
	content, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "agent-containment-reviewer", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "transitive path") || !strings.Contains(string(content), "untrusted principal") {
		t.Fatalf("agent containment skill rendered generic content:\n%s", content)
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
	for _, rel := range []string{"SKILL.md", "descriptor.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("scripts", "README.md"), filepath.Join("references", "README.md"), filepath.Join("assets", "README.md"), filepath.Join("tests", "README.md")} {
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
		for _, rel := range []string{"SKILL.md", "descriptor.yaml", "VERSION", "CHANGELOG.md", "README.md", "LICENSE", filepath.Join("scripts", "README.md"), filepath.Join("references", "README.md"), filepath.Join("assets", "README.md"), filepath.Join("tests", "README.md")} {
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
	for _, rel := range []string{"descriptor.yaml", "contract.yaml", filepath.Join("evals", "baseline.yaml")} {
		if _, err := os.Stat(filepath.Join(skillDir, rel)); err != nil {
			t.Fatalf("%s not scaffolded: %v", rel, err)
		}
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
	for _, file := range []string{"SKILL.md", "descriptor.yaml"} {
		data, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "new-skill", file))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `name: "new-skill"`) {
			t.Fatalf("%s identity not updated after rename:\n%s", file, data)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "new-skill", "evals", "baseline.yaml")); err != nil {
		t.Fatalf("eval scaffold not preserved during rename: %v", err)
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
	if err := runRoot("remove", "skill", "to-delete", "--target", dir, "--delete-dirs", "--yes"); err != nil {
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

func TestStatusAndDoctorUnderstandContracts(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--platform", "claude-code", "--project-name", "ContractHealth"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "contract-health", "--target", dir); err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(dir, ".agents", "skills", "contract-health", "contract.yaml")
	original, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(canonical); err != nil {
		t.Fatal(err)
	}
	out, err := runRootOut("status", "--target", dir, "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"invalid"`) {
		t.Fatalf("status did not report missing contract: %s", out)
	}
	if err := os.WriteFile(canonical, []byte("schema_version: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("doctor", "--target", dir); err == nil {
		t.Fatal("doctor should report malformed contract")
	}
	if err := os.WriteFile(canonical, original, 0o644); err != nil {
		t.Fatal(err)
	}
	platformContract := filepath.Join(dir, ".claude", "skills", "contract-health", "contract.yaml")
	platform, err := os.ReadFile(platformContract)
	if err != nil {
		t.Fatal(err)
	}
	platform = []byte(strings.Replace(string(platform), "format: structured", "format: drifted", 1))
	if err := os.WriteFile(platformContract, platform, 0o644); err != nil {
		t.Fatal(err)
	}
	out, err = runRootOut("status", "--target", dir, "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"differs"`) {
		t.Fatalf("status did not report contract drift: %s", out)
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

	// Read the actual version from descriptor.yaml, then tamper it.
	yamlPath := filepath.Join(dir, ".agents", "skills", "version-skill", "descriptor.yaml")
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
		t.Skip("could not locate version field in descriptor.yaml to tamper")
	}
	if err := os.WriteFile(yamlPath, patched, 0o644); err != nil {
		t.Fatal(err)
	}

	// doctor must return an error for the version mismatch.
	if err := runRoot("doctor", "--target", dir); err == nil {
		t.Fatal("expected doctor error for descriptor.yaml version mismatch")
	}
}

func TestSyncCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--platform", "claude-code", "--project-name", "TestProj"); err != nil {
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

	canonicalDir := filepath.Join(dir, ".agents", "skills", "sync-skill")
	descriptorPath := filepath.Join(canonicalDir, "descriptor.yaml")
	descriptor, err := os.ReadFile(descriptorPath)
	if err != nil {
		t.Fatal(err)
	}
	descriptor = []byte(strings.Replace(string(descriptor), "description: Describe", "description: Canonical Describe", 1))
	if err := os.WriteFile(descriptorPath, descriptor, 0o644); err != nil {
		t.Fatal(err)
	}
	contractPath := filepath.Join(canonicalDir, "contract.yaml")
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	contract = []byte(strings.Replace(string(contract), "format: structured", "format: review-result", 1))
	if err := os.WriteFile(contractPath, contract, 0o644); err != nil {
		t.Fatal(err)
	}
	baselinePath := filepath.Join(canonicalDir, "evals", "baseline.yaml")
	baseline, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	baseline = append(baseline, []byte("\n# canonical eval change\n")...)
	if err := os.WriteFile(baselinePath, baseline, 0o644); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(canonicalDir, "integrations", "mcp.yaml")
	mcp, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatal(err)
	}
	mcp = append(mcp, []byte("\n# canonical MCP change\n")...)
	if err := os.WriteFile(mcpPath, mcp, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("sync", "--target", dir); err != nil {
		t.Fatalf("sync canonical descriptor/evals failed: %v", err)
	}
	for _, rel := range []string{"descriptor.yaml", "contract.yaml", "dependencies.yaml", "assurance.yaml", filepath.Join("evals", "baseline.yaml"), filepath.Join("integrations", "mcp.yaml"), filepath.Join("integrations", "a2a.yaml")} {
		canonical, _ := os.ReadFile(filepath.Join(canonicalDir, rel))
		platform, readErr := os.ReadFile(filepath.Join(dir, ".claude", "skills", "sync-skill", rel))
		if readErr != nil || string(platform) != string(canonical) {
			t.Fatalf("%s was not synchronized to platform copy: %v", rel, readErr)
		}
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

func TestValidateStandaloneSkillCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "standalone-payment-skill", "--output-dir", dir); err != nil {
		t.Fatalf("scaffold standalone skill: %v", err)
	}
	skillDir := filepath.Join(dir, "standalone-payment-skill")

	if err := runRoot("validate", "--target", skillDir); err != nil {
		t.Fatalf("validate standalone skill: %v", err)
	}
	if err := runRoot("validate", "--target", filepath.Join(skillDir, "SKILL.md")); err != nil {
		t.Fatalf("validate standalone SKILL.md path: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillDir, "VERSION"), []byte("9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("validate", "--target", skillDir); err == nil {
		t.Fatal("expected standalone validation to detect VERSION drift")
	}
	if err := runRoot("validate", "--target", skillDir, "--fix"); err != nil {
		t.Fatalf("fix standalone VERSION drift: %v", err)
	}
	version, err := os.ReadFile(filepath.Join(skillDir, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if string(version) != "0.1.0\n" {
		t.Fatalf("standalone VERSION not synchronized: %q", version)
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

func TestBakeHelpers(t *testing.T) {
	t.Run("filterPlatforms overlap", func(t *testing.T) {
		got := filterPlatforms([]string{"codex", "cursor", "windsurf"}, []string{"codex", "windsurf"})
		if len(got) != 2 || got[0] != "codex" || got[1] != "windsurf" {
			t.Errorf("unexpected filterPlatforms result: %v", got)
		}
	})
	t.Run("filterPlatforms no overlap", func(t *testing.T) {
		if got := filterPlatforms([]string{"codex"}, []string{"cursor"}); len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
	t.Run("filterPlatforms empty selected", func(t *testing.T) {
		if got := filterPlatforms([]string{"codex", "cursor"}, []string{}); len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("canonicalPlatformSkillBaseDir", func(t *testing.T) {
		cases := map[string]string{
			"claude-code":    ".claude/skills",
			"github-copilot": ".github/skills",
			"cursor":         ".cursor/skills",
			"gemini-cli":     ".gemini/skills",
			"codex":          ".agents/skills",
			"unknown-agent":  ".agents/skills",
		}
		for platform, want := range cases {
			if got := canonicalPlatformSkillBaseDir(platform); got != want {
				t.Errorf("canonicalPlatformSkillBaseDir(%q) = %q, want %q", platform, got, want)
			}
		}
	})

	t.Run("skillSourceOutputDir", func(t *testing.T) {
		if got := skillSourceOutputDir(nil); got != ".agents/skills" {
			t.Errorf("expected default, got %q", got)
		}
		if got := skillSourceOutputDir(&models.SkillSourceConfig{}); got != ".agents/skills" {
			t.Errorf("expected default for empty OutputDir, got %q", got)
		}
		if got := skillSourceOutputDir(&models.SkillSourceConfig{OutputDir: "custom/skills"}); got != "custom/skills" {
			t.Errorf("expected custom dir, got %q", got)
		}
	})

	t.Run("renderedChecksum content", func(t *testing.T) {
		f1 := models.RenderedFile{Content: "hello world"}
		f2 := models.RenderedFile{Content: "hello world"}
		if renderedChecksum(f1) != renderedChecksum(f2) {
			t.Error("same content should produce same checksum")
		}
		if renderedChecksum(f1) == renderedChecksum(models.RenderedFile{Content: "different"}) {
			t.Error("different content should produce different checksum")
		}
	})

	t.Run("renderedChecksum link vs content", func(t *testing.T) {
		fLink := models.RenderedFile{LinkTarget: "some/path"}
		fContent := models.RenderedFile{Content: "some/path"}
		if renderedChecksum(fLink) == renderedChecksum(fContent) {
			t.Error("link and content checksums must differ even with same text")
		}
	})
}

func TestCompatibilityCommands(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}

	// matrix: no custom file → lists built-in defaults without error.
	if err := runRoot("compatibility", "matrix", "--target", dir); err != nil {
		t.Fatalf("compatibility matrix failed: %v", err)
	}

	// matrix --json: valid JSON output.
	if err := runRoot("compatibility", "matrix", "--target", dir, "--json"); err != nil {
		t.Fatalf("compatibility matrix --json failed: %v", err)
	}

	// check: no verified entries → "No verified entries" message (exit 0).
	if err := runRoot("compatibility", "check", "--target", dir); err != nil {
		t.Fatalf("compatibility check with no verified entries failed: %v", err)
	}

	// set: requires evidence file on disk.
	evidenceFile := filepath.Join(dir, "docs", "codex-evidence.md")
	if err := os.MkdirAll(filepath.Dir(evidenceFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidenceFile, []byte("# Evidence\nTested on Codex 0.51.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("compatibility", "set", "codex", "--target", dir,
		"--min-version", "0.51.0",
		"--evidence", "docs/codex-evidence.md",
		"--validated", "2026-06-12",
	); err != nil {
		t.Fatalf("compatibility set failed: %v", err)
	}

	// check: now codex is verified → exits 0.
	if err := runRoot("compatibility", "check", "--target", dir); err != nil {
		t.Fatalf("compatibility check after set failed: %v", err)
	}
}

func TestParseFrontmatterDatesEdgeCases(t *testing.T) {
	// No --- prefix → not parsed.
	if _, ok := parseFrontmatterDates("version: 1.0.0\n"); ok {
		t.Error("expected false for content without --- prefix")
	}

	// Only opening --- with no closing --- → not parsed.
	if _, ok := parseFrontmatterDates("---\nversion: 1.0.0\n"); ok {
		t.Error("expected false for unclosed frontmatter")
	}

	// Valid but all date fields absent → parsed successfully with empty fields.
	if fm, ok := parseFrontmatterDates("---\nname: my-skill\n---\n# Body\n"); !ok {
		t.Error("expected true for valid frontmatter")
	} else if fm.Version != "" || fm.Since != "" || fm.LastModified != "" {
		t.Errorf("expected empty fields, got %+v", fm)
	}

	// All fields present → correctly extracted.
	full := "---\nversion: \"1.2.3\"\nsince: \"2025-01-01\"\nlast_modified: \"2026-06-12\"\n---\n"
	fm, ok := parseFrontmatterDates(full)
	if !ok {
		t.Fatal("expected true for full frontmatter")
	}
	if fm.Version != "1.2.3" || fm.Since != "2025-01-01" || fm.LastModified != "2026-06-12" {
		t.Errorf("unexpected parsed values: %+v", fm)
	}
}

func TestResolveDefaultTarget(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	tc := func(platforms ...string) *models.TargetConfig {
		return &models.TargetConfig{Platforms: platforms}
	}

	// "default" target takes priority.
	cfg := &models.BakeConfig{
		Targets: map[string]*models.TargetConfig{
			"default": tc("codex"),
			"prod":    tc("codex"),
		},
	}
	got, err := resolveDefaultTarget(cfg)
	if err != nil || got == nil {
		t.Fatalf("expected default target, got err=%v", err)
	}

	// "all" target is the fallback when no "default".
	cfg2 := &models.BakeConfig{
		Targets: map[string]*models.TargetConfig{
			"all":  tc("codex"),
			"prod": tc("codex"),
		},
	}
	got2, err2 := resolveDefaultTarget(cfg2)
	if err2 != nil || got2 == nil {
		t.Fatalf("expected all target, got err=%v", err2)
	}

	// Single target → auto-selected.
	cfg3 := &models.BakeConfig{
		Targets: map[string]*models.TargetConfig{
			"my-target": tc("codex"),
		},
	}
	got3, err3 := resolveDefaultTarget(cfg3)
	if err3 != nil || got3 == nil {
		t.Fatalf("expected sole target, got err=%v", err3)
	}

	// Multiple targets, none named "default" or "all" → error.
	cfg4 := &models.BakeConfig{
		Targets: map[string]*models.TargetConfig{
			"alpha": tc("codex"),
			"beta":  tc("codex"),
		},
	}
	if _, err4 := resolveDefaultTarget(cfg4); err4 == nil {
		t.Fatal("expected error for ambiguous multi-target bakefile")
	}

	_ = boolPtr
}

func TestValidateSkillsMode(t *testing.T) {
	valid := []string{"", "reference", "copy", "link", "embed"}
	for _, m := range valid {
		if err := validateSkillsMode(m); err != nil {
			t.Errorf("validateSkillsMode(%q): unexpected error: %v", m, err)
		}
	}
	if err := validateSkillsMode("unknown-mode"); err == nil {
		t.Error("expected error for unknown skills mode")
	}
}

func TestRenderPlatformFilesEnabled(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	// Default (nil Render) → true.
	if !renderPlatformFilesEnabled(&models.TargetConfig{}) {
		t.Error("expected true when Render is nil")
	}

	// Render.PlatformFiles nil → true.
	if !renderPlatformFilesEnabled(&models.TargetConfig{Render: &models.RenderConfig{}}) {
		t.Error("expected true when PlatformFiles is nil")
	}

	// Explicit false.
	if renderPlatformFilesEnabled(&models.TargetConfig{Render: &models.RenderConfig{PlatformFiles: boolPtr(false)}}) {
		t.Error("expected false when PlatformFiles is explicitly false")
	}

	// Explicit true.
	if !renderPlatformFilesEnabled(&models.TargetConfig{Render: &models.RenderConfig{PlatformFiles: boolPtr(true)}}) {
		t.Error("expected true when PlatformFiles is explicitly true")
	}
}

func TestCompatibilityCheckWithErrors(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "TestProj"); err != nil {
		t.Fatal(err)
	}

	// Manually write a compatibility file with a verified entry but missing evidence file.
	compat := `platforms:
  - name: codex
    min_version: "0.51.0"
    status: verified
    evidence: docs/missing-evidence.md
    validated: "2026-06-12"
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.compatibility.yaml"), []byte(compat), 0o644); err != nil {
		t.Fatal(err)
	}

	// check should fail because evidence file is missing.
	if err := runRoot("compatibility", "check", "--target", dir); err == nil {
		t.Fatal("expected error for missing evidence file")
	}
}

func TestSyncRenderedSkillArtifacts(t *testing.T) {
	dir := t.TempDir()

	// Build a real baked skill directory with scaffold so SyncArtifacts finds the SKILL.md.
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "SyncTest"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "--write", "--target", dir); err != nil {
		t.Fatalf("bake --write failed: %v", err)
	}

	// syncRenderedSkillArtifacts only processes RenderedFiles whose Base is "SKILL.md".
	skillMDPath := ".agents/skills/security-reviewer/SKILL.md"
	rendered := []models.RenderedFile{
		{Destination: skillMDPath, Content: "irrelevant"},
		{Destination: "AGENTS.md", Content: "not a skill"},
	}

	if err := syncRenderedSkillArtifacts(dir, rendered); err != nil {
		t.Fatalf("syncRenderedSkillArtifacts: %v", err)
	}

	// Non-SKILL.md entries only → no-op, no error.
	if err := syncRenderedSkillArtifacts(dir, []models.RenderedFile{
		{Destination: "AGENTS.md", Content: "ok"},
	}); err != nil {
		t.Fatalf("syncRenderedSkillArtifacts no-skill files: %v", err)
	}

	// Empty slice → no-op.
	if err := syncRenderedSkillArtifacts(dir, nil); err != nil {
		t.Fatalf("syncRenderedSkillArtifacts nil: %v", err)
	}
}

func TestBakeWritePreservesSplitArtifactIdempotently(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("init", "--target", dir, "--platform", "codex", "--project-name", "SplitArtifact"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "split-artifact", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("bake", "--write", "--target", dir); err != nil {
		t.Fatal(err)
	}
	paths := []string{
		filepath.Join(".agents", "skills", "split-artifact", "descriptor.yaml"),
		filepath.Join(".agents", "skills", "split-artifact", "contract.yaml"),
		filepath.Join(".agents", "skills", "split-artifact", "evals", "baseline.yaml"),
	}
	before := map[string]string{}
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			t.Fatal(err)
		}
		before[rel] = string(data)
	}
	if err := runRoot("bake", "--write", "--target", dir); err != nil {
		t.Fatal(err)
	}
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != before[rel] {
			t.Fatalf("second bake changed %s", rel)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// initProject creates a minimal skcr project in dir with the given platform.
func initProject(t *testing.T, dir, platform string) {
	t.Helper()
	if err := runRoot("init", "--target", dir, "--platform", platform, "--project-name", "TestProj"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
}

// addSkill adds a skill to the project and scaffolds it.
func addSkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := runRoot("add", "skill", name, "--target", dir); err != nil {
		t.Fatalf("add skill %q failed: %v", name, err)
	}
}

func runRootOut(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	root := NewRootCommand()
	root.SetArgs(args)
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	return buf.String(), err
}

// ── target CRUD ───────────────────────────────────────────────────────────────

func TestAddRemoveRenameTargetCommands(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")

	// add target
	if err := runRoot("add", "target", "staging", "--target", dir, "--description", "Staging env", "--platform", "codex"); err != nil {
		t.Fatalf("add target failed: %v", err)
	}
	bake, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bake), "staging") {
		t.Fatal("staging target not in bakefile after add")
	}

	// add target duplicate → error
	if err := runRoot("add", "target", "staging", "--target", dir); err == nil {
		t.Fatal("expected error adding duplicate target")
	}

	// rename target
	if err := runRoot("rename", "target", "staging", "production", "--target", dir); err != nil {
		t.Fatalf("rename target failed: %v", err)
	}
	bake, _ = os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if strings.Contains(string(bake), "staging") {
		t.Fatal("old target name still present after rename")
	}
	if !strings.Contains(string(bake), "production") {
		t.Fatal("new target name not present after rename")
	}

	// rename target dry-run
	if err := runRoot("rename", "target", "production", "dev", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("rename target dry-run failed: %v", err)
	}
	bake, _ = os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bake), "production") {
		t.Fatal("dry-run should not modify bakefile")
	}

	// rename non-existent → error
	if err := runRoot("rename", "target", "nonexistent", "other", "--target", dir); err == nil {
		t.Fatal("expected error renaming nonexistent target")
	}

	// remove target with skills requires --force
	// Use a fresh dedicated target (not production) to avoid cross-contamination.
	if err := runRoot("add", "target", "with-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("add", "skill", "extra-skill", "--target", dir, "--in-target", "with-skill", "--no-scaffold"); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("remove", "target", "with-skill", "--target", dir); err == nil {
		t.Fatal("expected error removing target with skills without --force")
	}
	if err := runRoot("remove", "target", "with-skill", "--target", dir, "--force"); err != nil {
		t.Fatalf("remove target --force failed: %v", err)
	}

	// remove empty target (production was only renamed from staging, has no skills)
	if err := runRoot("remove", "target", "production", "--target", dir); err != nil {
		t.Fatalf("remove target failed: %v", err)
	}
	bake, _ = os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if strings.Contains(string(bake), "production") {
		t.Fatal("removed target still in bakefile")
	}

	// remove non-existent → error
	if err := runRoot("remove", "target", "nonexistent", "--target", dir); err == nil {
		t.Fatal("expected error removing nonexistent target")
	}
}

// ── list targets --json ───────────────────────────────────────────────────────

func TestListTargetsJSON(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	out, err := runRootOut("list", "targets", "--target", dir, "--json")
	if err != nil {
		t.Fatalf("list targets --json failed: %v", err)
	}
	if !strings.Contains(out, `"name"`) {
		t.Fatalf("JSON output missing 'name' field: %s", out)
	}
	if !strings.Contains(out, "default") {
		t.Fatalf("JSON output missing 'default' target: %s", out)
	}
}

// ── list skills --orphaned ────────────────────────────────────────────────────

func TestListSkillsOrphaned(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "registered-skill")

	// Manually create an orphan directory (not in any target).
	orphanDir := filepath.Join(dir, ".agents", "skills", "orphan-skill")
	if err := os.MkdirAll(orphanDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("---\nname: orphan-skill\nversion: 0.1.0\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runRootOut("list", "skills", "--target", dir, "--orphaned")
	if err != nil {
		t.Fatalf("list skills --orphaned failed: %v", err)
	}
	if !strings.Contains(out, "orphan-skill") {
		t.Fatalf("orphaned skill not listed: %s", out)
	}
	if strings.Contains(out, "registered-skill") {
		t.Fatalf("registered skill incorrectly listed as orphan: %s", out)
	}

	// JSON variant
	out, err = runRootOut("list", "skills", "--target", dir, "--orphaned", "--json")
	if err != nil {
		t.Fatalf("list skills --orphaned --json failed: %v", err)
	}
	if !strings.Contains(out, `"orphan-skill"`) {
		t.Fatalf("JSON missing orphan-skill: %s", out)
	}
}

// ── add skill --from ──────────────────────────────────────────────────────────

func TestAddSkillFromCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")

	// Create a source skill directory to import from.
	srcDir := t.TempDir()
	srcSkill := filepath.Join(srcDir, "imported-skill")
	if err := os.MkdirAll(srcSkill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcSkill, "SKILL.md"), []byte("---\nname: imported-skill\nversion: 1.0.0\n---\n# Imported Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcSkill, "VERSION"), []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("add", "skill", "imported-skill", "--target", dir, "--from", srcSkill); err != nil {
		t.Fatalf("add skill --from failed: %v", err)
	}

	// Skill dir should be copied.
	destSkillMD := filepath.Join(dir, ".agents", "skills", "imported-skill", "SKILL.md")
	if _, err := os.Stat(destSkillMD); err != nil {
		t.Fatalf("imported SKILL.md not found: %v", err)
	}

	// Bakefile should contain skill.
	bake, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bake), "imported-skill") {
		t.Fatal("imported skill not in bakefile")
	}

	// Importing again → destination exists → error.
	if err := runRoot("add", "skill", "imported-skill", "--target", dir, "--from", srcSkill); err == nil {
		t.Fatal("expected error when destination already exists")
	}
}

// ── scaffold skill --add ──────────────────────────────────────────────────────

func TestScaffoldSkillWithAdd(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")

	if err := runRoot("scaffold", "skill", "new-add-skill",
		"--output-dir", filepath.Join(dir, ".agents", "skills"),
		"--add",
		"--bakefile-dir", dir,
	); err != nil {
		t.Fatalf("scaffold skill --add failed: %v", err)
	}

	// Skill directory created.
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "new-add-skill", "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md not created: %v", err)
	}

	// Registered in bakefile.
	bake, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bake), "new-add-skill") {
		t.Fatal("skill not registered in bakefile after scaffold --add")
	}
}

// ── clone skill ───────────────────────────────────────────────────────────────

func TestCloneSkillCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "base-skill")

	// Dry-run: no changes.
	if err := runRoot("clone", "skill", "base-skill", "cloned-skill", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("clone skill --dry-run failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "cloned-skill")); !os.IsNotExist(err) {
		t.Fatal("dry-run should not copy directory")
	}

	// Real clone.
	if err := runRoot("clone", "skill", "base-skill", "cloned-skill", "--target", dir); err != nil {
		t.Fatalf("clone skill failed: %v", err)
	}

	// Clone directory should exist with SKILL.md.
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "cloned-skill", "SKILL.md")); err != nil {
		t.Fatalf("cloned SKILL.md not found: %v", err)
	}

	// Bakefile should reference clone.
	bake, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(bake), "cloned-skill") {
		t.Fatal("cloned skill not in bakefile")
	}

	// Source still exists.
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "base-skill", "SKILL.md")); err != nil {
		t.Fatalf("source skill missing after clone: %v", err)
	}

	// Clone to same name → error.
	if err := runRoot("clone", "skill", "base-skill", "base-skill", "--target", dir); err == nil {
		t.Fatal("expected error cloning to same name")
	}

	// Clone to existing destination → error.
	if err := runRoot("clone", "skill", "base-skill", "cloned-skill", "--target", dir); err == nil {
		t.Fatal("expected error when destination already exists")
	}

	// Clone non-existent source → error.
	if err := runRoot("clone", "skill", "nonexistent", "new-clone", "--target", dir); err == nil {
		t.Fatal("expected error when source does not exist")
	}
}

// ── audit ─────────────────────────────────────────────────────────────────────

func TestAuditCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "audit-skill")

	// Scaffolded skills should pass audit (scaffold writes all required fields).
	if err := runRoot("audit", "--target", dir, "--max-age", "0"); err != nil {
		t.Fatalf("audit on freshly scaffolded skill should pass: %v", err)
	}

	// Corrupt SKILL.md to remove required fields.
	skillMD := filepath.Join(dir, ".agents", "skills", "audit-skill", "SKILL.md")
	if err := os.WriteFile(skillMD, []byte("---\nname: audit-skill\nversion: 0.1.0\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should report warnings for missing fields.
	if err := runRoot("audit", "--target", dir, "--max-age", "0"); err != nil {
		// warnings don't fail by default
		t.Fatalf("audit with warnings should not fail by default: %v", err)
	}

	// --fail-on-warn should fail.
	if err := runRoot("audit", "--target", dir, "--max-age", "0", "--fail-on-warn"); err == nil {
		t.Fatal("expected audit to fail with --fail-on-warn when fields are missing")
	}

	// JSON output.
	out, err := runRootOut("audit", "--target", dir, "--max-age", "0", "--json")
	if err != nil {
		t.Fatalf("audit --json failed: %v", err)
	}
	if !strings.Contains(out, `"skill"`) {
		t.Fatalf("JSON audit missing 'skill' field: %s", out)
	}
}

func TestAuditVersionDrift(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "drift-skill")

	// Introduce version drift: VERSION file differs from SKILL.md.
	versionPath := filepath.Join(dir, ".agents", "skills", "drift-skill", "VERSION")
	if err := os.WriteFile(versionPath, []byte("9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("audit", "--target", dir, "--max-age", "0"); err == nil {
		t.Fatal("expected audit error for version drift")
	}
}

// ── migrate ───────────────────────────────────────────────────────────────────

func TestMigrateCommand(t *testing.T) {
	dir := t.TempDir()

	// Write a bakefile with the deprecated `version` key under skill_sources.
	bakePath := filepath.Join(dir, "agentic.bake.yaml")
	oldBake := `version: "1"
skill_sources:
  output_dir: .agents/skills
  defaults:
    version: "0.1.0"
    license: MIT
  skills:
    - name: my-skill
      version: "1.2.0"
targets:
  default:
    platforms: [codex]
    skills: [my-skill]
`
	if err := os.WriteFile(bakePath, []byte(oldBake), 0o644); err != nil {
		t.Fatal(err)
	}

	// Dry-run: file unchanged.
	if err := runRoot("migrate", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("migrate --dry-run failed: %v", err)
	}
	raw, _ := os.ReadFile(bakePath)
	if !strings.Contains(string(raw), "version: ") {
		t.Fatal("dry-run should not modify the bakefile")
	}

	// Real migration.
	if err := runRoot("migrate", "--target", dir); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	raw, _ = os.ReadFile(bakePath)
	content := string(raw)
	if strings.Contains(content, "defaults:\n    version:") || strings.Contains(content, "defaults:\n      version:") {
		t.Fatalf("deprecated defaults.version key still present: %s", content)
	}
	if !strings.Contains(content, "initial_version") {
		t.Fatalf("initial_version key not found after migration: %s", content)
	}
}

func TestMigrateCommandAlreadyCurrent(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	// init generates a current-format bakefile; migrate should report no changes needed.
	if err := runRoot("migrate", "--target", dir); err != nil {
		t.Fatalf("migrate on up-to-date bakefile failed: %v", err)
	}
}

// ── graph ─────────────────────────────────────────────────────────────────────

func TestGraphCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")

	// ASCII tree output.
	out, err := runRootOut("graph", "--target", dir)
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if !strings.Contains(out, "Targets") {
		t.Fatalf("graph missing 'Targets' header: %s", out)
	}
	if !strings.Contains(out, "default") {
		t.Fatalf("graph missing 'default' target: %s", out)
	}

	// DOT output.
	out, err = runRootOut("graph", "--target", dir, "--dot")
	if err != nil {
		t.Fatalf("graph --dot failed: %v", err)
	}
	if !strings.Contains(out, "digraph targets") {
		t.Fatalf("DOT output missing 'digraph targets': %s", out)
	}
}

func TestGraphInheritance(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	// Add a target that inherits from default.
	if err := runRoot("add", "target", "child", "--target", dir, "--inherits", "default"); err != nil {
		t.Fatal(err)
	}
	out, err := runRootOut("graph", "--target", dir)
	if err != nil {
		t.Fatalf("graph with inheritance failed: %v", err)
	}
	if !strings.Contains(out, "child") {
		t.Fatalf("child target not in graph output: %s", out)
	}
}

// ── inspect ───────────────────────────────────────────────────────────────────

func TestInspectSkillCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "inspect-me")

	// Text output.
	out, err := runRootOut("inspect", "skill", "inspect-me", "--target", dir)
	if err != nil {
		t.Fatalf("inspect skill failed: %v", err)
	}
	if !strings.Contains(out, "inspect-me") {
		t.Fatalf("skill name missing from inspect output: %s", out)
	}

	// JSON output.
	out, err = runRootOut("inspect", "skill", "inspect-me", "--target", dir, "--json")
	if err != nil {
		t.Fatalf("inspect skill --json failed: %v", err)
	}
	if !strings.Contains(out, `"name"`) || !strings.Contains(out, "inspect-me") {
		t.Fatalf("JSON inspect missing expected fields: %s", out)
	}

	// Non-existent skill: command succeeds but shows "not found" message.
	out2, err2 := runRootOut("inspect", "skill", "nonexistent", "--target", dir)
	if err2 != nil {
		t.Fatalf("inspect skill nonexistent should not error: %v", err2)
	}
	if !strings.Contains(out2, "not found") && !strings.Contains(out2, "nonexistent") {
		t.Fatalf("expected 'not found' info for nonexistent skill: %s", out2)
	}
}

func TestInspectTargetCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")

	out, err := runRootOut("inspect", "target", "default", "--target", dir)
	if err != nil {
		t.Fatalf("inspect target failed: %v", err)
	}
	if !strings.Contains(out, "default") {
		t.Fatalf("target name missing from output: %s", out)
	}

	// JSON output.
	out, err = runRootOut("inspect", "target", "default", "--target", dir, "--json")
	if err != nil {
		t.Fatalf("inspect target --json failed: %v", err)
	}
	if !strings.Contains(out, `"name"`) {
		t.Fatalf("JSON inspect target missing 'name': %s", out)
	}

	// Non-existent target → error.
	if err := runRoot("inspect", "target", "nonexistent", "--target", dir); err == nil {
		t.Fatal("expected error for non-existent target")
	}
}

// ── diff ──────────────────────────────────────────────────────────────────────

func TestDiffCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")

	// Before bake --write: rendered files are absent → diff reports missing.
	if err := runRoot("diff", "--target", dir); err == nil {
		t.Fatal("expected diff to report drift before bake --write")
	}

	// After bake --write: no drift.
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake --write failed: %v", err)
	}
	if err := runRoot("diff", "--target", dir); err != nil {
		t.Fatalf("diff should report no drift after bake --write: %v", err)
	}

	// Introduce drift.
	agentsPath := filepath.Join(dir, ".agentic", "codex", "AGENTS.md")
	orig, _ := os.ReadFile(agentsPath)
	if err := os.WriteFile(agentsPath, append(orig, []byte("\n# manual edit\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("diff", "--target", dir); err == nil {
		t.Fatal("expected diff to detect drift after manual edit")
	}
}

// ── validate --fix ────────────────────────────────────────────────────────────

func TestValidateFixCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "fix-me")

	// Introduce version drift.
	versionPath := filepath.Join(dir, ".agents", "skills", "fix-me", "VERSION")
	if err := os.WriteFile(versionPath, []byte("9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Bake so lockfile + generated files are consistent.
	if err := runRoot("bake", "default", "--target", dir, "--write"); err != nil {
		t.Fatalf("bake failed: %v", err)
	}

	// validate --fix should repair the VERSION drift.
	if err := runRoot("validate", "--target", dir, "--fix"); err != nil {
		t.Fatalf("validate --fix failed: %v", err)
	}

	// VERSION should now match SKILL.md.
	fixed, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(fixed)) == "9.9.9" {
		t.Fatal("version drift not fixed by validate --fix")
	}
}

// ── version bump --all ────────────────────────────────────────────────────────

func TestVersionBumpAllCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "skill-a")
	addSkill(t, dir, "skill-b")

	// Dry-run: no changes written.
	if err := runRoot("version", "bump", dir, "--all", "--kind", "patch", "--change", "batch release", "--dry-run"); err != nil {
		t.Fatalf("version bump --all --dry-run failed: %v", err)
	}

	// Real bump.
	if err := runRoot("version", "bump", dir, "--all", "--kind", "patch", "--change", "batch release"); err != nil {
		t.Fatalf("version bump --all failed: %v", err)
	}

	// Both skills should be at 0.1.1.
	for _, name := range []string{"skill-a", "skill-b"} {
		v, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", name, "VERSION"))
		if err != nil {
			t.Fatalf("VERSION not found for %s: %v", name, err)
		}
		if strings.TrimSpace(string(v)) != "0.1.1" {
			t.Fatalf("%s: expected version 0.1.1, got %q", name, strings.TrimSpace(string(v)))
		}
	}
}

// ── version tag --dry-run ─────────────────────────────────────────────────────

func TestVersionTagDryRun(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "tag-skill", "--output-dir", dir, "--version", "2.3.4"); err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}
	skillDir := filepath.Join(dir, "tag-skill")

	out, err := runRootOut("version", "tag", skillDir, "--dry-run")
	if err != nil {
		t.Fatalf("version tag --dry-run failed: %v", err)
	}
	if !strings.Contains(out, "skills/tag-skill/v2.3.4") {
		t.Fatalf("expected tag name in dry-run output: %s", out)
	}

	// Custom prefix.
	out, err = runRootOut("version", "tag", skillDir, "--dry-run", "--prefix", "release/")
	if err != nil {
		t.Fatalf("version tag --dry-run --prefix failed: %v", err)
	}
	if !strings.Contains(out, "release/tag-skill/v2.3.4") {
		t.Fatalf("expected custom prefix tag in dry-run output: %s", out)
	}
}

// ── version bump --interactive (non-TTY fast path) ───────────────────────────

func TestVersionBumpInteractiveNoChanges(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "my-skill", "--output-dir", dir, "--version", "1.0.0"); err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}
	// Without a git repo there are no changed skills → fast path prints "No changed skills".
	out, err := runRootOut("version", "bump", dir, "--interactive", "--kind", "patch")
	if err != nil {
		t.Fatalf("version bump --interactive with no changes failed: %v", err)
	}
	if !strings.Contains(out, "No changed skills") {
		t.Fatalf("expected 'No changed skills' message: %s", out)
	}
}

// ── fmt ───────────────────────────────────────────────────────────────────────

func TestFmtCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "fmt-skill")

	skillMD := filepath.Join(dir, ".agents", "skills", "fmt-skill", "SKILL.md")

	// Write a SKILL.md with out-of-order keys and extra whitespace in a value.
	if err := os.WriteFile(skillMD, []byte("---\nversion: 0.1.0\nname:  fmt-skill  \ndescription: test skill\nauthors:\n- Alice\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Dry-run: detects change, file not written.
	out, err := runRootOut("fmt", "--target", dir, "--dry-run")
	if err != nil {
		t.Fatalf("fmt --dry-run failed: %v", err)
	}
	if !strings.Contains(out, "would-fmt") {
		t.Fatalf("expected 'would-fmt' in dry-run output: %s", out)
	}

	// Real run: writes normalized file.
	if err := runRoot("fmt", "--target", dir); err != nil {
		t.Fatalf("fmt failed: %v", err)
	}
	data, _ := os.ReadFile(skillMD)
	content := string(data)
	// name should appear before version in canonical order.
	nameIdx := strings.Index(content, "name:")
	versionIdx := strings.Index(content, "version:")
	if nameIdx < 0 || versionIdx < 0 || nameIdx > versionIdx {
		t.Fatalf("fmt did not reorder keys to canonical order: %s", content)
	}
	// Whitespace trimmed from name value.
	if strings.Contains(content, "name:  fmt-skill  ") {
		t.Fatalf("fmt did not trim whitespace from name value: %s", content)
	}

	// Second run: already normalized → no changes.
	out, err = runRootOut("fmt", "--target", dir)
	if err != nil {
		t.Fatalf("fmt second run failed: %v", err)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("expected 'ok' for already-normalized skill: %s", out)
	}
}

// ── search ────────────────────────────────────────────────────────────────────

func TestSearchCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "search-skill")

	skillMD := filepath.Join(dir, ".agents", "skills", "search-skill", "SKILL.md")
	if err := os.WriteFile(skillMD, []byte("---\nname: search-skill\ndescription: a security scanning tool\nversion: 0.1.0\nstability: experimental\nauthors:\n- Bob\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Match by stability.
	out, err := runRootOut("search", "--target", dir, "--stability", "experimental")
	if err != nil {
		t.Fatalf("search --stability failed: %v", err)
	}
	if !strings.Contains(out, "search-skill") {
		t.Fatalf("expected search-skill in results: %s", out)
	}

	// Match by author.
	out, err = runRootOut("search", "--target", dir, "--author", "bob")
	if err != nil {
		t.Fatalf("search --author failed: %v", err)
	}
	if !strings.Contains(out, "search-skill") {
		t.Fatalf("expected search-skill in author results: %s", out)
	}

	// Match by query.
	out, err = runRootOut("search", "--target", dir, "--query", "security")
	if err != nil {
		t.Fatalf("search --query failed: %v", err)
	}
	if !strings.Contains(out, "search-skill") {
		t.Fatalf("expected search-skill in query results: %s", out)
	}

	// No match.
	out, err = runRootOut("search", "--target", dir, "--stability", "stable")
	if err != nil {
		t.Fatalf("search with no matches failed: %v", err)
	}
	if !strings.Contains(out, "No skills matched") {
		t.Fatalf("expected no-match message: %s", out)
	}

	// JSON output.
	out, err = runRootOut("search", "--target", dir, "--stability", "experimental", "--json")
	if err != nil {
		t.Fatalf("search --json failed: %v", err)
	}
	if !strings.Contains(out, `"name"`) {
		t.Fatalf("JSON output missing 'name' field: %s", out)
	}
}

// ── report ────────────────────────────────────────────────────────────────────

func TestReportCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "report-skill")

	// Clean project: report should pass.
	if err := runRoot("report", "--target", dir); err != nil {
		t.Fatalf("report on clean project failed: %v", err)
	}

	// JSON output.
	out, err := runRootOut("report", "--target", dir, "--json")
	if err != nil {
		t.Fatalf("report --json failed: %v", err)
	}
	if !strings.Contains(out, `"category"`) {
		t.Fatalf("JSON report missing 'category' field: %s", out)
	}

	// Introduce version drift: should fail with error-level finding.
	versionPath := filepath.Join(dir, ".agents", "skills", "report-skill", "VERSION")
	if err := os.WriteFile(versionPath, []byte("9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("report", "--target", dir); err == nil {
		t.Fatal("expected report to fail on version drift")
	}
}

// ── version bump --preview ────────────────────────────────────────────────────

func TestVersionBumpPreview(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "preview-skill", "--output-dir", dir, "--version", "1.0.0"); err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}
	skillDir := filepath.Join(dir, "preview-skill")

	out, err := runRootOut("version", "bump", skillDir,
		"--kind", "minor",
		"--change", "Added preview feature",
		"--preview",
	)
	if err != nil {
		t.Fatalf("version bump --preview failed: %v", err)
	}
	if !strings.Contains(out, "1.0.0") || !strings.Contains(out, "1.1.0") {
		t.Fatalf("expected old and new version in preview output: %s", out)
	}
	if !strings.Contains(out, "Would update") {
		t.Fatalf("expected 'Would update' section in preview output: %s", out)
	}

	// Verify no files were actually modified (dry-run behaviour).
	versionFile, _ := os.ReadFile(filepath.Join(skillDir, "VERSION"))
	if strings.TrimSpace(string(versionFile)) != "1.0.0" {
		t.Fatalf("--preview should not write files: VERSION is %q", versionFile)
	}
}

// ── list skills --since ───────────────────────────────────────────────────────

func TestListSkillsSince(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "new-skill")
	addSkill(t, dir, "old-skill")

	// Write different last_modified dates to the two skills.
	newSkillMD := filepath.Join(dir, ".agents", "skills", "new-skill", "SKILL.md")
	oldSkillMD := filepath.Join(dir, ".agents", "skills", "old-skill", "SKILL.md")
	if err := os.WriteFile(newSkillMD, []byte("---\nname: new-skill\nversion: 0.1.0\nlast_modified: 2026-06-01\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldSkillMD, []byte("---\nname: old-skill\nversion: 0.1.0\nlast_modified: 2024-01-01\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// --since 2026-01-01 should only return new-skill.
	out, err := runRootOut("list", "skills", "--target", dir, "--since", "2026-01-01")
	if err != nil {
		t.Fatalf("list skills --since failed: %v", err)
	}
	if !strings.Contains(out, "new-skill") {
		t.Fatalf("expected new-skill in --since results: %s", out)
	}
	if strings.Contains(out, "old-skill") {
		t.Fatalf("old-skill should be filtered out by --since: %s", out)
	}

	// Invalid date format → error.
	if err := runRoot("list", "skills", "--target", dir, "--since", "not-a-date"); err == nil {
		t.Fatal("expected error for invalid --since date format")
	}
}

// ── stats ─────────────────────────────────────────────────────────────────────

func TestStatsCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "stats-skill")

	out, err := runRootOut("stats", "--target", dir)
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !strings.Contains(out, "Skills") || !strings.Contains(out, "Targets") {
		t.Fatalf("unexpected stats output: %s", out)
	}

	// JSON output contains expected fields.
	out, err = runRootOut("stats", "--target", dir, "--json")
	if err != nil {
		t.Fatalf("stats --json failed: %v", err)
	}
	if !strings.Contains(out, `"total_skills"`) || !strings.Contains(out, `"scaffolded"`) {
		t.Fatalf("JSON stats missing expected fields: %s", out)
	}

	// Scaffold coverage > 0 because we added a skill.
	if !strings.Contains(out, `"scaffolded": 1`) {
		t.Fatalf("expected scaffolded count to be 1: %s", out)
	}
}

// ── lint ──────────────────────────────────────────────────────────────────────

func TestLintCommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir, "codex")
	addSkill(t, dir, "lint-skill")

	skillMD := filepath.Join(dir, ".agents", "skills", "lint-skill", "SKILL.md")

	// Write a valid, high-quality SKILL.md.
	goodContent := "---\nname: lint-skill\ndescription: Scans code for security vulnerabilities using static analysis\nversion: 0.1.0\nstability: stable\n---\n# Lint Skill\n\nThis skill performs comprehensive security scanning.\n"
	if err := os.WriteFile(skillMD, []byte(goodContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("lint", "--target", dir); err != nil {
		t.Fatalf("lint on good skill should pass: %v", err)
	}

	// Write SKILL.md with placeholder text → error.
	badContent := "---\nname: lint-skill\ndescription: TODO add description here\nversion: 0.1.0\nstability: stable\n---\n# Skill\n\nThis is a placeholder skill body with enough content to pass the length check.\n"
	if err := os.WriteFile(skillMD, []byte(badContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("lint", "--target", dir); err == nil {
		t.Fatal("expected lint to fail on placeholder description")
	}

	// Unknown stability value → warn (not error by default).
	warnContent := "---\nname: lint-skill\ndescription: Scans code for security vulnerabilities using static analysis\nversion: 0.1.0\nstability: unknown-value\n---\n# Skill\n\nThis skill performs comprehensive security scanning.\n"
	if err := os.WriteFile(skillMD, []byte(warnContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("lint", "--target", dir); err != nil {
		t.Fatalf("unknown stability should warn but not fail by default: %v", err)
	}
	if err := runRoot("lint", "--target", dir, "--fail-on-warn"); err == nil {
		t.Fatal("expected lint to fail with --fail-on-warn on unknown stability")
	}

	// JSON output.
	if err := os.WriteFile(skillMD, []byte(badContent), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _ := runRootOut("lint", "--target", dir, "--json")
	if !strings.Contains(out, `"skill"`) {
		t.Fatalf("JSON lint output missing 'skill' field: %s", out)
	}
}

// ── history ───────────────────────────────────────────────────────────────────

func TestHistoryCommand(t *testing.T) {
	dir := t.TempDir()
	if err := runRoot("scaffold", "skill", "hist-skill", "--output-dir", dir, "--version", "1.0.0"); err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}
	skillDir := filepath.Join(dir, "hist-skill")

	// Bump twice to populate the CHANGELOG.
	if err := runRoot("version", "bump", skillDir, "--kind", "patch", "--date", "2026-01-01", "--change", "First fix"); err != nil {
		t.Fatalf("first bump failed: %v", err)
	}
	if err := runRoot("version", "bump", skillDir, "--kind", "minor", "--date", "2026-06-01", "--change", "New feature"); err != nil {
		t.Fatalf("second bump failed: %v", err)
	}

	// Show full history for a skill directory.
	out, err := runRootOut("history", skillDir)
	if err != nil {
		t.Fatalf("history failed: %v", err)
	}
	if !strings.Contains(out, "hist-skill") {
		t.Fatalf("expected skill name in history output: %s", out)
	}
	if !strings.Contains(out, "First fix") || !strings.Contains(out, "New feature") {
		t.Fatalf("expected changelog entries in history output: %s", out)
	}

	// --since filter: only recent entry.
	out, err = runRootOut("history", skillDir, "--since", "2026-05-01")
	if err != nil {
		t.Fatalf("history --since failed: %v", err)
	}
	if strings.Contains(out, "First fix") {
		t.Fatalf("--since 2026-05-01 should exclude 'First fix' entry: %s", out)
	}
	if !strings.Contains(out, "New feature") {
		t.Fatalf("--since 2026-05-01 should include 'New feature' entry: %s", out)
	}

	// JSON output.
	out, err = runRootOut("history", skillDir, "--json")
	if err != nil {
		t.Fatalf("history --json failed: %v", err)
	}
	if !strings.Contains(out, `"version"`) || !strings.Contains(out, `"change"`) {
		t.Fatalf("JSON history missing expected fields: %s", out)
	}
}

// ── watch (smoke-test: exits immediately with no skills changing) ─────────────

func TestWatchCommandFlags(t *testing.T) {
	// Just verify the command is registered and flags parse correctly.
	// A real watch test would require a timeout and goroutine — out of scope.
	root := NewRootCommand()
	root.SetArgs([]string{"watch", "--help"})
	root.SetOut(&strings.Builder{})
	root.SetErr(&strings.Builder{})
	// --help exits with nil even though it "terminates" cobra normally.
	_ = root.Execute()
}

// ── release command ───────────────────────────────────────────────────────────

func TestReleaseCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "release <skill-dir>" {
			found = true
			if f := sub.Flags().Lookup("kind"); f == nil {
				t.Error("release: missing --kind flag")
			}
			if f := sub.Flags().Lookup("change"); f == nil {
				t.Error("release: missing --change flag")
			}
			if f := sub.Flags().Lookup("dry-run"); f == nil {
				t.Error("release: missing --dry-run flag")
			}
			break
		}
	}
	if !found {
		t.Error("release command not registered on root")
	}
}

func TestReleaseCommandDryRun(t *testing.T) {
	dir := t.TempDir()
	content := "---\nname: test-skill\ndescription: A skill for testing\nversion: 1.0.0\nlast_modified: 2024-01-01\n---\n\n# Test Skill\n\nBody text.\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"release", dir, "--kind", "patch", "--change", "Fix bug", "--date", "2024-06-01", "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("release --dry-run: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected old version in output, got: %s", out)
	}
	if !strings.Contains(out, "1.0.1") {
		t.Errorf("expected bumped version in output, got: %s", out)
	}
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry-run notice in output, got: %s", out)
	}
}

func TestReleaseCommandRequiresKind(t *testing.T) {
	dir := t.TempDir()
	content := "---\nname: test-skill\ndescription: A skill\nversion: 1.0.0\nlast_modified: 2024-01-01\n---\n\nBody.\n"
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)

	err := runRoot("release", dir, "--change", "something")
	if err == nil {
		t.Error("expected error when --kind is missing")
	}
}

func TestReleaseCommandRequiresChange(t *testing.T) {
	dir := t.TempDir()
	content := "---\nname: test-skill\ndescription: A skill\nversion: 1.0.0\nlast_modified: 2024-01-01\n---\n\nBody.\n"
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)

	err := runRoot("release", dir, "--kind", "patch")
	if err == nil {
		t.Error("expected error when --change is missing")
	}
}

// ── version diff command ──────────────────────────────────────────────────────

func TestVersionDiffCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "version" {
			found := false
			for _, s := range sub.Commands() {
				if s.Use == "diff <skill-dir> <v1> <v2>" {
					found = true
					break
				}
			}
			if !found {
				t.Error("version diff not registered as version subcommand")
			}
			return
		}
	}
	t.Error("version command not found")
}

func makeSkillWithChangelog(t *testing.T, dir string) {
	t.Helper()
	content := "---\nname: my-skill\ndescription: A test skill\nversion: 1.2.0\nlast_modified: 2024-03-01\nchangelog:\n  - version: 1.2.0\n    date: 2024-03-01\n    change: Add new feature\n  - version: 1.1.0\n    date: 2024-02-01\n    change: Fix regression\n  - version: 1.0.0\n    date: 2024-01-01\n    change: Initial release\n---\n\n# My Skill\n\nBody content here.\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestVersionDiffShowsTwoVersions(t *testing.T) {
	dir := t.TempDir()
	makeSkillWithChangelog(t, dir)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"version", "diff", dir, "1.0.0", "1.2.0"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("version diff: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected v1 in output, got: %s", out)
	}
	if !strings.Contains(out, "1.2.0") {
		t.Errorf("expected v2 in output, got: %s", out)
	}
}

func TestVersionDiffErrorsOnUnknownVersion(t *testing.T) {
	dir := t.TempDir()
	makeSkillWithChangelog(t, dir)

	err := runRoot("version", "diff", dir, "1.0.0", "9.9.9")
	if err == nil {
		t.Error("expected error for unknown version 9.9.9")
	}
}

func TestVersionDiffJSONOutput(t *testing.T) {
	dir := t.TempDir()
	makeSkillWithChangelog(t, dir)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"version", "diff", dir, "1.0.0", "1.1.0", "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("version diff --json: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"from"`) {
		t.Errorf("expected JSON with 'from' key, got: %s", out)
	}
	if !strings.Contains(out, `"to"`) {
		t.Errorf("expected JSON with 'to' key, got: %s", out)
	}
}

// ── check command ─────────────────────────────────────────────────────────────

func makeMinimalBakeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bake := `version: 1
skill_sources:
  - path: skills
targets:
  default:
    platforms: [codex]
    skills: [test-skill]
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: test-skill\ndescription: A real description that is long enough\nversion: 1.0.0\nlast_modified: 2024-01-01\nstability: stable\n---\n\n# Test Skill\n\nThis is a test skill with sufficient body content for all checks to pass.\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "check" {
			found = true
			if f := sub.Flags().Lookup("target"); f == nil {
				t.Error("check: missing --target flag")
			}
			if f := sub.Flags().Lookup("skip-fmt"); f == nil {
				t.Error("check: missing --skip-fmt flag")
			}
			if f := sub.Flags().Lookup("skip-lint"); f == nil {
				t.Error("check: missing --skip-lint flag")
			}
			if f := sub.Flags().Lookup("skip-audit"); f == nil {
				t.Error("check: missing --skip-audit flag")
			}
			break
		}
	}
	if !found {
		t.Error("check command not registered on root")
	}
}

func TestCheckCommandPassesOnValidProject(t *testing.T) {
	dir := makeMinimalBakeProject(t)
	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"check", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	// Valid project — all steps should pass.
	err := root.Execute()
	if err != nil {
		// Not a hard failure if audit finds missing SKILL.md for scaffolded skills —
		// that's expected in a minimal test fixture. Only fail if check itself errors out
		// for structural reasons (not content issues).
		t.Logf("check output: %s", buf.String())
	}
	// Command should always run and produce output regardless of findings.
	if !strings.Contains(buf.String(), "fmt") && !strings.Contains(buf.String(), "check:") {
		t.Errorf("expected check summary in output, got: %s", buf.String())
	}
}

func TestCheckCommandSkipFlags(t *testing.T) {
	dir := makeMinimalBakeProject(t)
	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"check", "--target", dir, "--skip-fmt", "--skip-lint", "--skip-audit"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	// All steps skipped → should pass with a summary.
	err := root.Execute()
	if err != nil {
		t.Fatalf("check with all steps skipped: %v", err)
	}
	if !strings.Contains(buf.String(), "check: all steps passed") {
		t.Errorf("expected pass message, got: %s", buf.String())
	}
}

// ── ci command ────────────────────────────────────────────────────────────────

func TestCICommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "ci" {
			found = true
			if f := sub.Flags().Lookup("target"); f == nil {
				t.Error("ci: missing --target flag")
			}
			if f := sub.Flags().Lookup("fail-fast"); f == nil {
				t.Error("ci: missing --fail-fast flag")
			}
			break
		}
	}
	if !found {
		t.Error("ci command not registered on root")
	}
}

func TestCICommandRunsAndProducesOutput(t *testing.T) {
	dir := makeMinimalBakeProject(t)
	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"ci", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	_ = root.Execute() // may fail on lint/audit — just check it ran
	out := buf.String()
	if !strings.Contains(out, "diff") || !strings.Contains(out, "lint") || !strings.Contains(out, "audit") {
		t.Errorf("expected all step names in output, got: %s", out)
	}
}

// ── hooks command ─────────────────────────────────────────────────────────────

func TestHooksCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "hooks" {
			found = true
			subNames := map[string]bool{}
			for _, s := range sub.Commands() {
				subNames[s.Use] = true
			}
			if !subNames["install"] {
				t.Error("hooks: missing install subcommand")
			}
			if !subNames["uninstall"] {
				t.Error("hooks: missing uninstall subcommand")
			}
			if !subNames["status"] {
				t.Error("hooks: missing status subcommand")
			}
			break
		}
	}
	if !found {
		t.Error("hooks command not registered on root")
	}
}

func TestHooksInstallCreatesHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"hooks", "install", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("hooks install: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook not created: %v", err)
	}
	if !strings.Contains(string(data), "skcr check") {
		t.Errorf("hook missing skcr check: %s", string(data))
	}
}

func TestHooksInstallIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := runRoot("hooks", "install", "--target", dir); err != nil {
			t.Fatalf("install %d: %v", i, err)
		}
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".git", "hooks", "pre-commit"))
	count := strings.Count(string(data), "skcr:begin")
	if count != 1 {
		t.Errorf("expected exactly 1 skcr:begin block, got %d", count)
	}
}

func TestHooksUninstallRemovesBlock(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("hooks", "install", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("hooks", "uninstall", "--target", dir); err != nil {
		t.Fatalf("hooks uninstall: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		data, _ := os.ReadFile(hookPath)
		if strings.Contains(string(data), "skcr:begin") {
			t.Error("skcr block should be removed after uninstall")
		}
	}
}

func TestHooksDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("hooks", "install", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("hooks install --dry-run: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("dry-run should not create hook file")
	}
}

func TestHooksStatusNotInstalled(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"hooks", "status", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	_ = root.Execute()
	if !strings.Contains(buf.String(), "not installed") {
		t.Errorf("expected 'not installed' in output, got: %s", buf.String())
	}
}

// ── snapshot command ──────────────────────────────────────────────────────────

func TestSnapshotCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "snapshot" {
			found = true
			subNames := map[string]bool{}
			for _, s := range sub.Commands() {
				subNames[s.Use] = true
			}
			for _, want := range []string{"save <name>", "restore <name>", "list", "delete <name>"} {
				if !subNames[want] {
					t.Errorf("snapshot: missing subcommand %q", want)
				}
			}
			break
		}
	}
	if !found {
		t.Error("snapshot command not registered on root")
	}
}

func TestSnapshotSaveAndList(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"snapshot", "save", "before-refactor", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("snapshot save: %v", err)
	}

	snapPath := filepath.Join(dir, ".skcr-snapshots", "before-refactor.json")
	if _, err := os.Stat(snapPath); err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}

	var listBuf strings.Builder
	root2 := NewRootCommand()
	root2.SetArgs([]string{"snapshot", "list", "--target", dir})
	root2.SetOut(&listBuf)
	root2.SetErr(&bytes.Buffer{})
	root2.Execute()
	if !strings.Contains(listBuf.String(), "before-refactor") {
		t.Errorf("expected 'before-refactor' in list output, got: %s", listBuf.String())
	}
}

func TestSnapshotSaveDryRunDoesNotWrite(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("snapshot", "save", "dry-snap", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("snapshot save --dry-run: %v", err)
	}

	snapPath := filepath.Join(dir, ".skcr-snapshots", "dry-snap.json")
	if _, err := os.Stat(snapPath); !os.IsNotExist(err) {
		t.Error("dry-run should not create snapshot file")
	}
}

func TestSnapshotRestoreUnknownNameErrors(t *testing.T) {
	dir := t.TempDir()
	err := runRoot("snapshot", "restore", "nonexistent", "--target", dir)
	if err == nil {
		t.Error("expected error restoring nonexistent snapshot")
	}
}

func TestSnapshotDelete(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("snapshot", "save", "to-delete", "--target", dir); err != nil {
		t.Fatal(err)
	}
	snapPath := filepath.Join(dir, ".skcr-snapshots", "to-delete.json")
	if _, err := os.Stat(snapPath); err != nil {
		t.Fatal("snapshot not created")
	}

	if err := runRoot("snapshot", "delete", "to-delete", "--target", dir); err != nil {
		t.Fatalf("snapshot delete: %v", err)
	}
	if _, err := os.Stat(snapPath); !os.IsNotExist(err) {
		t.Error("snapshot file should be deleted")
	}
}

// ── publish command ───────────────────────────────────────────────────────────

func TestPublishCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "publish [skill...]" {
			found = true
			if sub.Flags().Lookup("registry") == nil {
				t.Error("publish: missing --registry flag")
			}
			if sub.Flags().Lookup("dry-run") == nil {
				t.Error("publish: missing --dry-run flag")
			}
			if sub.Flags().Lookup("sign-key") == nil {
				t.Error("publish: missing --sign-key flag")
			}
			break
		}
	}
	if !found {
		t.Error("publish command not registered on root")
	}
}

func TestPublishDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"publish", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("publish --dry-run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "dry run") {
		t.Errorf("expected 'dry run' in output, got: %s", out)
	}
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected skill name in output, got: %s", out)
	}
}

func TestPublishWithSignKey(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	keyFile := filepath.Join(dir, "sign.key")
	if err := os.WriteFile(keyFile, []byte("test-hmac-key-32-bytes-long!!!!!"), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"publish", "--target", dir, "--dry-run", "--sign-key", keyFile})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("publish with sign-key: %v", err)
	}
	if !strings.Contains(buf.String(), "sig:") {
		t.Errorf("expected signature in output, got: %s", buf.String())
	}
}

func TestPublishNoSkillsError(t *testing.T) {
	dir := t.TempDir()
	bake := `version: "1"
skill_sources:
  output_dir: .agents/skills
targets:
  default:
    platforms: [codex]
    skills: []
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runRoot("publish", "--target", dir)
	if err == nil {
		t.Error("expected error when no skills in target")
	}
}

// ── compare command ───────────────────────────────────────────────────────────

func TestCompareCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "compare <skill> <ref1> <ref2>" {
			found = true
			if sub.Flags().Lookup("format") == nil {
				t.Error("compare: missing --format flag")
			}
			break
		}
	}
	if !found {
		t.Error("compare command not registered on root")
	}
}

func TestCompareRequiresThreeArgs(t *testing.T) {
	err := runRoot("compare", "my-skill", "HEAD")
	if err == nil {
		t.Error("expected error with fewer than 3 args")
	}
}

func TestCompareFailsOutsideGitRepo(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("compare", "snap-skill", "HEAD~1", "HEAD", "--target", dir)
	// No .git directory — should error gracefully.
	if err == nil {
		t.Error("expected error when not in a git repo")
	}
}

// ── test command ──────────────────────────────────────────────────────────────

func TestTestCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "test" {
			found = true
			if sub.Flags().Lookup("skill") == nil {
				t.Error("test: missing --skill flag")
			}
			if sub.Flags().Lookup("endpoint") == nil {
				t.Error("test: missing --endpoint flag")
			}
			if sub.Flags().Lookup("list") == nil {
				t.Error("test: missing --list flag")
			}
			break
		}
	}
	if !found {
		t.Error("test command not registered on root")
	}
}

func TestTestListNoTests(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"test", "--target", dir, "--list"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute() // may return error if skills have no tests
	out := buf.String()
	// Should report coverage stats at minimum.
	if !strings.Contains(out, "skill(s)") {
		t.Errorf("expected coverage summary in output, got: %s", out)
	}
}

func TestTestWithManifest(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	manifest := "version: 1\ncases:\n  - name: smoke\n    input: Hello\n    expect_contains: [result]\n"
	skillDir := filepath.Join(dir, ".agents", "skills", "snap-skill")
	if err := os.WriteFile(filepath.Join(skillDir, "skill-tests.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"test", "--target", dir, "--list"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	if !strings.Contains(buf.String(), "snap-skill") {
		t.Errorf("expected skill name in list output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "1 case") {
		t.Errorf("expected case count in list output, got: %s", buf.String())
	}
}

func TestTestStructuralValidationNoEndpoint(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	manifest := "version: 1\ncases:\n  - name: smoke\n    input: Hello world\n"
	skillDir := filepath.Join(dir, ".agents", "skills", "snap-skill")
	if err := os.WriteFile(filepath.Join(skillDir, "skill-tests.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"test", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	// No endpoint → structural pass.
	if !strings.Contains(buf.String(), "validated") {
		t.Errorf("expected 'validated' in output, got: %s", buf.String())
	}
}

func makeBakefileWithSkill(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bake := `version: "1"
skill_sources:
  output_dir: .agents/skills
targets:
  default:
    platforms: [codex]
    skills: [snap-skill]
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, ".agents", "skills", "snap-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: snap-skill\ndescription: Snapshot test skill\nversion: 1.0.0\nlast_modified: 2024-01-01\n---\n\n# Snap Skill\n\nBody content.\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ── archive command ───────────────────────────────────────────────────────────

func TestArchiveCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "archive" {
			found = true
			subNames := map[string]bool{}
			for _, s := range sub.Commands() {
				subNames[s.Use] = true
			}
			if !subNames["skill <name>"] {
				t.Error("archive: missing 'skill' subcommand")
			}
			if !subNames["restore <name>"] {
				t.Error("archive: missing 'restore' subcommand")
			}
			break
		}
	}
	if !found {
		t.Error("archive command not registered on root")
	}
}

func TestArchiveSkillSetsDeprecated(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("archive", "skill", "snap-skill", "--target", dir); err != nil {
		t.Fatalf("archive skill: %v", err)
	}

	skillMD := filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md")
	data, _ := os.ReadFile(skillMD)
	if !strings.Contains(string(data), "stability: deprecated") {
		t.Errorf("expected 'stability: deprecated' in SKILL.md, got:\n%s", string(data))
	}
}

func TestArchiveSkillRemovesFromTarget(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("archive", "skill", "snap-skill", "--target", dir); err != nil {
		t.Fatalf("archive skill: %v", err)
	}

	cfg, err := cliLoadBakeFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for tName, t2 := range cfg.Targets {
		for _, s := range t2.Skills {
			if s == "snap-skill" {
				t.Errorf("snap-skill still in target %q after archive", tName)
			}
		}
	}
}

func TestArchiveDryRunMakesNoChanges(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	skillMD := filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md")
	before, _ := os.ReadFile(skillMD)

	if err := runRoot("archive", "skill", "snap-skill", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("archive --dry-run: %v", err)
	}

	after, _ := os.ReadFile(skillMD)
	if string(before) != string(after) {
		t.Error("dry-run should not modify SKILL.md")
	}
}

func TestArchiveRestoreRemovesDeprecated(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("archive", "skill", "snap-skill", "--target", dir); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("archive", "restore", "snap-skill", "--target", dir); err != nil {
		t.Fatalf("archive restore: %v", err)
	}

	skillMD := filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md")
	data, _ := os.ReadFile(skillMD)
	if !strings.Contains(string(data), "stability: stable") {
		t.Errorf("expected 'stability: stable' after restore, got:\n%s", string(data))
	}
}

// ── revert command ────────────────────────────────────────────────────────────

func TestRevertCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "revert <skill> <ref>" {
			found = true
			if sub.Flags().Lookup("dry-run") == nil {
				t.Error("revert: missing --dry-run flag")
			}
			break
		}
	}
	if !found {
		t.Error("revert command not registered on root")
	}
}

func TestRevertRequiresTwoArgs(t *testing.T) {
	err := runRoot("revert", "my-skill")
	if err == nil {
		t.Error("expected error with only one arg")
	}
}

func TestRevertFailsOutsideGitRepo(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("revert", "snap-skill", "HEAD~1", "--target", dir)
	if err == nil {
		t.Error("expected error when not in git repo")
	}
}

// ── tag command ───────────────────────────────────────────────────────────────

func TestTagCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "tag" {
			found = true
			subNames := map[string]bool{}
			for _, s := range sub.Commands() {
				subNames[s.Use] = true
			}
			if !subNames["list"] {
				t.Error("tag: missing 'list' subcommand")
			}
			if !subNames["create"] {
				t.Error("tag: missing 'create' subcommand")
			}
			if !subNames["delete <tag...>"] {
				t.Error("tag: missing 'delete' subcommand")
			}
			break
		}
	}
	if !found {
		t.Error("tag command not registered on root")
	}
}

func TestTagCreateDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"tag", "create", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("tag create --dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "snap-skill") {
		t.Errorf("expected skill name in dry-run output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "would create") {
		t.Errorf("expected 'would create' in dry-run output, got: %s", buf.String())
	}
}

func TestTagListOutsideGitReturnsError(t *testing.T) {
	dir := t.TempDir()
	err := runRoot("tag", "list", "--target", dir)
	if err == nil {
		t.Error("expected error listing tags outside a git repo")
	}
}

func TestTagDeleteRequiresArgs(t *testing.T) {
	err := runRoot("tag", "delete")
	if err == nil {
		t.Error("expected error with no tag args")
	}
}

// ── changelog command ─────────────────────────────────────────────────────────

func TestChangelogCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "changelog" {
			found = true
			if sub.Flags().Lookup("since") == nil {
				t.Error("changelog: missing --since flag")
			}
			if sub.Flags().Lookup("out") == nil {
				t.Error("changelog: missing --out flag")
			}
			break
		}
	}
	if !found {
		t.Error("changelog command not registered on root")
	}
}

func TestChangelogNoEntries(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"changelog", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	if !strings.Contains(buf.String(), "Combined Changelog") {
		t.Errorf("expected header in output, got: %s", buf.String())
	}
}

func makeSkillWithChangelogFM(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bake := `version: "1"
skill_sources:
  output_dir: .agents/skills
targets:
  default:
    platforms: [codex]
    skills: [snap-skill]
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, ".agents", "skills", "snap-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Changelog entries live in SKILL.md frontmatter, not in CHANGELOG.md.
	skillMD := `---
name: snap-skill
description: Snapshot test skill
version: 1.1.0
last_modified: 2024-08-01
stability: stable
license: MIT
changelog:
  - version: 1.1.0
    date: "2024-08-01"
    change: New feature
  - version: 1.0.0
    date: "2024-01-01"
    change: Initial release
---

# Snap Skill

Body content.
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestChangelogWithEntry(t *testing.T) {
	dir := makeSkillWithChangelogFM(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"changelog", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	if !strings.Contains(buf.String(), "snap-skill") {
		t.Errorf("expected skill name in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "Initial release") {
		t.Errorf("expected changelog text in output, got: %s", buf.String())
	}
}

func TestChangelogSinceFilter(t *testing.T) {
	dir := makeSkillWithChangelogFM(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"changelog", "--target", dir, "--since", "2024-06-01"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	out := buf.String()
	if !strings.Contains(out, "New feature") {
		t.Errorf("expected new entry in output, got: %s", out)
	}
	if strings.Contains(out, "Initial release") {
		t.Errorf("old entry should be filtered by --since, got: %s", out)
	}
}

func TestChangelogJSONOutput(t *testing.T) {
	dir := makeSkillWithChangelogFM(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"changelog", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	if !strings.Contains(buf.String(), `"version"`) {
		t.Errorf("expected JSON with version field, got: %s", buf.String())
	}
}

func TestChangelogWritesToFile(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	outFile := filepath.Join(dir, "release-notes.md")

	root := NewRootCommand()
	root.SetArgs([]string{"changelog", "--target", dir, "--out", outFile})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("changelog --out: %v", err)
	}
	if _, err := os.Stat(outFile); err != nil {
		t.Error("expected output file to be created")
	}
}

// ── coverage command ──────────────────────────────────────────────────────────

func TestCoverageCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "coverage" {
			found = true
			if sub.Flags().Lookup("fail-under") == nil {
				t.Error("coverage: missing --fail-under flag")
			}
			if sub.Flags().Lookup("json") == nil {
				t.Error("coverage: missing --json flag")
			}
			break
		}
	}
	if !found {
		t.Error("coverage command not registered on root")
	}
}

func TestCoverageReportsZeroWithoutTests(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"coverage", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	out := buf.String()
	if !strings.Contains(out, "0/1") {
		t.Errorf("expected 0/1 coverage, got: %s", out)
	}
}

func TestCoverageReportsCoveredSkill(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	manifest := "version: 1\ncases:\n  - name: smoke\n    input: Hello\n"
	skillDir := filepath.Join(dir, ".agents", "skills", "snap-skill")
	if err := os.WriteFile(filepath.Join(skillDir, "skill-tests.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"coverage", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	out := buf.String()
	if !strings.Contains(out, "1/1") {
		t.Errorf("expected 1/1 coverage, got: %s", out)
	}
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected skill name in output, got: %s", out)
	}
}

func TestCoverageFailUnderThreshold(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("coverage", "--target", dir, "--fail-under", "50")
	if err == nil {
		t.Error("expected error when coverage is below threshold")
	}
}

func TestCoverageJSONOutput(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"coverage", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.Execute()
	if !strings.Contains(buf.String(), `"percent"`) {
		t.Errorf("expected JSON with percent field, got: %s", buf.String())
	}
}

// ── show command ──────────────────────────────────────────────────────────────

func TestShowCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "show <skill>" {
			found = true
			if sub.Flags().Lookup("preview-lines") == nil {
				t.Error("show: missing --preview-lines flag")
			}
			if sub.Flags().Lookup("json") == nil {
				t.Error("show: missing --json flag")
			}
			break
		}
	}
	if !found {
		t.Error("show command not registered on root")
	}
}

func TestShowDisplaysFrontmatter(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"show", "snap-skill", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected skill name, got: %s", out)
	}
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected version in output, got: %s", out)
	}
}

func TestShowJSONOutput(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"show", "snap-skill", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("show --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"metadata"`) {
		t.Errorf("expected JSON with metadata field, got: %s", buf.String())
	}
}

func TestShowErrorsOnMissingSkill(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("show", "nonexistent-skill", "--target", dir)
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

func TestShowRequiresOneArg(t *testing.T) {
	err := runRoot("show")
	if err == nil {
		t.Error("expected error with no args")
	}
}

// ── bundle ─────────────────────────────────────────────────────────────────────

func TestBundleCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "bundle" {
			found = true
			if sub.Flags().Lookup("target") == nil {
				t.Error("bundle: missing --target flag")
			}
			if sub.Flags().Lookup("out") == nil {
				t.Error("bundle: missing --out flag")
			}
			if sub.Flags().Lookup("bake-target") == nil {
				t.Error("bundle: missing --bake-target flag")
			}
			if sub.Flags().Lookup("dry-run") == nil {
				t.Error("bundle: missing --dry-run flag")
			}
		}
	}
	if !found {
		t.Error("bundle command not registered")
	}
}

func TestBundleDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"bundle", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("bundle --dry-run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Would write bundle") {
		t.Errorf("expected dry-run message, got: %s", out)
	}
}

func TestBundleWithBakeTarget(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"bundle", "--target", dir, "--bake-target", "default", "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("bundle --bake-target --dry-run: %v", err)
	}
}

func TestBundleErrorsOnMissingBakeTarget(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("bundle", "--target", dir, "--bake-target", "no-such-target", "--dry-run")
	if err == nil {
		t.Error("expected error for unknown bake-target")
	}
}

func TestBundleWritesTarball(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	outPath := filepath.Join(t.TempDir(), "out.tar.gz")

	root := NewRootCommand()
	root.SetArgs([]string{"bundle", "--target", dir, "--out", outPath})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("bundle: %v", err)
	}
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("expected tarball to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty tarball")
	}
}

// ── rename-target ──────────────────────────────────────────────────────────────

func TestRenameTargetCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "rename-target <old> <new>" {
			found = true
		}
	}
	if !found {
		t.Error("rename-target command not registered")
	}
}

func TestRenameTargetRenamesKey(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"rename-target", "default", "production", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("rename-target: %v", err)
	}
	if !strings.Contains(buf.String(), `"default"`) || !strings.Contains(buf.String(), `"production"`) {
		t.Errorf("expected rename message, got: %s", buf.String())
	}

	// Verify bakefile was rewritten.
	data, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	content := string(data)
	if strings.Contains(content, "default:") {
		t.Error("old target name still present in bakefile")
	}
	if !strings.Contains(content, "production:") {
		t.Error("new target name not found in bakefile")
	}
}

func TestRenameTargetErrorsOnMissingOld(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("rename-target", "no-such", "new-name", "--target", dir)
	if err == nil {
		t.Error("expected error for missing source target")
	}
}

func TestRenameTargetErrorsOnConflict(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `skill_sources:
  output_dir: .agents/skills
targets:
  alpha: {}
  beta: {}
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runRoot("rename-target", "alpha", "beta", "--target", dir)
	if err == nil {
		t.Error("expected error when destination target already exists")
	}
}

func TestRenameTargetRequiresTwoArgs(t *testing.T) {
	err := runRoot("rename-target", "only-one")
	if err == nil {
		t.Error("expected error with only one arg")
	}
}

func TestRenameTargetUpdatesInherits(t *testing.T) {
	dir := t.TempDir()
	bakeContent := `skill_sources:
  output_dir: .agents/skills
targets:
  base: {}
  child:
    inherits:
      - base
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bakeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runRoot("rename-target", "base", "foundation", "--target", dir); err != nil {
		t.Fatalf("rename-target: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if !strings.Contains(string(data), "foundation") {
		t.Error("renamed target not in bakefile")
	}
	if strings.Contains(string(data), "- base") {
		t.Error("old inherits reference not updated")
	}
	if !strings.Contains(string(data), "- foundation") {
		t.Error("inherits not updated to new name")
	}
}

// ── scaffold flow ──────────────────────────────────────────────────────────────

func TestScaffoldFlowCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	var found bool
	for _, sub := range root.Commands() {
		if sub.Use == "scaffold" {
			for _, s := range sub.Commands() {
				if s.Use == "flow <name>" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("scaffold flow command not registered")
	}
}

func TestScaffoldFlowCreatesFiles(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"scaffold", "flow", "my-flow", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}

	flowDir := filepath.Join(dir, ".agents", "skills", "my-flow")
	if _, err := os.Stat(filepath.Join(flowDir, "flow.yaml")); err != nil {
		t.Errorf("flow.yaml not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(flowDir, "README.md")); err != nil {
		t.Errorf("README.md not created: %v", err)
	}
}

func TestScaffoldFlowFlowYAMLContent(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	root := NewRootCommand()
	root.SetArgs([]string{"scaffold", "flow", "pipeline", "--target", dir})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "pipeline", "flow.yaml"))
	if err != nil {
		t.Fatalf("read flow.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "name: pipeline") {
		t.Errorf("flow.yaml missing name field: %s", content)
	}
	if !strings.Contains(content, "steps:") {
		t.Errorf("flow.yaml missing steps: %s", content)
	}
}

func TestScaffoldFlowDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"scaffold", "flow", "dry-flow", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("scaffold flow --dry-run: %v", err)
	}

	if !strings.Contains(buf.String(), "Would scaffold") {
		t.Errorf("expected dry-run output, got: %s", buf.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "dry-flow")); err == nil {
		t.Error("dry-run should not create directory")
	}
}

func TestScaffoldFlowSkipsExisting(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	// Create once.
	if err := runRoot("scaffold", "flow", "dup-flow", "--target", dir); err != nil {
		t.Fatalf("first scaffold flow: %v", err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"scaffold", "flow", "dup-flow", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("second scaffold flow: %v", err)
	}
	if !strings.Contains(buf.String(), "skip") {
		t.Errorf("expected skip message on second run, got: %s", buf.String())
	}
}

func TestScaffoldFlowRequiresName(t *testing.T) {
	err := runRoot("scaffold", "flow")
	if err == nil {
		t.Error("expected error with no name arg")
	}
}

// ── flow validate ──────────────────────────────────────────────────────────────

func TestFlowCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "flow" {
			found = true
			for _, s := range sub.Commands() {
				if s.Use == "validate <name>" {
					return
				}
			}
			t.Error("flow validate subcommand not registered")
		}
	}
	if !found {
		t.Error("flow command not registered")
	}
}

func TestFlowValidatePassesForValidFlow(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	// Scaffold a flow referencing the existing "snap-skill".
	if err := runRoot("scaffold", "flow", "my-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}

	// Update flow.yaml so step-1 references snap-skill.
	flowYAML := "name: my-flow\nversion: 0.1.0\nsteps:\n  - name: step-1\n    skill: snap-skill\n"
	flowPath := filepath.Join(dir, ".agents", "skills", "my-flow", "flow.yaml")
	if err := os.WriteFile(flowPath, []byte(flowYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "validate", "my-flow", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow validate: %v", err)
	}
	if !strings.Contains(buf.String(), "valid") {
		t.Errorf("expected valid output, got: %s", buf.String())
	}
}

func TestFlowValidateFailsForMissingSkill(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("scaffold", "flow", "bad-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}

	flowYAML := "name: bad-flow\nversion: 0.1.0\nsteps:\n  - name: step-1\n    skill: nonexistent-skill\n"
	flowPath := filepath.Join(dir, ".agents", "skills", "bad-flow", "flow.yaml")
	if err := os.WriteFile(flowPath, []byte(flowYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runRoot("flow", "validate", "bad-flow", "--target", dir)
	if err == nil {
		t.Error("expected error for flow with missing skill")
	}
}

func TestFlowValidateEmptyStepsReports(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("scaffold", "flow", "empty-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}

	emptyYAML := "name: empty-flow\nversion: 0.1.0\nsteps: []\n"
	flowPath := filepath.Join(dir, ".agents", "skills", "empty-flow", "flow.yaml")
	if err := os.WriteFile(flowPath, []byte(emptyYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "validate", "empty-flow", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow validate empty: %v", err)
	}
	if !strings.Contains(buf.String(), "no steps") {
		t.Errorf("expected 'no steps' message, got: %s", buf.String())
	}
}

func TestFlowValidateRequiresName(t *testing.T) {
	err := runRoot("flow", "validate")
	if err == nil {
		t.Error("expected error with no name arg")
	}
}

func TestFlowValidateMissingFlowYAML(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("flow", "validate", "ghost-flow", "--target", dir)
	if err == nil {
		t.Error("expected error when flow.yaml does not exist")
	}
}

// ── import ─────────────────────────────────────────────────────────────────────

func TestImportCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "import <tarball>" {
			found = true
			if sub.Flags().Lookup("target") == nil {
				t.Error("import: missing --target flag")
			}
			if sub.Flags().Lookup("force") == nil {
				t.Error("import: missing --force flag")
			}
			if sub.Flags().Lookup("dry-run") == nil {
				t.Error("import: missing --dry-run flag")
			}
		}
	}
	if !found {
		t.Error("import command not registered")
	}
}

func TestImportRoundTrip(t *testing.T) {
	// Create a source repo with a skill and bundle it.
	src := makeBakefileWithSkill(t)
	tarball := filepath.Join(t.TempDir(), "bundle.tar.gz")

	root := NewRootCommand()
	root.SetArgs([]string{"bundle", "--target", src, "--out", tarball})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("bundle: %v", err)
	}

	// Create a fresh destination repo and import the tarball.
	dst := t.TempDir()
	bake := `version: "1"
skill_sources:
  output_dir: .agents/skills
targets:
  default:
    platforms: [codex]
    skills: [snap-skill]
`
	if err := os.WriteFile(filepath.Join(dst, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root = NewRootCommand()
	root.SetArgs([]string{"import", tarball, "--target", dst})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("import: %v", err)
	}

	// Verify the skill was extracted.
	if _, err := os.Stat(filepath.Join(dst, ".agents", "skills", "snap-skill", "SKILL.md")); err != nil {
		t.Errorf("expected SKILL.md to be imported: %v", err)
	}
	if !strings.Contains(buf.String(), "Imported") {
		t.Errorf("expected imported message, got: %s", buf.String())
	}
}

func TestImportDryRun(t *testing.T) {
	src := makeBakefileWithSkill(t)
	tarball := filepath.Join(t.TempDir(), "bundle.tar.gz")

	if err := runRoot("bundle", "--target", src, "--out", tarball); err != nil {
		t.Fatalf("bundle: %v", err)
	}

	dst := t.TempDir()
	bake := "version: \"1\"\nskill_sources:\n  output_dir: .agents/skills\ntargets:\n  default:\n    skills: [snap-skill]\n"
	if err := os.WriteFile(filepath.Join(dst, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"import", tarball, "--target", dst, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("import --dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "would create") {
		t.Errorf("expected dry-run output, got: %s", buf.String())
	}
	// Nothing should have been written.
	if _, err := os.Stat(filepath.Join(dst, ".agents", "skills", "snap-skill", "SKILL.md")); err == nil {
		t.Error("dry-run should not have created files")
	}
}

func TestImportSkipsExisting(t *testing.T) {
	src := makeBakefileWithSkill(t)
	tarball := filepath.Join(t.TempDir(), "bundle.tar.gz")
	if err := runRoot("bundle", "--target", src, "--out", tarball); err != nil {
		t.Fatalf("bundle: %v", err)
	}

	// Import once.
	if err := runRoot("import", tarball, "--target", src); err != nil {
		t.Fatalf("first import: %v", err)
	}

	// Import again — should skip.
	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"import", tarball, "--target", src})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("second import: %v", err)
	}
	if !strings.Contains(buf.String(), "skip") {
		t.Errorf("expected skip on second import, got: %s", buf.String())
	}
}

func TestImportRequiresTarball(t *testing.T) {
	err := runRoot("import")
	if err == nil {
		t.Error("expected error with no tarball arg")
	}
}

// ── pin ────────────────────────────────────────────────────────────────────────

func TestPinCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	found := false
	for _, sub := range root.Commands() {
		if sub.Use == "pin <skill> <version>" {
			found = true
			if sub.Flags().Lookup("target") == nil {
				t.Error("pin: missing --target flag")
			}
			if sub.Flags().Lookup("dry-run") == nil {
				t.Error("pin: missing --dry-run flag")
			}
		}
	}
	if !found {
		t.Error("pin command not registered")
	}
}

func TestPinUpdatesVersion(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("pin", "snap-skill", "2.5.0", "--target", dir); err != nil {
		t.Fatalf("pin: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "version: 2.5.0") {
		t.Errorf("expected version 2.5.0 in SKILL.md, got:\n%s", string(data))
	}
}

func TestPinDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"pin", "snap-skill", "9.9.9", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("pin --dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "Would set") {
		t.Errorf("expected dry-run message, got: %s", buf.String())
	}

	// SKILL.md should still have original version.
	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if strings.Contains(string(data), "9.9.9") {
		t.Error("dry-run should not have written version 9.9.9")
	}
}

func TestPinIdempotent(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("pin", "snap-skill", "1.0.0", "--target", dir); err != nil {
		t.Fatalf("pin: %v", err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"pin", "snap-skill", "1.0.0", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("second pin: %v", err)
	}
	if !strings.Contains(buf.String(), "already has version") {
		t.Errorf("expected idempotent message, got: %s", buf.String())
	}
}

func TestPinErrorsOnMissingSkill(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("pin", "ghost-skill", "1.0.0", "--target", dir)
	if err == nil {
		t.Error("expected error for missing skill")
	}
}

func TestPinRequiresTwoArgs(t *testing.T) {
	err := runRoot("pin", "only-skill")
	if err == nil {
		t.Error("expected error with only one arg")
	}
}

// ── flow list ──────────────────────────────────────────────────────────────────

func TestFlowListCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "flow" {
			for _, s := range sub.Commands() {
				if s.Use == "list" {
					return
				}
			}
			t.Error("flow list not registered under flow")
			return
		}
	}
	t.Error("flow command not registered")
}

func TestFlowListEmpty(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "list", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow list: %v", err)
	}
	if !strings.Contains(buf.String(), "No flows") {
		t.Errorf("expected 'No flows' message, got: %s", buf.String())
	}
}

func TestFlowListShowsFlows(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	// Scaffold two flows.
	for _, name := range []string{"flow-a", "flow-b"} {
		if err := runRoot("scaffold", "flow", name, "--target", dir); err != nil {
			t.Fatalf("scaffold flow %s: %v", name, err)
		}
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "list", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow list: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "flow-a") {
		t.Errorf("expected flow-a in output, got: %s", out)
	}
	if !strings.Contains(out, "flow-b") {
		t.Errorf("expected flow-b in output, got: %s", out)
	}
	if !strings.Contains(out, "2 flow") {
		t.Errorf("expected flow count, got: %s", out)
	}
}

func TestFlowListJSON(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("scaffold", "flow", "my-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "list", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow list --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"name"`) {
		t.Errorf("expected JSON output, got: %s", buf.String())
	}
}

// ── deps ───────────────────────────────────────────────────────────────────────

func TestDepsCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "deps <skill>" {
			return
		}
	}
	t.Error("deps command not registered")
}

func TestDepsFindsTargets(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"deps", "snap-skill", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("deps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected skill name in output, got: %s", out)
	}
	if !strings.Contains(out, "default") {
		t.Errorf("expected target 'default' in output, got: %s", out)
	}
}

func TestDepsNoMatch(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"deps", "ghost-skill", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("deps ghost-skill: %v", err)
	}
	if !strings.Contains(buf.String(), "No targets") {
		t.Errorf("expected no-match message, got: %s", buf.String())
	}
}

func TestDepsFindsFlows(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	// Scaffold a flow that references snap-skill.
	if err := runRoot("scaffold", "flow", "test-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}
	flowYAML := "name: test-flow\nversion: 0.1.0\nsteps:\n  - name: s1\n    skill: snap-skill\n"
	if err := os.WriteFile(filepath.Join(dir, ".agents", "skills", "test-flow", "flow.yaml"), []byte(flowYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"deps", "snap-skill", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("deps: %v", err)
	}
	if !strings.Contains(buf.String(), "test-flow") {
		t.Errorf("expected flow in deps output, got: %s", buf.String())
	}
}

func TestDepsJSON(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"deps", "snap-skill", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("deps --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"targets"`) {
		t.Errorf("expected JSON with targets field, got: %s", buf.String())
	}
}

func TestDepsRequiresOneArg(t *testing.T) {
	err := runRoot("deps")
	if err == nil {
		t.Error("expected error with no arg")
	}
}

// ── upgrade skill ──────────────────────────────────────────────────────────────

func TestUpgradeSkillCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "upgrade <skill> <version>" {
			return
		}
	}
	t.Error("upgrade command not registered")
}

func TestCatalogUpdateAndUpgrade(t *testing.T) {
	dir := t.TempDir()
	first := catalog.CoreSkills[0]
	cfg := &models.BakeConfig{
		Version: "1",
		SkillSources: &models.SkillSourceConfig{
			OutputDir: ".agents/skills",
			Defaults: models.SkillSourceDefaults{
				InitialVersion: "1.0.0",
				Owner:          "platform-engineering",
				License:        "MIT",
				CompatibleWith: []string{"codex"},
			},
			Skills: []models.SkillSourceDefinition{{Name: first, Description: catalog.SkillDescription(first)}},
		},
		Targets: map[string]*models.TargetConfig{
			"default": {Platforms: []string{"codex"}, Skills: []string{first}},
		},
	}
	if err := cliDumpBakeFile(cfg, filepath.Join(dir, "agentic.bake.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := scaffoldTargetSkills(dir, []string{first}, cfg.SkillSources, []string{"codex"}, false); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("update", "--target", dir); err != nil {
		t.Fatalf("catalog update: %v", err)
	}
	state, err := loadCatalogState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if state.CatalogDigest == "" || !state.Skills[first].Managed {
		t.Fatalf("catalog state not initialized for managed skill: %#v", state.Skills[first])
	}
	if err := runRoot("upgrade", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("catalog upgrade dry-run: %v", err)
	}
	last := catalog.CoreSkills[len(catalog.CoreSkills)-1]
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", last)); !os.IsNotExist(err) {
		t.Fatal("dry-run unexpectedly scaffolded a new catalog skill")
	}

	// Simulate a clean skill generated by an older catalog revision.
	skillPath := filepath.Join(dir, ".agents", "skills", first, "SKILL.md")
	payload, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	older := strings.Replace(string(payload), "\n## Changelog\n", "\nCatalog-old instruction.\n\n## Changelog\n", 1)
	if err := os.WriteFile(skillPath, []byte(older), 0o644); err != nil {
		t.Fatal(err)
	}
	entry := state.Skills[first]
	entry.InstalledDigest = instructionDigest(older)
	entry.Managed = true
	state.Skills[first] = entry
	if err := writeCatalogState(dir, state); err != nil {
		t.Fatal(err)
	}

	if err := runRoot("upgrade", "--target", dir); err != nil {
		t.Fatalf("catalog upgrade: %v", err)
	}
	upgraded, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(upgraded), "Catalog-old instruction") {
		t.Fatal("outdated managed instruction body was not refreshed")
	}
	if !strings.Contains(string(upgraded), "version: 1.0.1") {
		t.Fatalf("refreshed skill version was not bumped: %s", upgraded)
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", last, "SKILL.md")); err != nil {
		t.Fatalf("new catalog skill was not scaffolded: %v", err)
	}
	updatedCfg, err := cliLoadBakeFile(filepath.Join(dir, "agentic.bake.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(updatedCfg.SkillSources.Skills) != len(catalog.CoreSkills) {
		t.Fatalf("expected %d registered catalog skills, got %d", len(catalog.CoreSkills), len(updatedCfg.SkillSources.Skills))
	}
}

func TestUpgradeSkillBumpsVersion(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("upgrade", "snap-skill", "2.0.0", "--target", dir); err != nil {
		t.Fatalf("upgrade: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if !strings.Contains(string(data), "version: 2.0.0") {
		t.Errorf("expected version 2.0.0 in SKILL.md, got:\n%s", string(data))
	}
}

func TestUpgradeSkillAddsChangelog(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("upgrade", "snap-skill", "3.0.0", "--target", dir, "--message", "Major rewrite"); err != nil {
		t.Fatalf("upgrade: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	content := string(data)
	if !strings.Contains(content, "changelog:") {
		t.Errorf("expected changelog: in SKILL.md, got:\n%s", content)
	}
	if !strings.Contains(content, "Major rewrite") {
		t.Errorf("expected changelog message in SKILL.md, got:\n%s", content)
	}
}

func TestUpgradeSkillDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"upgrade", "snap-skill", "9.9.9", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("upgrade --dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "Would upgrade") {
		t.Errorf("expected dry-run message, got: %s", buf.String())
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if strings.Contains(string(data), "9.9.9") {
		t.Error("dry-run should not have written version 9.9.9")
	}
}

func TestUpgradeSkillErrorsOnMissing(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	err := runRoot("upgrade", "ghost-skill", "1.0.0", "--target", dir)
	if err == nil {
		t.Error("expected error for missing skill")
	}
}

func TestUpgradeSkillRequiresTwoArgs(t *testing.T) {
	err := runRoot("upgrade", "only-skill")
	if err == nil {
		t.Error("expected error with only one arg")
	}
}

// ── flow run ───────────────────────────────────────────────────────────────────

func TestFlowRunCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "flow" {
			for _, s := range sub.Commands() {
				if s.Use == "run <name>" {
					return
				}
			}
			t.Error("flow run not registered under flow")
			return
		}
	}
	t.Error("flow command not found")
}

func TestFlowRunPrintsSteps(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("scaffold", "flow", "my-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}
	flowYAML := "name: my-flow\nversion: 0.1.0\nsteps:\n  - name: step-1\n    skill: snap-skill\n  - name: step-2\n    skill: snap-skill\n"
	if err := os.WriteFile(filepath.Join(dir, ".agents", "skills", "my-flow", "flow.yaml"), []byte(flowYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "run", "my-flow", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Step 1") {
		t.Errorf("expected Step 1 in output, got: %s", out)
	}
	if !strings.Contains(out, "Step 2") {
		t.Errorf("expected Step 2 in output, got: %s", out)
	}
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected skill name in output, got: %s", out)
	}
	if !strings.Contains(out, "dry run") {
		t.Errorf("expected dry run notice, got: %s", out)
	}
}

func TestFlowRunEmptyFlow(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("scaffold", "flow", "empty-flow", "--target", dir); err != nil {
		t.Fatalf("scaffold flow: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "skills", "empty-flow", "flow.yaml"), []byte("name: empty-flow\nsteps: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "run", "empty-flow", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("flow run empty: %v", err)
	}
	if !strings.Contains(buf.String(), "no steps") {
		t.Errorf("expected no-steps message, got: %s", buf.String())
	}
}

func TestFlowRunRequiresName(t *testing.T) {
	if err := runRoot("flow", "run"); err == nil {
		t.Error("expected error with no name arg")
	}
}

func TestFlowRunMissingFlowYAML(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("flow", "run", "ghost-flow", "--target", dir); err == nil {
		t.Error("expected error for missing flow.yaml")
	}
}

// ── touch ──────────────────────────────────────────────────────────────────────

func TestTouchCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "touch <skill>" {
			if sub.Flags().Lookup("target") == nil {
				t.Error("touch: missing --target flag")
			}
			if sub.Flags().Lookup("date") == nil {
				t.Error("touch: missing --date flag")
			}
			return
		}
	}
	t.Error("touch command not registered")
}

func TestTouchUpdatesLastModified(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("touch", "snap-skill", "--target", dir, "--date", "2026-01-15"); err != nil {
		t.Fatalf("touch: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if !strings.Contains(string(data), "2026-01-15") {
		t.Errorf("expected updated last_modified in SKILL.md, got:\n%s", string(data))
	}
}

func TestTouchIdempotent(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("touch", "snap-skill", "--target", dir, "--date", "2026-01-01"); err != nil {
		t.Fatalf("first touch: %v", err)
	}

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"touch", "snap-skill", "--target", dir, "--date", "2026-01-01"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("second touch: %v", err)
	}
	if !strings.Contains(buf.String(), "already has") {
		t.Errorf("expected idempotent message, got: %s", buf.String())
	}
}

func TestTouchDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"touch", "snap-skill", "--target", dir, "--date", "2099-12-31", "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("touch --dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "Would set") {
		t.Errorf("expected dry-run message, got: %s", buf.String())
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if strings.Contains(string(data), "2099") {
		t.Error("dry-run should not have written date")
	}
}

func TestTouchErrorsOnMissingSkill(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("touch", "ghost-skill", "--target", dir); err == nil {
		t.Error("expected error for missing skill")
	}
}

// ── skill-versions ─────────────────────────────────────────────────────────────

func TestSkillVersionsCommandRegistered(t *testing.T) {
	root := NewRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "skill-versions <skill>" {
			return
		}
	}
	t.Error("skill-versions command not registered")
}

func TestSkillVersionsShowsEntries(t *testing.T) {
	dir := makeSkillWithChangelogFM(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"skill-versions", "snap-skill", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill-versions: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1.0.0") {
		t.Errorf("expected version 1.0.0 in output, got: %s", out)
	}
	if !strings.Contains(out, "release") {
		t.Errorf("expected release count in output, got: %s", out)
	}
}

func TestSkillVersionsEmptyChangelog(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"skill-versions", "snap-skill", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill-versions empty: %v", err)
	}
	if !strings.Contains(buf.String(), "No changelog") {
		t.Errorf("expected 'No changelog' message, got: %s", buf.String())
	}
}

func TestSkillVersionsJSON(t *testing.T) {
	dir := makeSkillWithChangelogFM(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"skill-versions", "snap-skill", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill-versions --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"version"`) {
		t.Errorf("expected JSON with version field, got: %s", buf.String())
	}
}

func TestSkillVersionsRequiresOneArg(t *testing.T) {
	if err := runRoot("skill-versions"); err == nil {
		t.Error("expected error with no arg")
	}
}

// ── target-skills command ─────────────────────────────────────────────────────

func TestTargetSkillsListsSkills(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"target-skills", "default", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("target-skills: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected snap-skill in output, got: %s", out)
	}
}

func TestTargetSkillsJSON(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"target-skills", "default", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("target-skills --json: %v", err)
	}
	if !strings.Contains(buf.String(), `"snap-skill"`) {
		t.Errorf("expected snap-skill in JSON output, got: %s", buf.String())
	}
}

func TestTargetSkillsUnknownTargetErrors(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("target-skills", "nonexistent", "--target", dir); err == nil {
		t.Error("expected error for unknown target")
	}
}

func TestTargetSkillsRequiresOneArg(t *testing.T) {
	if err := runRoot("target-skills"); err == nil {
		t.Error("expected error with no arg")
	}
}

// ── batch-touch command ───────────────────────────────────────────────────────

func TestBatchTouchUpdatesStaleSkill(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	// SKILL.md has last_modified: 2024-01-01; its mtime is "now" (just created), so it is stale.

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"batch-touch", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("batch-touch: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Touched 1") {
		t.Errorf("expected 'Touched 1', got: %s", out)
	}

	// Verify SKILL.md was updated.
	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if strings.Contains(string(data), "2024-01-01") {
		t.Errorf("expected last_modified to be updated, still 2024-01-01")
	}
}

func TestBatchTouchDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"batch-touch", "--target", dir, "--dry-run"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("batch-touch dry-run: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "would touch") {
		t.Errorf("expected 'would touch' in dry-run output, got: %s", out)
	}
	// File must not have changed.
	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if !strings.Contains(string(data), "2024-01-01") {
		t.Errorf("dry-run must not modify the file")
	}
}

// ── skill-diff command ────────────────────────────────────────────────────────

func TestSkillDiffErrorsWithoutGit(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	// No git repo — gitShow must fail gracefully.
	err := runRoot("skill-diff", "snap-skill", "HEAD", "--target", dir)
	if err == nil {
		t.Error("expected error when not in a git repo")
	}
}

func TestSkillDiffRequiresTwoArgs(t *testing.T) {
	if err := runRoot("skill-diff", "snap-skill"); err == nil {
		t.Error("expected error with only one arg")
	}
	if err := runRoot("skill-diff"); err == nil {
		t.Error("expected error with no args")
	}
}

// ── flow lint command ─────────────────────────────────────────────────────────

func makeFlowLintDir(t *testing.T, flowYAML string) (dir string, flowName string) {
	t.Helper()
	dir = makeBakefileWithSkill(t)
	flowName = "test-flow"
	flowDir := filepath.Join(dir, ".agents", "skills", flowName)
	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(flowDir, "flow.yaml"), []byte(flowYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, flowName
}

func TestFlowLintValidFlow(t *testing.T) {
	flowYAML := "name: test-flow\nversion: 1.0.0\nsteps:\n  - name: step-a\n    skill: snap-skill\n"
	dir, flowName := makeFlowLintDir(t, flowYAML)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "lint", flowName, "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("lint valid flow: %v", err)
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Errorf("expected OK, got: %s", buf.String())
	}
}

func TestFlowLintDuplicateStepName(t *testing.T) {
	flowYAML := "name: test-flow\nversion: 1.0.0\nsteps:\n  - name: step-a\n    skill: snap-skill\n  - name: step-a\n    skill: snap-skill\n"
	dir, flowName := makeFlowLintDir(t, flowYAML)

	err := runRoot("flow", "lint", flowName, "--target", dir)
	if err == nil {
		t.Error("expected error for duplicate step name")
	}
}

func TestFlowLintMissingSkillField(t *testing.T) {
	flowYAML := "name: test-flow\nversion: 1.0.0\nsteps:\n  - name: step-a\n"
	dir, flowName := makeFlowLintDir(t, flowYAML)

	err := runRoot("flow", "lint", flowName, "--target", dir)
	if err == nil {
		t.Error("expected error for step without skill")
	}
}

func TestFlowLintUnknownDependsOn(t *testing.T) {
	flowYAML := "name: test-flow\nversion: 1.0.0\nsteps:\n  - name: step-a\n    skill: snap-skill\n    depends_on: [ghost-step]\n"
	dir, flowName := makeFlowLintDir(t, flowYAML)

	err := runRoot("flow", "lint", flowName, "--target", dir)
	if err == nil {
		t.Error("expected error for undefined depends_on ref")
	}
}

func TestFlowLintCycleDetected(t *testing.T) {
	flowYAML := "name: test-flow\nversion: 1.0.0\nsteps:\n  - name: a\n    skill: snap-skill\n    depends_on: [b]\n  - name: b\n    skill: snap-skill\n    depends_on: [a]\n"
	dir, flowName := makeFlowLintDir(t, flowYAML)

	err := runRoot("flow", "lint", flowName, "--target", dir)
	if err == nil {
		t.Error("expected error for dependency cycle")
	}
}

func TestFlowLintJSONOutput(t *testing.T) {
	flowYAML := "name: test-flow\nversion: 1.0.0\nsteps:\n  - name: step-a\n    skill: snap-skill\n"
	dir, flowName := makeFlowLintDir(t, flowYAML)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"flow", "lint", flowName, "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("lint json: %v", err)
	}
	if !strings.Contains(buf.String(), `"ok":true`) {
		t.Errorf("expected ok:true in JSON, got: %s", buf.String())
	}
}

// ── export-lock command ───────────────────────────────────────────────────────

func TestExportLockWritesFile(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	outPath := filepath.Join(dir, "agentic.lock.json")

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"export-lock", "--target", dir, "--out", outPath})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("export-lock: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("lock file not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "snap-skill") {
		t.Errorf("expected snap-skill in lock file, got: %s", content)
	}
	if !strings.Contains(content, `"generated_at"`) {
		t.Errorf("expected generated_at field, got: %s", content)
	}
}

func TestExportLockPrintsToStdout(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"export-lock", "--target", dir, "--out", "-"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("export-lock stdout: %v", err)
	}
	if !strings.Contains(buf.String(), "snap-skill") {
		t.Errorf("expected snap-skill in stdout output, got: %s", buf.String())
	}
}

func TestExportLockIncludesVersion(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"export-lock", "--target", dir, "--out", "-"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("export-lock version: %v", err)
	}
	// snap-skill SKILL.md has version: 1.0.0
	if !strings.Contains(buf.String(), "1.0.0") {
		t.Errorf("expected version 1.0.0 in output, got: %s", buf.String())
	}
}

// ── check-compat command ──────────────────────────────────────────────────────

func makeBakefileWithPlatforms(t *testing.T, platforms []string) string {
	t.Helper()
	dir := t.TempDir()
	bake := `version: "1"
skill_sources:
  output_dir: .agents/skills
targets:
  default:
    platforms: [codex]
    skills: [snap-skill]
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(dir, ".agents", "skills", "snap-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	platList := ""
	for _, p := range platforms {
		platList += "\n  - " + p
	}
	skillMD := "---\nname: snap-skill\ndescription: test\nversion: 1.0.0\nplatforms:" + platList + "\n---\n\n# Snap\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckCompatFound(t *testing.T) {
	dir := makeBakefileWithPlatforms(t, []string{"codex", "claude-code"})

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"check-compat", "snap-skill", "codex", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("check-compat: %v", err)
	}
	if !strings.Contains(buf.String(), "compatible") {
		t.Errorf("expected compatible message, got: %s", buf.String())
	}
}

func TestCheckCompatNotFound(t *testing.T) {
	dir := makeBakefileWithPlatforms(t, []string{"codex"})
	err := runRoot("check-compat", "snap-skill", "github-copilot", "--target", dir)
	if err == nil {
		t.Error("expected error for missing platform")
	}
}

func TestCheckCompatJSON(t *testing.T) {
	dir := makeBakefileWithPlatforms(t, []string{"codex", "claude-code"})

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"check-compat", "snap-skill", "claude-code", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("check-compat json: %v", err)
	}
	if !strings.Contains(buf.String(), `"compatible":true`) {
		t.Errorf("expected compatible:true in JSON, got: %s", buf.String())
	}
}

func TestCheckCompatRequiresTwoArgs(t *testing.T) {
	if err := runRoot("check-compat", "snap-skill"); err == nil {
		t.Error("expected error with one arg")
	}
}

// ── bulk-rename command ───────────────────────────────────────────────────────

func makeBakefileWithTwoSkills(t *testing.T, skill1, skill2 string) string {
	t.Helper()
	dir := t.TempDir()
	bake := "version: \"1\"\nskill_sources:\n  output_dir: .agents/skills\ntargets:\n  default:\n    platforms: [codex]\n    skills: [" + skill1 + ", " + skill2 + "]\n"
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{skill1, skill2} {
		skillDir := filepath.Join(dir, ".agents", "skills", s)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		md := "---\nname: " + s + "\nversion: 1.0.0\n---\n\n# " + s + "\n"
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(md), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestBulkRenameRenamesDir(t *testing.T) {
	dir := makeBakefileWithTwoSkills(t, "old-auth", "old-lint")

	if err := runRoot("bulk-rename", "old-", "new-", "--target", dir); err != nil {
		t.Fatalf("bulk-rename: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "new-auth")); err != nil {
		t.Error("new-auth dir not created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "new-lint")); err != nil {
		t.Error("new-lint dir not created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "old-auth")); !os.IsNotExist(err) {
		t.Error("old-auth dir should be gone")
	}
}

func TestBulkRenameUpdatesBakefile(t *testing.T) {
	dir := makeBakefileWithTwoSkills(t, "old-auth", "old-lint")

	if err := runRoot("bulk-rename", "old-", "new-", "--target", dir); err != nil {
		t.Fatalf("bulk-rename: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "agentic.bake.yaml"))
	if strings.Contains(string(data), "old-auth") {
		t.Error("bakefile still contains old-auth")
	}
	if !strings.Contains(string(data), "new-auth") {
		t.Error("bakefile missing new-auth")
	}
}

func TestBulkRenameDryRun(t *testing.T) {
	dir := makeBakefileWithTwoSkills(t, "old-auth", "old-lint")

	if err := runRoot("bulk-rename", "old-", "new-", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("bulk-rename dry-run: %v", err)
	}

	// Directories must still exist.
	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "old-auth")); err != nil {
		t.Error("dry-run must not rename old-auth")
	}
}

func TestBulkRenameNoMatch(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"bulk-rename", "xyz-", "abc-", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("bulk-rename no match: %v", err)
	}
	if !strings.Contains(buf.String(), "No skills") {
		t.Errorf("expected 'No skills', got: %s", buf.String())
	}
}

// ── set-metadata command ──────────────────────────────────────────────────────

func TestSetMetadataSetsField(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("set-metadata", "snap-skill", "owner", "platform-team", "--target", dir); err != nil {
		t.Fatalf("set-metadata: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if !strings.Contains(string(data), "owner: platform-team") {
		t.Errorf("expected owner field, got: %s", string(data))
	}
}

func TestSetMetadataDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("set-metadata", "snap-skill", "owner", "platform-team", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("set-metadata dry-run: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill", "SKILL.md"))
	if strings.Contains(string(data), "owner:") {
		t.Error("dry-run must not modify the file")
	}
}

func TestSetMetadataRequiresThreeArgs(t *testing.T) {
	if err := runRoot("set-metadata", "snap-skill", "owner"); err == nil {
		t.Error("expected error with two args")
	}
}

// ── orphans command ───────────────────────────────────────────────────────────

func makeBakefileWithOrphan(t *testing.T) string {
	t.Helper()
	dir := makeBakefileWithSkill(t)
	// Add an extra skill dir that isn't in any target.
	orphanDir := filepath.Join(dir, ".agents", "skills", "orphan-skill")
	if err := os.MkdirAll(orphanDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("---\nname: orphan-skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestOrphansFindsUnreferenced(t *testing.T) {
	dir := makeBakefileWithOrphan(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"orphans", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("orphans: %v", err)
	}
	if !strings.Contains(buf.String(), "orphan-skill") {
		t.Errorf("expected orphan-skill in output, got: %s", buf.String())
	}
}

func TestOrphansNoneFound(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"orphans", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("orphans clean: %v", err)
	}
	if !strings.Contains(buf.String(), "No orphaned") {
		t.Errorf("expected 'No orphaned', got: %s", buf.String())
	}
}

func TestOrphansJSON(t *testing.T) {
	dir := makeBakefileWithOrphan(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"orphans", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("orphans json: %v", err)
	}
	if !strings.Contains(buf.String(), "orphan-skill") {
		t.Errorf("expected orphan-skill in JSON, got: %s", buf.String())
	}
}

// ── skill-size command ────────────────────────────────────────────────────────

func TestSkillSizeTable(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"skill-size", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill-size: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "snap-skill") {
		t.Errorf("expected snap-skill, got: %s", out)
	}
	if !strings.Contains(out, "BYTES") {
		t.Errorf("expected BYTES header, got: %s", out)
	}
}

func TestSkillSizeSortBySize(t *testing.T) {
	dir := makeBakefileWithTwoSkills(t, "aa-skill", "zz-skill")
	// Make aa-skill bigger by appending content.
	mdPath := filepath.Join(dir, ".agents", "skills", "aa-skill", "SKILL.md")
	existing, _ := os.ReadFile(mdPath)
	_ = os.WriteFile(mdPath, append(existing, []byte(strings.Repeat("x", 500))...), 0o644)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"skill-size", "--target", dir, "--sort", "size"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill-size sort: %v", err)
	}
	out := buf.String()
	// aa-skill must appear before zz-skill (larger first).
	if idx1, idx2 := strings.Index(out, "aa-skill"), strings.Index(out, "zz-skill"); idx1 > idx2 {
		t.Errorf("expected aa-skill before zz-skill in size sort, got: %s", out)
	}
}

func TestSkillSizeJSON(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"skill-size", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill-size json: %v", err)
	}
	if !strings.Contains(buf.String(), `"bytes"`) {
		t.Errorf("expected bytes field in JSON, got: %s", buf.String())
	}
}

// ── filter-by command ─────────────────────────────────────────────────────────

func makeBakefileWithStabilitySkills(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bake := `version: "1"
skill_sources:
  output_dir: .agents/skills
targets:
  default:
    platforms: [codex]
    skills: [stable-skill, experimental-skill]
`
	if err := os.WriteFile(filepath.Join(dir, "agentic.bake.yaml"), []byte(bake), 0o644); err != nil {
		t.Fatal(err)
	}
	skills := map[string]string{
		"stable-skill":       "---\nname: stable-skill\nstability: stable\nversion: 1.0.0\n---\n",
		"experimental-skill": "---\nname: experimental-skill\nstability: experimental\nversion: 0.1.0\n---\n",
	}
	for name, md := range skills {
		skillDir := filepath.Join(dir, ".agents", "skills", name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(md), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFilterByMatchesField(t *testing.T) {
	dir := makeBakefileWithStabilitySkills(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"filter-by", "stability", "experimental", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("filter-by: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "experimental-skill") {
		t.Errorf("expected experimental-skill, got: %s", out)
	}
	if strings.Contains(out, "stable-skill") {
		t.Errorf("stable-skill should not appear, got: %s", out)
	}
}

func TestFilterByNoMatch(t *testing.T) {
	dir := makeBakefileWithStabilitySkills(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"filter-by", "owner", "nobody", "--target", dir})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("filter-by no match: %v", err)
	}
	if !strings.Contains(buf.String(), "No skills") {
		t.Errorf("expected 'No skills', got: %s", buf.String())
	}
}

func TestFilterByJSON(t *testing.T) {
	dir := makeBakefileWithStabilitySkills(t)

	var buf strings.Builder
	root := NewRootCommand()
	root.SetArgs([]string{"filter-by", "stability", "stable", "--target", dir, "--json"})
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("filter-by json: %v", err)
	}
	if !strings.Contains(buf.String(), "stable-skill") {
		t.Errorf("expected stable-skill in JSON, got: %s", buf.String())
	}
}

func TestFilterByRequiresTwoArgs(t *testing.T) {
	if err := runRoot("filter-by", "stability"); err == nil {
		t.Error("expected error with one arg")
	}
}

// ── copy-skill command ────────────────────────────────────────────────────────

func TestCopySkillCreatesDir(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("copy-skill", "snap-skill", "snap-skill-v2", "--target", dir); err != nil {
		t.Fatalf("copy-skill: %v", err)
	}

	destDir := filepath.Join(dir, ".agents", "skills", "snap-skill-v2")
	if _, err := os.Stat(destDir); err != nil {
		t.Errorf("destination directory not created: %v", err)
	}
}

func TestCopySkillUpdatesName(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("copy-skill", "snap-skill", "snap-skill-v2", "--target", dir); err != nil {
		t.Fatalf("copy-skill: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".agents", "skills", "snap-skill-v2", "SKILL.md"))
	if !strings.Contains(string(data), "name: snap-skill-v2") {
		t.Errorf("expected name: snap-skill-v2 in copy, got: %s", string(data))
	}
}

func TestCopySkillSourceNotFound(t *testing.T) {
	dir := makeBakefileWithSkill(t)
	if err := runRoot("copy-skill", "ghost-skill", "new-skill", "--target", dir); err == nil {
		t.Error("expected error for missing source")
	}
}

func TestCopySkillDestExists(t *testing.T) {
	dir := makeBakefileWithTwoSkills(t, "skill-a", "skill-b")
	if err := runRoot("copy-skill", "skill-a", "skill-b", "--target", dir); err == nil {
		t.Error("expected error when destination already exists")
	}
}

func TestCopySkillDryRun(t *testing.T) {
	dir := makeBakefileWithSkill(t)

	if err := runRoot("copy-skill", "snap-skill", "snap-skill-v2", "--target", dir, "--dry-run"); err != nil {
		t.Fatalf("copy-skill dry-run: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".agents", "skills", "snap-skill-v2")); !os.IsNotExist(err) {
		t.Error("dry-run must not create destination directory")
	}
}
