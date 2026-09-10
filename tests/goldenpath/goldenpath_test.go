package goldenpath_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/domehahn/skcr/v2/internal/compiler"
	"github.com/domehahn/skcr/v2/internal/platforms"
	"github.com/domehahn/skcr/v2/internal/scaffold"
)

// TestSixRepositoryGoldenPathProducer verifies that skcr produces deterministic,
// contract-compliant source and compiled digests for toolchain consumption across
// skcr -> skil -> skgate -> skpm -> SkillForge -> skrun.
func TestSixRepositoryGoldenPathProducer(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "goldenpath-skill")

	// 1. Scaffold canonical skill fixture
	writeRes, err := scaffold.WriteSkillSafe(scaffold.SkillOptions{
		Name:      "security-reviewer",
		OutputDir: sourceDir,
		Platforms: []string{"codex", "claude-code", "github-copilot"},
		Force:     true,
	})
	if err != nil {
		t.Fatalf("scaffold Golden Path skill failed: %v", err)
	}
	if len(writeRes.Created) == 0 {
		t.Fatal("expected scaffolded files for Golden Path producer")
	}

	skillPath := filepath.Join(sourceDir, "security-reviewer")

	// 2. Compile source skill to skil target
	outDir := filepath.Join(tempDir, "compiled-output")
	res, err := compiler.CompileSkill(skillPath, compiler.Options{
		OutputRoot:      outDir,
		RequireLossless: false,
		CompilerVersion: "2.0.0",
	})
	if err != nil {
		t.Fatalf("compile Golden Path skill failed: %v", err)
	}

	// 3. Assert output directory structure
	if res.OutputDir == "" {
		t.Fatal("expected compiled OutputDir in result")
	}
	manifestPath := filepath.Join(res.OutputDir, "build-manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("missing build-manifest.json: %v", err)
	}

	checksumsPath := filepath.Join(res.OutputDir, "checksums.txt")
	checksumsData, err := os.ReadFile(checksumsPath)
	if err != nil {
		t.Fatalf("missing checksums.txt: %v", err)
	}
	if len(checksumsData) == 0 {
		t.Fatal("checksums.txt must not be empty")
	}

	// 4. Validate Build Identity Contract & digests for multi-repo toolchain consumption
	var manifest compiler.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("invalid build-manifest.json format: %v", err)
	}

	if manifest.SchemaVersion != "1.0.0" {
		t.Errorf("manifest.schema_version = %q, want \"1.0.0\"", manifest.SchemaVersion)
	}
	if manifest.Tool != "skcr" {
		t.Errorf("manifest.tool = %q, want \"skcr\"", manifest.Tool)
	}
	if manifest.Target != "skil" {
		t.Errorf("manifest.target = %q, want \"skil\"", manifest.Target)
	}
	if manifest.SourceDigest == "" || manifest.CompiledDigest == "" || manifest.BuildParametersDigest == "" {
		t.Fatalf("Golden Path digests missing: source=%q compiled=%q params=%q", manifest.SourceDigest, manifest.CompiledDigest, manifest.BuildParametersDigest)
	}

	// 5. Verify repeatable N-pass compilation produces byte-identical Golden Path digests
	outDirPass2 := filepath.Join(tempDir, "compiled-output-pass2")
	resPass2, err := compiler.CompileSkill(skillPath, compiler.Options{
		OutputRoot:      outDirPass2,
		RequireLossless: false,
		CompilerVersion: "2.0.0",
	})
	if err != nil {
		t.Fatalf("compile Pass 2 failed: %v", err)
	}

	if res.Manifest.SourceDigest != resPass2.Manifest.SourceDigest {
		t.Fatalf("Golden Path source_digest mismatch across compilation passes:\nPass 1: %s\nPass 2: %s", res.Manifest.SourceDigest, resPass2.Manifest.SourceDigest)
	}
	if res.Manifest.CompiledDigest != resPass2.Manifest.CompiledDigest {
		t.Fatalf("Golden Path compiled_digest mismatch across compilation passes:\nPass 1: %s\nPass 2: %s", res.Manifest.CompiledDigest, resPass2.Manifest.CompiledDigest)
	}
}

func TestCanonicalCapabilityConstantsExist(t *testing.T) {
	caps := []string{
		platforms.CapFilesystemRead,
		platforms.CapFilesystemWrite,
		platforms.CapFilesystemDelete,
		platforms.CapProcessExec,
		platforms.CapProcessSpawn,
		platforms.CapNetworkEgress,
		platforms.CapNetworkListen,
		platforms.CapSecretRead,
		platforms.CapToolInvoke,
		platforms.CapMCPInvoke,
	}

	for _, capName := range caps {
		if capName == "" {
			t.Errorf("expected non-empty canonical capability constant")
		}
	}
}
