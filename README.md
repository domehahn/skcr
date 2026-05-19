# Agentic Template Kit

`agentic-template-kit` is a production-oriented Python CLI for generating agentic configuration files, instructions, skill bundles, GitLab Duo customization files, Copilot instructions, Claude Code files, OpenHands/OpenCode instructions, and optional Ollama model wrappers.

It follows a **Docker Bake-like model**:

- Define reusable targets in `agentic.bake.yaml`.
- Select one or more platform adapters per target.
- Render platform-specific outputs from versioned templates.
- Apply the same setup to many repositories without copy and paste.
- Use dry-runs, validation, conflict detection, and a lockfile for controlled updates.

## Supported platform adapters

| Platform key | Generated artifacts | Notes |
|---|---|---|
| `codex` | `AGENTS.md`, `.agents/skills/<skill>/SKILL.md` | Codex project skills. |
| `gitlab-duo` | `AGENTS.md`, `skills/<skill>/SKILL.md`, `.gitlab/duo/chat-rules.md`, `.gitlab/duo/flows/*.yaml` | GitLab-native paths. Flow files are source-of-truth templates and must be created/updated in GitLab AI > Flows. |
| `github-copilot` | `.github/copilot-instructions.md`, `.github/prompts/agentic-default.prompt.md` | Repository instructions and prompt files. |
| `claude` | `CLAUDE.md`, `.claude/skills/<skill>/SKILL.md` | Claude Code project guidance and filesystem skills. |
| `openhands` | `AGENTS.md`, `.openhands/instructions.md` | OpenHands-oriented project instructions. |
| `opencode` | `AGENTS.md`, `.opencode/instructions.md` | OpenCode-oriented project instructions. |
| `ollama` | `.ollama/Modelfile`, `.ollama/README.md` | Optional local model wrapper. Ollama itself does not read repo instruction files automatically. |
| `generic` | `.agentic/generic/SKILL.md` | Portable generic skill bundle placeholder. |

## Installation

From the project root:

```bash
uv tool install .
```

For local development:

```bash
uv sync --extra dev
uv run agentic-template --help
```

## Quickstart

Create a starter bake file in your target repository:

```bash
agentic-template init --target /path/to/repo
```

List targets:

```bash
agentic-template list-targets --target /path/to/repo
```

Preview changes:

```bash
agentic-template bake gitlab --target /path/to/repo --dry-run
```

Write changes:

```bash
agentic-template bake gitlab --target /path/to/repo --write
```

Validate generated files:

```bash
agentic-template validate --target /path/to/repo
```

## Core concepts

### Target

A `target` is a named bake recipe. It answers:

> What agentic configuration should be generated for this use case?

Examples:

- `codex`
- `copilot`
- `claude`
- `gitlab`
- `local-ai`
- `default`
- `all`

A target can include multiple platforms, profiles, skills, flows, rules, model settings, and inherited targets.

### Platform

A `platform` is the target system for generated files.

Examples:

- `gitlab-duo`
- `codex`
- `github-copilot`
- `claude`
- `openhands`
- `opencode`
- `ollama`

### Profile

A `profile` is a named rule bundle used by templates. Profiles are declarative labels such as:

- `base`
- `devsecops`
- `documentation`
- `gitlab-governance`
- `local-models`

### Skill

A `skill` is a reusable capability template, for example:

- `cost-based-planner`
- `safe-implementer`
- `verification-reviewer`
- `security-reviewer`
- `documentation-maintainer`
- `universal-skill-creator`

Different adapters render skills to different paths:

| Adapter | Skill path |
|---|---|
| Codex | `.agents/skills/<skill>/SKILL.md` |
| GitLab Duo | `skills/<skill>/SKILL.md` |
| Claude Code | `.claude/skills/<skill>/SKILL.md` |

### Flow

A `flow` is currently GitLab Duo-specific in this kit. Flow templates are rendered to:

```text
.gitlab/duo/flows/<flow>.yaml
```

These files are **not automatically activated by GitLab from the repository**. They are rendered as source-of-truth templates for GitLab Custom Flows and must be created or updated in GitLab through AI > Flows / AI Catalog.

## Example `agentic.bake.yaml`

```yaml
version: "1"

variables:
  project_name: CoachIQ
  owner_team: platform-engineering
  default_language: de
  governance_level: strict

targets:
  codex:
    description: Codex AGENTS.md and project skills
    platforms:
      - codex
    profiles:
      - base
      - devsecops
    skills:
      - cost-based-planner
      - safe-implementer
      - verification-reviewer
      - security-reviewer
      - documentation-maintainer

  copilot:
    description: GitHub Copilot repository instructions and prompt files
    platforms:
      - github-copilot
    profiles:
      - base
      - devsecops
      - documentation

  claude:
    description: Claude Code CLAUDE.md and project skills
    platforms:
      - claude
    profiles:
      - base
      - devsecops
    skills:
      - cost-based-planner
      - safe-implementer
      - verification-reviewer
      - security-reviewer
      - documentation-maintainer

  gitlab:
    description: GitLab Duo Agent Platform setup with AGENTS.md, project-level skills, custom rules, and flow templates
    platforms:
      - gitlab-duo
    profiles:
      - base
      - gitlab-governance
      - devsecops
      - documentation
    skills:
      - cost-based-planner
      - safe-implementer
      - verification-reviewer
      - security-reviewer
      - documentation-maintainer
      - universal-skill-creator
    rules:
      no_direct_push: true
      require_merge_request: true
      require_tests: true
      require_security_review: true
      forbid_secret_files: true
      forbid_env_file_access: true
      require_diff_summary: true
      require_validation_summary: true
      allow_autonomous_changes: false
    flows:
      - secure-code-change
      - documentation-review
      - ci-cd-review
      - dependency-review
      - security-policy-review

  local-ai:
    description: Local Ollama/OpenCode/OpenHands setup
    platforms:
      - opencode
      - openhands
      - ollama
    profiles:
      - base
      - local-models
    model:
      provider: ollama
      default_model: qwen2.5-coder:7b
      base_url: http://localhost:11434

  default:
    description: Standard daily-development setup
    inherits:
      - codex
      - copilot

  all:
    description: Generate all supported platform artifacts
    inherits:
      - codex
      - copilot
      - claude
      - gitlab
      - local-ai
```

## GitLab Duo output

For `platforms: [gitlab-duo]`, the kit now renders GitLab-native project-level artifacts:

```text
AGENTS.md
skills/<skill-name>/SKILL.md
.gitlab/duo/chat-rules.md
.gitlab/duo/flows/<flow>.yaml
.gitlab/duo/flows/README.md
.agentic-template.lock
```

Important distinctions:

- `skills/<skill-name>/SKILL.md` is the GitLab project-level Agent Skill path.
- `.gitlab/duo/chat-rules.md` is the GitLab project-level custom rules path.
- `.gitlab/duo/flows/*.yaml` are source-of-truth templates; GitLab does not treat these as automatically active Custom Flows by merely committing them.
- Flow YAML includes optional inputs for `context:inputs.user_rule` and `context:inputs.workspace_agent_skills`, so AGENTS.md and project skills can be passed into custom flow agents.

## Safety model

The CLI is safe by default:

- `bake` does not write unless `--write` is set.
- Existing unmanaged files are not overwritten unless `--force` is set.
- A `.agentic-template.lock` file records managed files, checksums, targets, platforms, and template-pack version.
- `diff` and `dry-run` show planned changes before write.
- `validate` checks expected structures, core metadata, GitLab-native paths, and GitLab flow restrictions.

## Typical workflow

```bash
agentic-template init --target .
agentic-template bake gitlab --target . --dry-run
agentic-template bake gitlab --target . --write
agentic-template validate --target .
git diff
```

## Development

```bash
uv sync --extra dev
uv run pytest
uv run ruff check .
```

## Documentation

- `docs/CONFIGURATION.md` explains every `agentic.bake.yaml` parameter.
- `docs/USAGE.md` contains CLI examples.
- `docs/ARCHITECTURE.md` explains the renderer, adapters, lockfile, and validation model.
