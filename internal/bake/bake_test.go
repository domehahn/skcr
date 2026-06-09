package bake

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/domehahn/skcr/internal/models"
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
	if cfg.Skills == nil || cfg.Skills.Source != "agent-skills.lock" || cfg.Skills.Mode != models.SkillModeReference {
		t.Fatalf("expected default skill integration config, got %#v", cfg.Skills)
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

func TestLoadBakeFileWithSkillSources(t *testing.T) {
	dir := t.TempDir()
	content := `version: "1"
skill_sources:
  output_dir: custom-skills
  defaults:
    version: 0.2.0
    owner: team-a
    license: Apache-2.0
    compatible_with:
      - codex
      - gitlab
  skills:
    - name: my-skill
      description: A great skill.
      compatible_with:
        - claude
targets:
  default:
    platforms:
      - codex
`
	path := filepath.Join(dir, "agentic.bake.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadBakeFile(path)
	if err != nil {
		t.Fatalf("LoadBakeFile: %v", err)
	}
	if cfg.SkillSources == nil {
		t.Fatal("expected skill_sources to be loaded")
	}
	if cfg.SkillSources.OutputDir != "custom-skills" {
		t.Fatalf("unexpected output_dir: %q", cfg.SkillSources.OutputDir)
	}
	if cfg.SkillSources.Defaults.Version != "0.2.0" {
		t.Fatalf("unexpected defaults version: %q", cfg.SkillSources.Defaults.Version)
	}
	// Aliases should be normalized: gitlab -> gitlab-duo, claude -> claude-code.
	for _, p := range cfg.SkillSources.Defaults.CompatibleWith {
		if p == "gitlab" || p == "claude" {
			t.Fatalf("platform alias not normalized in defaults: %q", p)
		}
	}
	if len(cfg.SkillSources.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(cfg.SkillSources.Skills))
	}
	skill := cfg.SkillSources.Skills[0]
	if skill.Name != "my-skill" {
		t.Fatalf("unexpected skill name: %q", skill.Name)
	}
	for _, p := range skill.CompatibleWith {
		if p == "claude" {
			t.Fatalf("platform alias not normalized in skill: %q", p)
		}
	}
}

func TestLoadBakeFileSkillSourceDefaults(t *testing.T) {
	dir := t.TempDir()
	// skill_sources present but with minimal fields — defaults should be applied.
	content := `version: "1"
skill_sources:
  skills:
    - name: test-skill
targets:
  default:
    platforms:
      - codex
`
	path := filepath.Join(dir, "agentic.bake.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadBakeFile(path)
	if err != nil {
		t.Fatalf("LoadBakeFile: %v", err)
	}
	if cfg.SkillSources.OutputDir != ".agents/skills" {
		t.Fatalf("expected default output_dir '.agents/skills', got %q", cfg.SkillSources.OutputDir)
	}
	if cfg.SkillSources.Defaults.Version != "0.1.0" {
		t.Fatalf("expected default version '0.1.0', got %q", cfg.SkillSources.Defaults.Version)
	}
	if cfg.SkillSources.Defaults.License != "MIT" {
		t.Fatalf("expected default license 'MIT', got %q", cfg.SkillSources.Defaults.License)
	}
}

func TestLoadBakeFileNoSkillSources(t *testing.T) {
	// Existing bakefiles without skill_sources must still load cleanly.
	dir := t.TempDir()
	content := `version: "1"
targets:
  default:
    platforms:
      - codex
`
	path := filepath.Join(dir, "agentic.bake.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadBakeFile(path)
	if err != nil {
		t.Fatalf("LoadBakeFile: %v", err)
	}
	if cfg.SkillSources != nil {
		t.Fatalf("expected nil skill_sources for old-style bakefile, got %#v", cfg.SkillSources)
	}
}

func TestBuildInitialConfigHasSkillSources(t *testing.T) {
	cfg, err := BuildInitialConfig([]string{"codex"}, "Demo", "team-x", "en", "standard", "")
	if err != nil {
		t.Fatalf("BuildInitialConfig: %v", err)
	}
	if cfg.SkillSources == nil {
		t.Fatal("expected skill_sources block in generated config")
	}
	if cfg.SkillSources.Defaults.Version != "0.1.0" {
		t.Fatalf("expected default version '0.1.0', got %q", cfg.SkillSources.Defaults.Version)
	}
	if cfg.SkillSources.Defaults.Owner != "team-x" {
		t.Fatalf("expected owner 'team-x', got %q", cfg.SkillSources.Defaults.Owner)
	}
}
