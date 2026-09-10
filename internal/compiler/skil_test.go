package compiler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/skcr/v2/internal/compiler"
	"github.com/domehahn/skcr/v2/internal/scaffold"
	"github.com/domehahn/skcr/v2/internal/skillmeta"
	"gopkg.in/yaml.v3"
)

func sourceSkill(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files, err := scaffold.PlanSkill(scaffold.SkillOptions{Name: "complete-skill", OutputDir: root, Description: "Review a repository safely."})
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
	return filepath.Join(root, "complete-skill")
}

func TestCompileSkillProducesNativeSkilArtifacts(t *testing.T) {
	source := sourceSkill(t)
	result, err := compiler.CompileSkill(source, compiler.Options{OutputRoot: filepath.Join(t.TempDir(), "build")})
	if err != nil {
		t.Fatal(err)
	}
	contract, err := os.ReadFile(filepath.Join(result.OutputDir, "skill.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"version: 1", "skill:", "capabilities:", "delete: []", "confirm_destructive: true"} {
		if !strings.Contains(string(contract), want) {
			t.Errorf("compiled contract missing %q:\n%s", want, contract)
		}
	}
	for _, rel := range []string{"SKILL.md", "VERSION", "CHANGELOG.md", "checksums.txt", "build-manifest.json", filepath.Join("evals", "reject-read-only-bypass.yaml"), filepath.Join("evals", "respects-capability-boundaries.yaml")} {
		if _, err := os.Stat(filepath.Join(result.OutputDir, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	adversarial, err := os.ReadFile(filepath.Join(result.OutputDir, "evals", "reject-read-only-bypass.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type: adversarial", "category: permission-bypass", "required: true", "require_enforcement: true", "require_native_isolation: true", "containment_compliant", "filesystem.delete"} {
		if !strings.Contains(string(adversarial), want) {
			t.Errorf("compiled Eval v2 mapping missing %q:\n%s", want, adversarial)
		}
	}
	if result.Manifest.Source.DescriptorDigest == "" || result.Manifest.Source.IntegrationsDigest == "" || result.Manifest.Source.DependenciesDigest == "" || result.Manifest.Source.AssuranceDigest == "" || result.Manifest.TargetMetadata.ArtifactDigest == "" || result.Manifest.Provenance.SourceArtifactDigest == "" || result.Manifest.Provenance.MappingDigest == "" {
		t.Fatal("manifest digests must be populated")
	}
	if result.Manifest.SchemaVersion != "1.0.0" || result.Manifest.Target != "skil" || result.Manifest.BuildParametersDigest == "" {
		t.Fatalf("manifest identity contract failed: schema_version=%q, target=%q, params_digest=%q", result.Manifest.SchemaVersion, result.Manifest.Target, result.Manifest.BuildParametersDigest)
	}
}

func TestCompileSkillRejectsUnsupportedCommandArgumentMapping(t *testing.T) {
	source := sourceSkill(t)
	path := filepath.Join(source, "contract.yaml")
	contract, err := skillmeta.LoadContract(path)
	if err != nil {
		t.Fatal(err)
	}
	contract.Capabilities.Runtime.Required.Commands.Execute = boolPointer(true)
	contract.Capabilities.Runtime.Allowed.Commands.Execute = boolPointer(true)
	rule := skillmeta.CommandRule{Executable: "git", ArgvPrefix: []string{"diff"}}
	contract.Capabilities.Runtime.Required.Commands.Allow = []skillmeta.CommandRule{rule}
	contract.Capabilities.Runtime.Allowed.Commands.Allow = []skillmeta.CommandRule{rule}
	data, err := yaml.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = compiler.CompileSkill(source, compiler.Options{OutputRoot: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "unsupported security field") {
		t.Fatalf("expected unsupported security error, got %v", err)
	}
}

func boolPointer(value bool) *bool { return &value }

func TestCompileSkillRequireLosslessPreservesVerificationOnlyFields(t *testing.T) {
	result, err := compiler.CompileSkill(sourceSkill(t), compiler.Options{OutputRoot: t.TempDir(), RequireLossless: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Manifest.Mapping.VerificationOnly) == 0 {
		t.Fatal("expected preserved verification-only mappings")
	}
}

func TestCompileSkillRejectsInconsistentLegacySecuritySummary(t *testing.T) {
	source := sourceSkill(t)
	path := filepath.Join(source, "descriptor.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("security:\n  requires_network: true\n  requires_secrets: false\n  writes_files: false\n  runs_commands: false\n")...)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = compiler.CompileSkill(source, compiler.Options{OutputRoot: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "inconsistent security posture") {
		t.Fatalf("expected inconsistent security posture error, got %v", err)
	}
}

func TestCompileSkillDeterministicOutput(t *testing.T) {
	source := sourceSkill(t)
	outA := filepath.Join(t.TempDir(), "buildA")
	outB := filepath.Join(t.TempDir(), "buildB")

	resA, err := compiler.CompileSkill(source, compiler.Options{OutputRoot: outA, CompilerVersion: "1.0.0"})
	if err != nil {
		t.Fatalf("compile A failed: %v", err)
	}
	resB, err := compiler.CompileSkill(source, compiler.Options{OutputRoot: outB, CompilerVersion: "1.0.0"})
	if err != nil {
		t.Fatalf("compile B failed: %v", err)
	}

	checksumsA, err := os.ReadFile(filepath.Join(resA.OutputDir, "checksums.txt"))
	if err != nil {
		t.Fatalf("read checksums A: %v", err)
	}
	checksumsB, err := os.ReadFile(filepath.Join(resB.OutputDir, "checksums.txt"))
	if err != nil {
		t.Fatalf("read checksums B: %v", err)
	}

	if string(checksumsA) != string(checksumsB) {
		t.Fatalf("deterministic compilation failure: checksums mismatch:\nA:\n%s\nB:\n%s", checksumsA, checksumsB)
	}

	manifestA, err := os.ReadFile(filepath.Join(resA.OutputDir, "build-manifest.json"))
	if err != nil {
		t.Fatalf("read build-manifest A: %v", err)
	}
	manifestB, err := os.ReadFile(filepath.Join(resB.OutputDir, "build-manifest.json"))
	if err != nil {
		t.Fatalf("read build-manifest B: %v", err)
	}

	if string(manifestA) != string(manifestB) {
		t.Fatalf("deterministic compilation failure: build-manifest mismatch:\nA:\n%s\nB:\n%s", manifestA, manifestB)
	}
}
