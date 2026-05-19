# Architecture

`agentic-template-kit` has four core parts:

1. Bake-file parser
2. Target resolver
3. Platform renderer
4. Validator and lockfile manager

## Bake-file parser

Reads `agentic.bake.yaml` and validates it with Pydantic models.

## Target resolver

Resolves `inherits` relationships and merges platforms, profiles, skills, flows, rules, variables, and model settings.

## Platform renderer

The renderer maps platform keys to concrete files.

| Platform | Main output logic |
|---|---|
| `codex` | Root `AGENTS.md` plus `.agents/skills/<skill>/SKILL.md`. |
| `gitlab-duo` | Root `AGENTS.md`, GitLab project skills under `skills/`, custom rules under `.gitlab/duo/chat-rules.md`, custom flow templates under `.gitlab/duo/flows/*.yaml`. |
| `github-copilot` | `.github/copilot-instructions.md` and prompt files. |
| `claude` | `CLAUDE.md` and `.claude/skills/<skill>/SKILL.md`. |
| `openhands` | Root `AGENTS.md` plus `.openhands/instructions.md`. |
| `opencode` | Root `AGENTS.md` plus `.opencode/instructions.md`. |
| `ollama` | `.ollama/Modelfile` and `.ollama/README.md`. |
| `generic` | Portable placeholder `SKILL.md`. |

When multiple platforms render the same destination, such as `AGENTS.md`, the renderer merges the platform contributions into one file with generated boundary comments.

## Lockfile

`.agentic-template.lock` records managed files and checksums. This allows safe updates while detecting local modifications.

## Validator

The validator checks generated files and key platform-specific constraints. For GitLab Duo it checks native paths and custom-flow constraints, including:

- project skills under `skills/<skill>/SKILL.md`
- custom rules under `.gitlab/duo/chat-rules.md`
- no legacy `.gitlab/duo/custom-rules.md` or `.gitlab/duo/rules/**`
- flow YAML does not contain forbidden top-level fields
- flow `environment` is `ambient`
- flow prompts do not contain a `model` field
- flow AgentComponents receive `context:inputs.user_rule` and `context:inputs.workspace_agent_skills`
