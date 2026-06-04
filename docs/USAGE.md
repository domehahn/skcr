# Usage

This document uses the naming Skill Creator (`skcr`) consistently.

## Build the CLI

```bash
make build
```

## Install the CLI

```bash
make install
```

Alternative:

```bash
go install ./cmd/skcr
```

## Initialize for GitLab Duo

```bash
./skcr init --target . --platform "gitlab-duo" --project-name MyProject
./skcr bake default --target . --plan
./skcr bake default --target . --write
./skcr validate --target .
```

## Initialize for multiple platforms

```bash
./skcr init --target . --platform "gitlab-duo,codex,github-copilot" --project-name MyProject
./skcr bake default --target . --write
```

## Initialize with defaults (all supported platforms)

```bash
./skcr init --target . --project-name MyProject
./skcr bake default --target . --write
```

Default platforms when `--platform` and `--preset` are not provided:

- `codex`
- `github-copilot`
- `claude`
- `gitlab-duo`
- `opencode`
- `openhands`
- `ollama`

## Output layout for multi-platform bakes

- Root `AGENTS.md` is a generated platform index.
- Platform-specific instruction files are written to `.agentic/<platform>/AGENTS.md`.
- Platform-native files remain in their expected locations, for example:
  - `.github/copilot-instructions.md`
  - `.claude/skills/...`
  - `skills/...` and `.gitlab/duo/flows/...`

## Presets

```bash
./skcr init --target . --preset gitlab
./skcr init --target . --preset enterprise
./skcr init --target . --preset local-ai
./skcr init --target . --preset all
```
