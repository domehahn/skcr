package bake

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentic-template-kit/skcr/internal/models"
)

func TestLoadDumpAndResolveTarget(t *testing.T) {
	dir := t.TempDir()
	cfg := &models.BakeConfig{
		Targets: map[string]*models.TargetConfig{
			"base": {
				Description: "base",
				Platforms:   []string{"codex"},
				Skills:      []string{"safe-implementer"},
				Rules:       map[string]any{"a": true, "nested": map[string]any{"x": 1}},
				Model:       map[string]any{"m": "one"},
				GitLabDuo:   map[string]any{"slash_command": true},
			},
			"child": {
				Description: "child",
				Inherits:    []string{"base"},
				Platforms:   []string{"gitlab-duo"},
				Skills:      []string{"security-reviewer"},
				Rules:       map[string]any{"nested": map[string]any{"y": 2}},
				Model:       map[string]any{"m": "two"},
				GitLabDuo:   map[string]any{"slash_command": false},
			},
		},
	}
	path := filepath.Join(dir, "agentic.bake.yaml")
	if err := DumpBakeFile(cfg, path); err != nil {
		t.Fatalf("DumpBakeFile: %v", err)
	}
	loaded, err := LoadBakeFile(path)
	if err != nil {
		t.Fatalf("LoadBakeFile: %v", err)
	}
	resolved, err := ResolveTarget(loaded, "child")
	if err != nil {
		t.Fatalf("ResolveTarget: %v", err)
	}
	if len(resolved.Platforms) != 2 {
		t.Fatalf("expected merged platforms, got %#v", resolved.Platforms)
	}
	if got := resolved.Model["m"]; got != "two" {
		t.Fatalf("expected child override, got %v", got)
	}
	if _, ok := resolved.Rules["nested"].(map[string]any); !ok {
		t.Fatalf("expected nested merge, got %#v", resolved.Rules["nested"])
	}
	if got := resolved.GitLabDuo["slash_command"]; got != false {
		t.Fatalf("expected merged gitlab option false, got %v", got)
	}
}

func TestLoadBakeFileDefaultsAndErrors(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.yaml")
	if _, err := LoadBakeFile(missing); err == nil {
		t.Fatal("expected missing file error")
	}

	invalid := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(invalid, []byte("targets: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBakeFile(invalid); err == nil {
		t.Fatal("expected yaml error")
	}

	empty := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(empty, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadBakeFile(empty)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != "1" || cfg.Variables == nil || cfg.Targets == nil {
		t.Fatalf("expected defaults initialized, got %#v", cfg)
	}
}

func TestResolveTargetErrorsAndNormalization(t *testing.T) {
	cfg := &models.BakeConfig{Targets: map[string]*models.TargetConfig{}}
	if _, err := ResolveTarget(cfg, "x"); err == nil {
		t.Fatal("expected unknown target error")
	}

	cfg.Targets["a"] = &models.TargetConfig{Inherits: []string{"b"}}
	cfg.Targets["b"] = &models.TargetConfig{Inherits: []string{"a"}}
	if _, err := ResolveTarget(cfg, "a"); err == nil || !strings.Contains(err.Error(), "circular") {
		t.Fatalf("expected circular error, got %v", err)
	}

	cfg2 := &models.BakeConfig{Targets: map[string]*models.TargetConfig{"a": {Inherits: []string{"missing"}}}}
	if _, err := ResolveTarget(cfg2, "a"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing parent error, got %v", err)
	}

	normalizeTarget(nil)
	target := &models.TargetConfig{}
	normalizeTarget(target)
	if target.Inherits == nil || target.Platforms == nil || target.Rules == nil || target.GitLabDuo == nil {
		t.Fatalf("normalizeTarget did not initialize fields: %#v", target)
	}
}

func TestBuildInitialConfigVariants(t *testing.T) {
	if _, err := BuildInitialConfig(nil, "Demo", "team", "de", "standard", "unknown"); err == nil {
		t.Fatal("expected unsupported preset error")
	}

	cfg, err := BuildInitialConfig(nil, "Demo", "team", "de", "standard", "")
	if err != nil {
		t.Fatalf("BuildInitialConfig default: %v", err)
	}
	if cfg.Targets["default"] == nil || cfg.Targets["all"] == nil {
		t.Fatal("expected default/all targets")
	}
	if cfg.Targets["gitlab"].GitLabDuo["slash_command"] != true {
		t.Fatal("expected gitlab slash_command default true")
	}

	minimal, err := BuildInitialConfig(nil, "Demo", "team", "de", "standard", "minimal")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := minimal.Targets["codex"]; !ok {
		t.Fatal("expected codex target for minimal")
	}

	local, err := BuildInitialConfig(nil, "Demo", "team", "de", "standard", "local-ai")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := local.Targets["local-ai"]; !ok {
		t.Fatal("expected local-ai target")
	}

	if _, err := BuildInitialConfig([]string{"generic"}, "Demo", "team", "de", "standard", ""); err == nil {
		t.Fatal("expected no targets generated error for generic-only")
	}

	for _, preset := range []string{"gitlab", "enterprise", "all"} {
		if _, err := BuildInitialConfig(nil, "Demo", "team", "de", "standard", preset); err != nil {
			t.Fatalf("expected preset %q to work: %v", preset, err)
		}
	}
}

func TestDumpBakeFileMarshalError(t *testing.T) {
	orig := bakeYAMLMarshal
	t.Cleanup(func() { bakeYAMLMarshal = orig })
	bakeYAMLMarshal = func(any) ([]byte, error) { return nil, errors.New("boom") }
	err := DumpBakeFile(&models.BakeConfig{}, filepath.Join(t.TempDir(), "x.yaml"))
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}
