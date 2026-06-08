package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/domehahn/skcr/internal/catalog"
	"github.com/domehahn/skcr/internal/models"
	"github.com/flosch/pongo2/v6"
)

var (
	rendererCaller = runtime.Caller
	rendererStat   = os.Stat
)

type Options struct {
	LockedSkills []map[string]any
	SkillsMode   string
}

func templateRoot() (string, error) {
	_, file, _, ok := rendererCaller(0)
	if !ok {
		return "", fmt.Errorf("unable to determine template root")
	}
	root := filepath.Join(filepath.Dir(file), "templates")
	if _, err := rendererStat(root); err != nil {
		return "", err
	}
	return root, nil
}

func skillMeta(skillNames []string) []map[string]any {
	items := make([]map[string]any, 0, len(skillNames))
	for _, skill := range skillNames {
		items = append(items, map[string]any{
			"name":        skill,
			"title":       catalog.SkillTitle(skill),
			"description": catalog.SkillDescription(skill),
		})
	}
	return items
}

func gitlabSlashCommandEnabled(target *models.TargetConfig) bool {
	if target == nil || target.GitLabDuo == nil {
		return true
	}
	value, ok := target.GitLabDuo["slash_command"]
	if !ok {
		return true
	}
	enabled, ok := value.(bool)
	if !ok {
		return true
	}
	return enabled
}

func claudeSubagentsForTarget(target *models.TargetConfig) []map[string]any {
	available := map[string]struct{}{}
	for _, s := range target.Skills {
		available[s] = struct{}{}
	}

	selected := []map[string]any{}
	for _, agent := range catalog.ClaudeSubagents {
		agentSkills := []string{}
		for _, s := range agent.Skills {
			if _, ok := available[s]; ok {
				agentSkills = append(agentSkills, s)
			}
		}
		if len(agentSkills) == 0 {
			continue
		}
		selected = append(selected, map[string]any{
			"name":           agent.Name,
			"title":          catalog.SkillTitle(agent.Name),
			"description":    agent.Description,
			"tools":          agent.Tools,
			"model":          agent.Model,
			"permissionMode": agent.PermissionMode,
			"maxTurns":       agent.MaxTurns,
			"skills":         agentSkills,
			"prompt":         agent.Prompt,
		})
	}
	return selected
}

func RenderFiles(config *models.BakeConfig, target *models.TargetConfig) ([]models.RenderedFile, error) {
	return RenderFilesWithOptions(config, target, Options{})
}

func RenderFilesWithOptions(config *models.BakeConfig, target *models.TargetConfig, opts Options) ([]models.RenderedFile, error) {
	if opts.SkillsMode == "" {
		opts.SkillsMode = models.SkillModeReference
	}
	if opts.LockedSkills == nil {
		opts.LockedSkills = []map[string]any{}
	}
	root, err := templateRoot()
	if err != nil {
		return nil, err
	}
	set := pongo2.NewSet("skcr", pongo2.MustNewLocalFileSystemLoader(root))

	skills := skillMeta(target.Skills)
	claudeSubagents := claudeSubagentsForTarget(target)

	baseContext := pongo2.Context{
		"variables":        config.Variables,
		"target":           target,
		"skills":           skills,
		"claude_subagents": claudeSubagents,
		"flows":            target.Flows,
		"rules":            target.Rules,
		"model":            target.Model,
		"locked_skills":    opts.LockedSkills,
		"skills_mode":      opts.SkillsMode,
	}

	files := []models.RenderedFile{}
	platformDocs := []map[string]any{}

	add := func(platform, template, destination string, extra map[string]any) error {
		ctx := pongo2.Context{}
		for k, v := range baseContext {
			ctx[k] = v
		}
		for k, v := range extra {
			ctx[k] = v
		}
		tpl, err := set.FromFile(template)
		if err != nil {
			return err
		}
		content, err := safeExecuteTemplate(tpl, ctx)
		if err != nil {
			return err
		}
		files = append(files, models.RenderedFile{
			Source:      template,
			Destination: destination,
			Content:     content,
			Platform:    platform,
		})
		return nil
	}

	contains := func(platform string) bool {
		for _, p := range target.Platforms {
			if p == platform {
				return true
			}
		}
		return false
	}

	if contains("codex") {
		if err := add("codex", "codex/AGENTS.md.j2", ".agentic/codex/AGENTS.md", nil); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "codex", "path": ".agentic/codex/AGENTS.md", "description": "Codex instructions and skill routing"})
		for _, skill := range skills {
			extra := map[string]any{"skill": skill, "invocation_prefix": "$", "slash_command": false}
			if err := add("codex", "shared/SKILL.md.j2", fmt.Sprintf(".agents/skills/%s/SKILL.md", skill["name"]), extra); err != nil {
				return nil, err
			}
		}
	}

	if contains("gitlab-duo") {
		if err := add("gitlab-duo", "gitlab-duo/AGENTS.md.j2", ".agentic/gitlab-duo/AGENTS.md", nil); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "gitlab-duo", "path": ".agentic/gitlab-duo/AGENTS.md", "description": "GitLab Duo instructions, skills, and flow context requirements"})
		if err := add("gitlab-duo", "gitlab-duo/chat-rules.md.j2", ".gitlab/duo/chat-rules.md", nil); err != nil {
			return nil, err
		}
		slashCommandEnabled := gitlabSlashCommandEnabled(target)
		for _, skill := range skills {
			extra := map[string]any{"skill": skill, "invocation_prefix": "/", "slash_command": slashCommandEnabled}
			if err := add("gitlab-duo", "shared/SKILL.md.j2", fmt.Sprintf("skills/%s/SKILL.md", skill["name"]), extra); err != nil {
				return nil, err
			}
		}
		if err := add("gitlab-duo", "gitlab-duo/flows/README.md.j2", ".gitlab/duo/flows/README.md", nil); err != nil {
			return nil, err
		}
		for _, flow := range target.Flows {
			extra := map[string]any{"flow_name": flow, "flow_title": catalog.SkillTitle(flow)}
			if err := add("gitlab-duo", "gitlab-duo/flows/flow.yaml.j2", fmt.Sprintf(".gitlab/duo/flows/%s.yaml", flow), extra); err != nil {
				return nil, err
			}
		}
	}

	if contains("claude-code") {
		if err := add("claude", "claude/CLAUDE.md.j2", ".agentic/claude/AGENTS.md", nil); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "claude", "path": ".agentic/claude/AGENTS.md", "description": "Claude Code integration, skills, and subagent routing"})
		for _, skill := range skills {
			extra := map[string]any{"skill": skill, "invocation_prefix": "/", "slash_command": false}
			if err := add("claude", "shared/SKILL.md.j2", fmt.Sprintf(".claude/skills/%s/SKILL.md", skill["name"]), extra); err != nil {
				return nil, err
			}
		}
		for _, sub := range claudeSubagents {
			if err := add("claude", "claude/agent.md.j2", fmt.Sprintf(".claude/agents/%s.md", sub["name"]), map[string]any{"subagent": sub}); err != nil {
				return nil, err
			}
		}
	}

	if contains("github-copilot") {
		if err := add("github-copilot", "github-copilot/copilot-instructions.md.j2", ".github/copilot-instructions.md", nil); err != nil {
			return nil, err
		}
		extra := map[string]any{"platform_label": "GitHub Copilot", "target_file": ".github/copilot-instructions.md", "target_description": "Repository-level Copilot instructions"}
		if err := add("github-copilot", "shared/platform-reference.md.j2", ".agentic/github-copilot/AGENTS.md", extra); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "github-copilot", "path": ".agentic/github-copilot/AGENTS.md", "description": "Pointer to GitHub Copilot repository instructions and prompts"})
		for _, skill := range skills {
			if err := add("github-copilot", "github-copilot/prompt.prompt.md.j2", fmt.Sprintf(".github/prompts/%s.prompt.md", skill["name"]), map[string]any{"skill": skill}); err != nil {
				return nil, err
			}
		}
	}

	if contains("openhands") {
		if err := add("openhands", "openhands/AGENTS.md.j2", ".agentic/openhands/AGENTS.md", nil); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "openhands", "path": ".agentic/openhands/AGENTS.md", "description": "OpenHands instructions and capability overview"})
		if err := add("openhands", "openhands/instructions.md.j2", ".openhands/instructions.md", nil); err != nil {
			return nil, err
		}
	}

	if contains("opencode") {
		if err := add("opencode", "opencode/AGENTS.md.j2", ".agentic/opencode/AGENTS.md", nil); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "opencode", "path": ".agentic/opencode/AGENTS.md", "description": "OpenCode instructions and capability overview"})
		if err := add("opencode", "opencode/instructions.md.j2", ".opencode/instructions.md", nil); err != nil {
			return nil, err
		}
	}

	if contains("ollama") {
		if err := add("ollama", "ollama/Modelfile.j2", ".ollama/Modelfile", nil); err != nil {
			return nil, err
		}
		if err := add("ollama", "ollama/README.md.j2", ".ollama/README.md", nil); err != nil {
			return nil, err
		}
		extra := map[string]any{"platform_label": "Ollama", "target_file": ".ollama/README.md", "target_description": "Local model configuration and usage notes"}
		if err := add("ollama", "shared/platform-reference.md.j2", ".agentic/ollama/AGENTS.md", extra); err != nil {
			return nil, err
		}
		platformDocs = append(platformDocs, map[string]any{"platform": "ollama", "path": ".agentic/ollama/AGENTS.md", "description": "Pointer to Ollama model and runtime configuration"})
	}

	if contains("generic") {
		for _, skill := range skills {
			extra := map[string]any{"skill": skill, "invocation_prefix": "$", "slash_command": false}
			if err := add("generic", "shared/SKILL.md.j2", fmt.Sprintf(".agentic/skills/%s/SKILL.md", skill["name"]), extra); err != nil {
				return nil, err
			}
		}
	}

	if len(platformDocs) > 0 {
		if err := add("shared", "shared/AGENTS.index.md.j2", "AGENTS.md", map[string]any{"platform_docs": platformDocs}); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func safeExecuteTemplate(tpl *pongo2.Template, ctx pongo2.Context) (content string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("template execution panic: %v", recovered)
		}
	}()
	return tpl.Execute(ctx)
}
