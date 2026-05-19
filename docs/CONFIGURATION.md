# Configuration Reference

This document explains all supported `agentic.bake.yaml` parameters.

## Top-level schema

```yaml
version: "1"
variables: {}
targets: {}
```

## `version`

Required. Currently only `"1"` is supported.

```yaml
version: "1"
```

## `variables`

Optional global variables available to all templates and inherited by all targets.

Common variables:

| Variable | Meaning | Example |
|---|---|---|
| `project_name` | Human-readable project name | `CoachIQ` |
| `owner_team` | Owning team | `platform-engineering` |
| `default_language` | Default documentation language | `de` |
| `governance_level` | Governance strictness | `standard`, `strict` |

Example:

```yaml
variables:
  project_name: CoachIQ
  owner_team: platform-engineering
  default_language: de
  governance_level: strict
```

If `project_name` is left as `"{{ project_name }}"`, the CLI replaces it with the detected repository directory name.

## `targets`

A mapping of named bake targets.

```yaml
targets:
  gitlab:
    platforms:
      - gitlab-duo
```

A target is a reusable recipe. It can be platform-oriented (`gitlab`, `codex`, `claude`), use-case-oriented (`devsecops`, `documentation`), or broad (`default`, `all`).

## Target fields

### `description`

Optional human-readable description shown by `list-targets`.

```yaml
description: GitLab Duo Agent Platform setup
```

### `inherits`

Optional list of parent targets. Parent values are merged first; the current target overlays them.

```yaml
default:
  inherits:
    - codex
    - copilot
```

Merge behavior:

| Field | Merge behavior |
|---|---|
| `platforms` | append, preserve first occurrence |
| `profiles` | append, preserve first occurrence |
| `skills` | append, preserve first occurrence |
| `flows` | append, preserve first occurrence |
| `rules` | shallow merge, child overrides parent |
| `variables` | shallow merge, child overrides parent |
| `model` | child replaces parent |

### `platforms`

List of platform adapters to render.

Supported values:

| Platform | Output |
|---|---|
| `codex` | `AGENTS.md`, `.agents/skills/<skill>/SKILL.md` |
| `gitlab-duo` | `AGENTS.md`, `skills/<skill>/SKILL.md`, `.gitlab/duo/chat-rules.md`, `.gitlab/duo/flows/*.yaml` |
| `github-copilot` | `.github/copilot-instructions.md`, `.github/prompts/*.prompt.md` |
| `claude` | `CLAUDE.md`, `.claude/skills/<skill>/SKILL.md` |
| `openhands` | `AGENTS.md`, `.openhands/instructions.md` |
| `opencode` | `AGENTS.md`, `.opencode/instructions.md` |
| `ollama` | `.ollama/Modelfile`, `.ollama/README.md` |
| `generic` | `.agentic/generic/SKILL.md` |

### `profiles`

Labels that influence rendered content. Templates can branch on profiles or simply list them as policy context.

Examples:

```yaml
profiles:
  - base
  - devsecops
  - documentation
  - gitlab-governance
  - local-models
```

### `skills`

List of skill templates to render.

Available built-in skills:

```yaml
skills:
  - cost-based-planner
  - safe-implementer
  - verification-reviewer
  - security-reviewer
  - documentation-maintainer
  - universal-skill-creator
```

Adapter-specific skill paths:

| Adapter | Path |
|---|---|
| `codex` | `.agents/skills/<skill>/SKILL.md` |
| `gitlab-duo` | `skills/<skill>/SKILL.md` |
| `claude` | `.claude/skills/<skill>/SKILL.md` |

### `flows`

List of GitLab Duo flow templates to render. Currently used by `gitlab-duo`.

Built-in flows:

```yaml
flows:
  - secure-code-change
  - documentation-review
  - ci-cd-review
  - dependency-review
  - security-policy-review
```

Generated path:

```text
.gitlab/duo/flows/<flow>.yaml
```

These YAML files are rendered as source-of-truth templates. They must be created or updated in GitLab AI > Flows / AI Catalog to become active Custom Flows.

### `rules`

Arbitrary governance flags consumed by templates.

Common rule flags:

| Rule | Meaning |
|---|---|
| `no_direct_push` | Instruct agents not to push directly to protected branches. |
| `require_merge_request` | Require MR workflow for code changes. |
| `require_tests` | Require tests/checks where practical. |
| `require_security_review` | Require security review for sensitive changes. |
| `forbid_secret_files` | Forbid reading secret-like files. |
| `forbid_env_file_access` | Forbid `.env` access unless explicitly approved. |
| `require_diff_summary` | Require final diff summary. |
| `require_validation_summary` | Require final validation summary. |
| `allow_autonomous_changes` | Whether autonomous edits are allowed. |

Example:

```yaml
rules:
  no_direct_push: true
  require_merge_request: true
  require_tests: true
  require_security_review: true
  forbid_secret_files: true
  forbid_env_file_access: true
  allow_autonomous_changes: false
```

### `model`

Optional model metadata used by local-AI templates, especially `ollama`.

```yaml
model:
  provider: ollama
  default_model: qwen2.5-coder:7b
  base_url: http://localhost:11434
```

Fields:

| Field | Meaning |
|---|---|
| `provider` | Logical model provider, for example `ollama`. |
| `default_model` | Default local or remote model identifier. |
| `base_url` | Base URL used by local tools, such as `http://localhost:11434`. |

### `variables`

Target-local variables. These override top-level `variables`.

```yaml
targets:
  gitlab:
    variables:
      governance_level: strict
```

## Recommended GitLab Duo target

```yaml
targets:
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
```

Generated output:

```text
AGENTS.md
skills/<skill-name>/SKILL.md
.gitlab/duo/chat-rules.md
.gitlab/duo/flows/<flow>.yaml
.gitlab/duo/flows/README.md
.agentic-template.lock
```
