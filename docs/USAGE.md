# Usage

## Initialize a repository

```bash
agentic-template init --target /path/to/repo
```

## Preview GitLab Duo output

```bash
agentic-template bake gitlab --target /path/to/repo --dry-run
```

Expected files:

```text
AGENTS.md
skills/<skill-name>/SKILL.md
.gitlab/duo/chat-rules.md
.gitlab/duo/flows/<flow>.yaml
.gitlab/duo/flows/README.md
.agentic-template.lock
```

## Apply GitLab Duo output

```bash
agentic-template bake gitlab --target /path/to/repo --write
```

## Apply Codex output

```bash
agentic-template bake codex --target /path/to/repo --write
```

## Apply Claude Code output

```bash
agentic-template bake claude --target /path/to/repo --write
```

## Apply local AI output

```bash
agentic-template bake local-ai --target /path/to/repo --write
```

This generates OpenCode, OpenHands, and Ollama-oriented files.

## Validate

```bash
agentic-template validate --target /path/to/repo
```

## Force overwrite conflicts

```bash
agentic-template bake gitlab --target /path/to/repo --write --force
```

Use this only when you intentionally want to replace unmanaged or locally modified generated files.
