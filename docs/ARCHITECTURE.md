# Architecture

Skill Creator (`skcr`) has four core responsibilities:

1. Parse `agentic.bake.yaml`
2. Resolve targets and inheritance
3. Render platform-specific templates
4. Validate generated project configuration

Platform adapters are implemented in the renderer by mapping one platform to its native output paths.

## Output path model

| Platform | Main instruction file | Capability files |
|---|---|---|
| Shared | `AGENTS.md` | Platform index for active targets |
| Codex | `.agentic/codex/AGENTS.md` | `.agents/skills/<skill>/SKILL.md` |
| GitLab Duo | `.agentic/gitlab-duo/AGENTS.md` | `skills/<skill>/SKILL.md`, `.gitlab/duo/chat-rules.md`, `.gitlab/duo/flows/*.yaml` |
| Claude Code | `.agentic/claude/AGENTS.md` | `.claude/skills/<skill>/SKILL.md`, `.claude/agents/*.md` |
| GitHub Copilot | `.github/copilot-instructions.md` | `.github/prompts/*.prompt.md` |
| OpenHands | `.agentic/openhands/AGENTS.md`, `.openhands/instructions.md` | capability registry in instructions |
| OpenCode | `.agentic/opencode/AGENTS.md`, `.opencode/instructions.md` | capability registry in instructions |
| Ollama | `.agentic/ollama/AGENTS.md`, `.ollama/Modelfile` | `.ollama/README.md` |
