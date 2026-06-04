# Configuration Reference

This reference applies to Skill Creator (`skcr`).

## `agentic.bake.yaml`

Top-level fields:

| Field | Purpose |
|---|---|
| `version` | Bakefile schema version. Current value: `1`. |
| `variables` | Values rendered into templates. |
| `targets` | Named generation plans. |

## Variables

| Variable | Purpose |
|---|---|
| `project_name` | Human-readable project name. |
| `owner_team` | Team responsible for the repository. |
| `default_language` | Preferred documentation language, for example `de` or `en`. |
| `governance_level` | `relaxed`, `standard`, or `strict`. |

## Target fields

| Field | Purpose |
|---|---|
| `description` | Human-readable target description. |
| `inherits` | List of target names to merge into this target. |
| `platforms` | Platforms to render. |
| `profiles` | Behavioral profile names used in templates. |
| `skills` | Skill names to generate/register. |
| `rules` | Governance flags rendered into instructions/rules. |
| `flows` | Flow template names to generate. |
| `model` | Local model configuration for Ollama/local AI. |

## Supported platforms

- `codex`
- `gitlab-duo`
- `github-copilot`
- `claude`
- `openhands`
- `opencode`
- `ollama`
- `generic`

Aliases:

- `gitlab` -> `gitlab-duo`
- `copilot` -> `github-copilot`
- `claude-code` -> `claude`

## Init with comma-separated platforms

```bash
./skcr init --target . --platform "gitlab-duo,codex,github-copilot"
```

## Init defaults

If neither `--platform` nor `--preset` is provided, `init` configures all supported runtime platforms:

- `codex`
- `github-copilot`
- `claude`
- `gitlab-duo`
- `opencode`
- `openhands`
- `ollama`
