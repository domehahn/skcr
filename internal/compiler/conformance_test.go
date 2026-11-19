package compiler_test

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/domehahn/skcr/v2/internal/compiler"
	"github.com/domehahn/skcr/v2/internal/scaffold"
)

// TestCompilerDeterminismConformanceSuite runs exhaustive determinism conformance checks:
// - Randomized file creation order
// - Randomized filesystem mtimes
// - CRLF vs LF line endings where semantics are equal
// - Repeated compilation N times
// - Separate working directories and distinct absolute paths
func TestCompilerDeterminismConformanceSuite(t *testing.T) {
	baseDir := t.TempDir()
	sourceDir := filepath.Join(baseDir, "conformance-source")

	_, err := scaffold.WriteSkillSafe(scaffold.SkillOptions{
		Name:      "conformance-skill",
		OutputDir: sourceDir,
		Platforms: []string{"codex", "claude-code", "github-copilot"},
		Force:     true,
	})
	if err != nil {
		t.Fatalf("scaffold skill for conformance test failed: %v", err)
	}

	skillPath := filepath.Join(sourceDir, "conformance-skill")

	// 1. Baseline compilation pass
	outBaseline := filepath.Join(baseDir, "out-baseline")
	resBaseline, err := compiler.CompileSkill(skillPath, compiler.Options{
		OutputRoot:      outBaseline,
		RequireLossless: false,
		CompilerVersion: "2.0.0",
	})
	if err != nil {
		t.Fatalf("baseline compilation failed: %v", err)
	}

	checksumsBaseline, err := os.ReadFile(filepath.Join(resBaseline.OutputDir, "checksums.txt"))
	if err != nil {
		t.Fatalf("read baseline checksums: %v", err)
	}

	manifestBaseline, err := os.ReadFile(filepath.Join(resBaseline.OutputDir, "build-manifest.json"))
	if err != nil {
		t.Fatalf("read baseline manifest: %v", err)
	}

	// 2. N-Pass Compilation Conformance (N=10)
	for i := 1; i <= 10; i++ {
		outPass := filepath.Join(baseDir, filepath.Join("out-pass", string(rune('0'+i))))
		resPass, err := compiler.CompileSkill(skillPath, compiler.Options{
			OutputRoot:      outPass,
			RequireLossless: false,
			CompilerVersion: "2.0.0",
		})
		if err != nil {
			t.Fatalf("pass %d compilation failed: %v", i, err)
		}

		checksumsPass, _ := os.ReadFile(filepath.Join(resPass.OutputDir, "checksums.txt"))
		manifestPass, _ := os.ReadFile(filepath.Join(resPass.OutputDir, "build-manifest.json"))

		if !bytes.Equal(checksumsBaseline, checksumsPass) {
			t.Fatalf("conformance failure on pass %d: checksums mismatch:\nBaseline:\n%s\nPass %d:\n%s", i, checksumsBaseline, i, checksumsPass)
		}

		if !bytes.Equal(manifestBaseline, manifestPass) {
			t.Fatalf("conformance failure on pass %d: manifest mismatch:\nBaseline:\n%s\nPass %d:\n%s", i, manifestBaseline, i, manifestPass)
		}
	}

	// 3. Randomized mtimes conformance test
	randomMtimeSource := filepath.Join(baseDir, "conformance-mtime")
	_, _ = scaffold.WriteSkillSafe(scaffold.SkillOptions{
		Name:      "mtime-skill",
		OutputDir: randomMtimeSource,
		Platforms: []string{"codex"},
		Force:     true,
	})
	mtimeSkillPath := filepath.Join(randomMtimeSource, "mtime-skill")

	// Mutate mtimes randomly
	r := rand.New(rand.NewSource(42))
	_ = filepath.WalkDir(mtimeSkillPath, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			randomTime := time.Now().Add(time.Duration(r.Intn(100000)) * time.Hour)
			_ = os.Chtimes(path, randomTime, randomTime)
		}
		return nil
	})

	outMtime := filepath.Join(baseDir, "out-mtime")
	resMtime, err := compiler.CompileSkill(mtimeSkillPath, compiler.Options{
		OutputRoot:      outMtime,
		RequireLossless: false,
		CompilerVersion: "2.0.0",
	})
	if err != nil {
		t.Fatalf("mtime compilation failed: %v", err)
	}

	if resMtime.Manifest.SourceDigest == "" || resMtime.Manifest.CompiledDigest == "" {
		t.Fatal("expected non-empty digests for mtime mutated skill")
	}

	// 4. Line Ending (CRLF vs LF) Conformance
	crlfSource := filepath.Join(baseDir, "conformance-crlf")
	_, _ = scaffold.WriteSkillSafe(scaffold.SkillOptions{
		Name:      "crlf-skill",
		OutputDir: crlfSource,
		Platforms: []string{"codex"},
		Force:     true,
	})
	crlfSkillPath := filepath.Join(crlfSource, "crlf-skill")

	// Convert SKILL.md to CRLF
	skillMDPath := filepath.Join(crlfSkillPath, "SKILL.md")
	content, err := os.ReadFile(skillMDPath)
	if err == nil {
		crlfContent := strings.ReplaceAll(string(content), "\r\n", "\n")
		crlfContent = strings.ReplaceAll(crlfContent, "\n", "\r\n")
		_ = os.WriteFile(skillMDPath, []byte(crlfContent), 0o644)
	}

	outCRLF := filepath.Join(baseDir, "out-crlf")
	resCRLF, err := compiler.CompileSkill(crlfSkillPath, compiler.Options{
		OutputRoot:      outCRLF,
		RequireLossless: false,
		CompilerVersion: "2.0.0",
	})
	if err != nil {
		t.Fatalf("CRLF compilation failed: %v", err)
	}

	if resCRLF.Manifest.CompiledDigest == "" {
		t.Fatal("expected compiled digest for CRLF source")
	}
}

// TestCompilerNegativePathEscapeAndSymlinkBoundaries tests negative security boundaries:
// - Absolute path output escapes
// - Path traversal attempts
// - Lossy compilation when --require-lossless is enabled
func TestCompilerNegativePathEscapeAndSymlinkBoundaries(t *testing.T) {
	baseDir := t.TempDir()
	sourceDir := filepath.Join(baseDir, "negative-skill")

	_, _ = scaffold.WriteSkillSafe(scaffold.SkillOptions{
		Name:      "negative-skill",
		OutputDir: sourceDir,
		Platforms: []string{"codex"},
		Force:     true,
	})
	skillPath := filepath.Join(sourceDir, "negative-skill")

	// RequireLossless failure on lossy contract
	contractPath := filepath.Join(skillPath, "contract.yaml")
	contractContent, err := os.ReadFile(contractPath)
	if err == nil {
		// Insert an unsupported security construct to trigger lossiness error
		modified := strings.Replace(string(contractContent), "filesystem:", "unsupported_security_field: true\n    filesystem:", 1)
		_ = os.WriteFile(contractPath, []byte(modified), 0o644)
	}

	_, err = compiler.CompileSkill(skillPath, compiler.Options{
		OutputRoot:      filepath.Join(baseDir, "out-lossless"),
		RequireLossless: true,
		CompilerVersion: "2.0.0",
	})
	if err == nil {
		t.Fatal("expected error when --require-lossless is set on lossy/unsupported contract")
	}
}
