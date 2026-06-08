package lockfile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/models"
)

func TestShaAndWriteLoadLockfile(t *testing.T) {
	if got := Sha256Text("abc"); !strings.HasPrefix(got, "sha256:") {
		t.Fatalf("unexpected hash prefix: %q", got)
	}

	dir := t.TempDir()
	files := []models.RenderedFile{{Source: "s", Destination: "a.txt", Content: "hello", Platform: "codex"}}
	if err := WriteLockfile(dir, files, "default"); err != nil {
		t.Fatalf("WriteLockfile: %v", err)
	}
	loaded, err := LoadLockfile(dir)
	if err != nil {
		t.Fatalf("LoadLockfile: %v", err)
	}
	if loaded["target"] != "default" {
		t.Fatalf("unexpected target: %v", loaded["target"])
	}
	byPath := ManagedFilesByPath(loaded)
	if _, ok := byPath["a.txt"]; !ok {
		t.Fatalf("expected a.txt in managed files, got %#v", byPath)
	}
}

func TestLoadLockfileEdgeCasesAndManagedFilesByPath(t *testing.T) {
	dir := t.TempDir()
	missing, err := LoadLockfile(dir)
	if err != nil || len(missing) != 0 {
		t.Fatalf("expected empty missing lockfile, got=%v err=%v", missing, err)
	}

	invalid := filepath.Join(dir, LockfileName)
	if err := os.WriteFile(invalid, []byte("["), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadLockfile(dir); err == nil {
		t.Fatal("expected yaml unmarshal error")
	}

	origMarshal := lockYAMLMarshal
	lockYAMLMarshal = func(any) ([]byte, error) { return nil, errors.New("marshal failed") }
	if err := WriteLockfile(dir, nil, "x"); err == nil || !strings.Contains(err.Error(), "marshal failed") {
		t.Fatalf("expected marshal fail, got %v", err)
	}
	lockYAMLMarshal = origMarshal

	if err := os.WriteFile(invalid, []byte("null\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadLockfile(dir)
	if err != nil || len(loaded) != 0 {
		t.Fatalf("expected empty map for null, got=%v err=%v", loaded, err)
	}

	origUnmarshal := lockYAMLUnmarshal
	lockYAMLUnmarshal = func([]byte, any) error { return errors.New("unmarshal failed") }
	if _, err := LoadLockfile(dir); err == nil || !strings.Contains(err.Error(), "unmarshal failed") {
		t.Fatalf("expected forced unmarshal error, got %v", err)
	}
	lockYAMLUnmarshal = origUnmarshal

	m := ManagedFilesByPath(map[string]any{"managed_files": []any{
		"bad",
		map[string]any{"path": "", "platform": "x"},
		map[string]any{"path": "ok.txt", "platform": "codex"},
	}})
	if len(m) != 1 || m["ok.txt"] == nil {
		t.Fatalf("unexpected managed file map: %#v", m)
	}

	if len(ManagedFilesByPath(map[string]any{"managed_files": "nope"})) != 0 {
		t.Fatal("expected empty map for invalid managed_files type")
	}

	if _, err := LoadLockfile(string([]byte{0})); err == nil {
		t.Fatal("expected os.Stat error for invalid target path")
	}

	readErrDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(readErrDir, LockfileName), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadLockfile(readErrDir); err == nil {
		t.Fatal("expected read error when lockfile path is a directory")
	}
}
