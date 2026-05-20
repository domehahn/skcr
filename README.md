# Agentic Template Kit - DevSecOps SDLC Edition

Generate agentic configuration for Codex, GitLab Duo Agent Platform, GitHub Copilot, Claude Code, OpenHands, OpenCode, Ollama, and generic agentic systems.

This edition includes a full DevSecOps SDLC skill set and platform-specific render targets.

## Included SDLC Skills

### Planning and analysis

- `requirements-analyst`
- `cost-based-planner`
- `architecture-reviewer`
- `threat-modeler`

### Implementation and validation

- `safe-implementer`
- `test-strategy-engineer`
- `verification-reviewer`

### Security and governance

- `security-reviewer`
- `secrets-reviewer`
- `dependency-supply-chain-reviewer`
- `ci-cd-reviewer`
- `iac-gitops-reviewer`
- `compliance-governance-reviewer`

### Delivery and operations

- `release-readiness-reviewer`
- `observability-reviewer`
- `incident-postmortem-assistant`

### Knowledge and reuse

- `documentation-maintainer`
- `universal-skill-creator`

## Quickstart

```bash
uv sync

uv run agentic-template init \
  --target /path/to/repo \
  --platform "gitlab-duo,codex,github-copilot" \
  --project-name MyProject

uv run agentic-template bake default --target /path/to/repo --dry-run
uv run agentic-template bake default --target /path/to/repo --write
uv run agentic-template validate --target /path/to/repo
```

## GitLab Duo only

```bash
uv run agentic-template init --target /path/to/repo --platform "gitlab-duo" --project-name MyProject
uv run agentic-template bake default --target /path/to/repo --write
```

Generated GitLab structure:

```text
AGENTS.md
skills/<skill-name>/SKILL.md
.gitlab/duo/chat-rules.md
.gitlab/duo/flows/*.yaml
.gitlab/duo/flows/README.md
.agentic-template.lock
```

Flow YAML files are source-of-truth templates. They must be created or updated in GitLab AI > Flows to become active.

## Documentation

- `docs/CONFIGURATION.md`
- `docs/GITLAB_DUO.md`
- `docs/SDLC_SKILLS.md`
- `docs/USAGE.md`
- `docs/ARCHITECTURE.md`
