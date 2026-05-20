# Usage

## Initialize for GitLab Duo

```bash
agentic-template init --target . --platform "gitlab-duo" --project-name MyProject
agentic-template bake default --target . --dry-run
agentic-template bake default --target . --write
agentic-template validate --target .
```

## Initialize for multiple platforms

```bash
agentic-template init --target . --platform "gitlab-duo,codex,github-copilot" --project-name MyProject
agentic-template bake default --target . --write
```

## Presets

```bash
agentic-template init --target . --preset gitlab
agentic-template init --target . --preset enterprise
agentic-template init --target . --preset local-ai
agentic-template init --target . --preset all
```
