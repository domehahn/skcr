# Architecture

The tool has four core responsibilities:

1. Parse `agentic.bake.yaml`
2. Resolve targets and inheritance
3. Render platform-specific templates
4. Validate generated project configuration

Platform adapters are implemented in the renderer by mapping one platform to its native output paths.

## Output path model

| Platform | Main instruction file | Capability files |
|---|---|---|
| Codex | `AGENTS.md` | `.agents/skills/<skill>/SKILL.md` |
| GitLab Duo | `AGENTS.md` | `skills/<skill>/SKILL.md`, `.gitlab/duo/chat-rules.md`, `.gitlab/duo/flows/*.yaml` |
| Claude Code | `CLAUDE.md` | `.claude/skills/<skill>/SKILL.md` |
| GitHub Copilot | `.github/copilot-instructions.md` | `.github/prompts/*.prompt.md` |
| OpenHands | `AGENTS.md`, `.openhands/instructions.md` | capability registry in instructions |
| OpenCode | `AGENTS.md`, `.opencode/instructions.md` | capability registry in instructions |
| Ollama | `.ollama/Modelfile` | `.ollama/README.md` |
