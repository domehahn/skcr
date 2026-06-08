package renderer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentic-template-kit/skcr/internal/bake"
	"github.com/agentic-template-kit/skcr/internal/models"
)

func TestTemplateRootAndHelpers(t *testing.T) {
	root, err := templateRoot()
	if err != nil {
		t.Fatalf("templateRoot error: %v", err)
	}
	if stat, err := os.Stat(root); err != nil || !stat.IsDir() {
		t.Fatalf("expected template root dir, stat=%v err=%v", stat, err)
	}

	skills := skillMeta([]string{"security-reviewer"})
	if skills[0]["name"] != "security-reviewer" || skills[0]["description"] == "" {
		t.Fatalf("unexpected skill meta: %#v", skills[0])
	}

	if !gitlabSlashCommandEnabled(nil) {
		t.Fatal("nil target should default to true")
	}
	if !gitlabSlashCommandEnabled(&models.TargetConfig{}) {
		t.Fatal("missing config should default to true")
	}
	if !gitlabSlashCommandEnabled(&models.TargetConfig{GitLabDuo: map[string]any{}}) {
		t.Fatal("missing slash_command key should default to true")
	}
	if !gitlabSlashCommandEnabled(&models.TargetConfig{GitLabDuo: map[string]any{"slash_command": "yes"}}) {
		t.Fatal("non-bool should default to true")
	}
	if gitlabSlashCommandEnabled(&models.TargetConfig{GitLabDuo: map[string]any{"slash_command": false}}) {
		t.Fatal("expected false")
	}

	subs := claudeSubagentsForTarget(&models.TargetConfig{Skills: []string{"security-reviewer"}})
	if len(subs) == 0 {
		t.Fatal("expected at least one claude subagent")
	}
	if len(claudeSubagentsForTarget(&models.TargetConfig{Skills: []string{"unknown"}})) != 0 {
		t.Fatal("expected no subagents for unrelated skills")
	}
}

func TestRenderFilesAllPlatformsAndContent(t *testing.T) {
	cfg, err := bake.BuildInitialConfig([]string{"codex", "gitlab-duo", "github-copilot", "claude", "openhands", "opencode", "ollama", "generic"}, "Demo", "team", "de", "strict", "")
	if err != nil {
		t.Fatal(err)
	}
	target := cfg.Targets["default"]
	resolved, err := bake.ResolveTarget(cfg, "default")
	if err != nil {
		t.Fatal(err)
	}
	files, err := RenderFiles(cfg, resolved)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected rendered files")
	}
	paths := map[string]string{}
	for _, f := range files {
		paths[f.Destination] = f.Content
	}
	required := []string{
		"AGENTS.md",
		".agentic/codex/AGENTS.md",
		".agentic/gitlab-duo/AGENTS.md",
		".gitlab/duo/chat-rules.md",
		".agentic/claude/AGENTS.md",
		".claude/agents/security-reviewer.md",
		".github/copilot-instructions.md",
		".agentic/github-copilot/AGENTS.md",
		".agentic/openhands/AGENTS.md",
		".agentic/opencode/AGENTS.md",
		".ollama/Modelfile",
		".agentic/ollama/AGENTS.md",
	}
	for _, p := range required {
		if _, ok := paths[p]; !ok {
			t.Fatalf("missing rendered file: %s", p)
		}
	}

	if strings.Contains(paths["skills/security-reviewer/SKILL.md"], "slash-command: enabled") == false {
		t.Fatal("expected gitlab skill slash metadata by default")
	}

	resolved.GitLabDuo = map[string]any{"slash_command": false}
	files2, err := RenderFiles(cfg, resolved)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files2 {
		if f.Destination == "skills/security-reviewer/SKILL.md" && strings.Contains(f.Content, "slash-command: enabled") {
			t.Fatal("slash metadata should be disabled")
		}
	}

	_ = target
}

func TestRenderFilesErrorWhenTemplateMissing(t *testing.T) {
	root, err := templateRoot()
	if err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(root, "codex", "AGENTS.md.j2")
	rename := bad + ".bak"
	if err := os.Rename(bad, rename); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Rename(rename, bad) }()

	cfg, err := bake.BuildInitialConfig([]string{"codex"}, "Demo", "team", "de", "strict", "")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := bake.ResolveTarget(cfg, "default")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RenderFiles(cfg, resolved); err == nil {
		t.Fatal("expected render error when template is missing")
	}
}

func TestTemplateRootErrorBranches(t *testing.T) {
	origCaller := rendererCaller
	origStat := rendererStat
	t.Cleanup(func() {
		rendererCaller = origCaller
		rendererStat = origStat
	})

	rendererCaller = func(skip int) (uintptr, string, int, bool) { return 0, "", 0, false }
	if _, err := templateRoot(); err == nil {
		t.Fatal("expected caller failure")
	}

	rendererCaller = origCaller
	rendererStat = func(string) (os.FileInfo, error) { return nil, errors.New("stat fail") }
	if _, err := templateRoot(); err == nil || !strings.Contains(err.Error(), "stat fail") {
		t.Fatalf("expected stat error, got %v", err)
	}
}

func TestRenderFilesWithoutPlatformDocs(t *testing.T) {
	cfg := &models.BakeConfig{Variables: map[string]any{"project_name": "Demo"}}
	target := &models.TargetConfig{Platforms: []string{}, Skills: []string{}, Flows: []string{}, Rules: map[string]any{}, Model: map[string]any{}, GitLabDuo: map[string]any{}}
	files, err := RenderFiles(cfg, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files for empty platforms, got %d", len(files))
	}
}

func TestRenderFilesWithLockedSkillReferences(t *testing.T) {
	cfg, err := bake.BuildInitialConfig([]string{"codex"}, "Demo", "team", "de", "strict", "")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := bake.ResolveTarget(cfg, "default")
	if err != nil {
		t.Fatal(err)
	}
	files, err := RenderFilesWithOptions(cfg, resolved, Options{
		SkillsMode: models.SkillModeReference,
		LockedSkills: []map[string]any{{
			"name":    "secure-code-review",
			"version": "v1.2.3",
			"path":    ".agents/skills/secure-code-review/SKILL.md",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.Destination == ".agentic/codex/AGENTS.md" && strings.Contains(file.Content, "secure-code-review v1.2.3") {
			return
		}
	}
	t.Fatal("expected locked skill reference in codex AGENTS.md")
}

func TestRenderFilesExecuteError(t *testing.T) {
	root, err := templateRoot()
	if err != nil {
		t.Fatal(err)
	}
	tpl := filepath.Join(root, "codex", "AGENTS.md.j2")
	orig, err := os.ReadFile(tpl)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tpl, []byte("{{ 1 / 0 }}"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.WriteFile(tpl, orig, 0o644) }()

	cfg, err := bake.BuildInitialConfig([]string{"codex"}, "Demo", "team", "de", "strict", "")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := bake.ResolveTarget(cfg, "default")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RenderFiles(cfg, resolved); err == nil {
		t.Fatal("expected execute error due invalid template syntax")
	}
}

func TestRenderFilesTemplateRootError(t *testing.T) {
	origCaller := rendererCaller
	rendererCaller = func(skip int) (uintptr, string, int, bool) { return 0, "", 0, false }
	defer func() { rendererCaller = origCaller }()

	cfg := &models.BakeConfig{Variables: map[string]any{"project_name": "Demo"}}
	target := &models.TargetConfig{Platforms: []string{"codex"}, Skills: []string{"security-reviewer"}}
	if _, err := RenderFiles(cfg, target); err == nil {
		t.Fatal("expected render files to fail when template root fails")
	}
}

func TestRenderFilesPerPlatformAddErrorPaths(t *testing.T) {
	type tc struct {
		name      string
		platforms []string
		brokenTpl string
		cfg       *models.BakeConfig
		target    *models.TargetConfig
	}

	cases := []tc{
		{name: "codex-main", platforms: []string{"codex"}, brokenTpl: "codex/AGENTS.md.j2"},
		{name: "codex-skill", platforms: []string{"codex"}, brokenTpl: "shared/SKILL.md.j2"},
		{name: "gitlab-main", platforms: []string{"gitlab-duo"}, brokenTpl: "gitlab-duo/AGENTS.md.j2"},
		{name: "gitlab-chat", platforms: []string{"gitlab-duo"}, brokenTpl: "gitlab-duo/chat-rules.md.j2"},
		{name: "gitlab-skill", platforms: []string{"gitlab-duo"}, brokenTpl: "shared/SKILL.md.j2"},
		{name: "gitlab-flow-readme", platforms: []string{"gitlab-duo"}, brokenTpl: "gitlab-duo/flows/README.md.j2"},
		{name: "gitlab-flow-yaml", platforms: []string{"gitlab-duo"}, brokenTpl: "gitlab-duo/flows/flow.yaml.j2"},
		{name: "claude-main", platforms: []string{"claude"}, brokenTpl: "claude/CLAUDE.md.j2"},
		{name: "claude-skill", platforms: []string{"claude"}, brokenTpl: "shared/SKILL.md.j2"},
		{name: "claude-subagent", platforms: []string{"claude"}, brokenTpl: "claude/agent.md.j2"},
		{name: "copilot-main", platforms: []string{"github-copilot"}, brokenTpl: "github-copilot/copilot-instructions.md.j2"},
		{name: "copilot-ref", platforms: []string{"github-copilot"}, brokenTpl: "shared/platform-reference.md.j2"},
		{name: "copilot-prompt", platforms: []string{"github-copilot"}, brokenTpl: "github-copilot/prompt.prompt.md.j2"},
		{name: "openhands-main", platforms: []string{"openhands"}, brokenTpl: "openhands/AGENTS.md.j2"},
		{name: "openhands-instructions", platforms: []string{"openhands"}, brokenTpl: "openhands/instructions.md.j2"},
		{name: "opencode-main", platforms: []string{"opencode"}, brokenTpl: "opencode/AGENTS.md.j2"},
		{name: "opencode-instructions", platforms: []string{"opencode"}, brokenTpl: "opencode/instructions.md.j2"},
		{name: "ollama-main", platforms: []string{"ollama"}, brokenTpl: "ollama/Modelfile.j2"},
		{name: "ollama-readme", platforms: []string{"ollama"}, brokenTpl: "ollama/README.md.j2"},
		{name: "ollama-ref", platforms: []string{"ollama"}, brokenTpl: "shared/platform-reference.md.j2"},
		{name: "shared-index", platforms: []string{"codex"}, brokenTpl: "shared/AGENTS.index.md.j2"},
		{
			name:      "generic-skill",
			brokenTpl: "shared/SKILL.md.j2",
			cfg:       &models.BakeConfig{Variables: map[string]any{"project_name": "Demo"}},
			target: &models.TargetConfig{
				Platforms: []string{"generic"},
				Skills:    []string{"security-reviewer"},
				Rules:     map[string]any{},
				Model:     map[string]any{},
				Flows:     []string{},
				GitLabDuo: map[string]any{},
			},
		},
	}

	root, err := templateRoot()
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tplPath := filepath.Join(root, c.brokenTpl)
			backup := tplPath + ".bak"
			if err := os.Rename(tplPath, backup); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Rename(backup, tplPath) })

			cfg := c.cfg
			target := c.target
			if cfg == nil || target == nil {
				cfg, err = bake.BuildInitialConfig(c.platforms, "Demo", "team", "de", "strict", "")
				if err != nil {
					t.Fatal(err)
				}
				target, err = bake.ResolveTarget(cfg, "default")
				if err != nil {
					t.Fatal(err)
				}
			}

			if _, err := RenderFiles(cfg, target); err == nil {
				t.Fatalf("expected error with missing template %s", c.brokenTpl)
			}
		})
	}
}
