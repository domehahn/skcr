package compiler_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/domehahn/skcr/v2/internal/compiler"
)

func FuzzCompileSkill(f *testing.F) {
	// Seed corpus with minimal YAML structures
	f.Add([]byte("schema_version: \"2.0.0\"\nname: fuzz-skill\nversion: \"1.0.0\"\ndescription: fuzzing skill\ncontract: {file: contract.yaml}\n"))
	f.Add([]byte("schema_version: \"2.0.0\"\nname: ../../path-escape\nversion: \"1.0.0\"\ndescription: attack skill\ncontract: {file: contract.yaml}\n"))
	f.Add([]byte("invalid yaml payload"))

	f.Fuzz(func(t *testing.T, payload []byte) {
		tmpDir := t.TempDir()
		skillDir := filepath.Join(tmpDir, "fuzz-skill")
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return
		}

		// Write payload as descriptor.yaml
		_ = os.WriteFile(filepath.Join(skillDir, "descriptor.yaml"), payload, 0o644)
		_ = os.WriteFile(filepath.Join(skillDir, "contract.yaml"), payload, 0o644)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), payload, 0o644)

		outDir := filepath.Join(tmpDir, "out")

		// Compiler must never panic regardless of hostile input
		_, _ = compiler.CompileSkill(skillDir, compiler.Options{
			OutputRoot:      outDir,
			RequireLossless: false,
			CompilerVersion: "fuzz",
		})
	})
}
