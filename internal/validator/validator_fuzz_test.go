package validator_test

import (
	"path/filepath"
	"testing"

	"github.com/domehahn/skcr/v2/internal/validator"
)

func FuzzPathNormalization(f *testing.F) {
	f.Add("normal/path/file.txt")
	f.Add("../../etc/passwd")
	f.Add("/absolute/path/escape")
	f.Add(".\\windows\\path\\traversal")

	f.Fuzz(func(t *testing.T, relPath string) {
		tmpDir := t.TempDir()
		out, err := validator.SanitizeRelativePath(tmpDir, relPath)
		if err == nil {
			// If no error, out MUST be strictly within tmpDir
			rel, err := filepath.Rel(tmpDir, out)
			if err != nil || rel == ".." || filepath.IsAbs(rel) {
				t.Fatalf("SanitizeRelativePath escaped boundary: base=%s input=%s out=%s", tmpDir, relPath, out)
			}
		}
	})
}
