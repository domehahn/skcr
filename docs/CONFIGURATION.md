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
| `delivery` | Optional delivery intent such as `skills`, `commands`, or `both` for future profile-aware rendering. |
| `skills` | Skill names to generate/register. |
| `rules` | Governance flags rendered into instructions/rules. |
| `flows` | Flow template names to generate. |
| `model` | Local model configuration for Ollama/local AI. |
| `gitlab_duo` | GitLab Duo-specific options, e.g. `slash_command`. |

## Generated skill directory layout

`skcr` creates each skill as a directory with `SKILL.md` at the root. It also materializes the Agent Skills optional resource directories so tools can discover the expected paths immediately:

```text
skill-name/
├── SKILL.md
├── scripts/
├── references/
├── assets/
└── tests/
```

`scripts/`, `references/`, and `assets/` match the Agent Skills specification. `tests/` is an additional skcr lifecycle directory for examples, fixtures, and expected outputs.

## Supported platforms

- `codex`
- `gitlab-duo`
- `github-copilot`
- `claude`
- `openhands`
- `opencode`
- `ollama`
- `antigravity`
- `amazon-q`
- `cline`
- `kilocode`
- `qoder`
- `qwen`
- `generic`

Aliases:

- `gitlab` -> `gitlab-duo`
- `copilot` -> `github-copilot`
- `claude-code` -> `claude`
- `amazon`, `amazon-q-developer` -> `amazon-q`
- `kilo`, `kilo-code` -> `kilocode`
- `qwen-code` -> `qwen`

`skcr` keeps a central capability matrix for skill and command surfaces inspired by OpenSpec-style tool integrations. New tool IDs can be accepted as skills-first targets even before minimum compatible platform versions are validated; in that case generated `min_platform_version` entries stay `"unknown"`.

## `agentic.compatibility.yaml`

`agentic.compatibility.yaml` is optional repository-local compatibility evidence. It is created by `skcr compatibility set` and contains only verified overrides for platform minimum versions. `skcr bake` merges it with the built-in compatibility matrix before rendering `min_platform_version`.

```yaml
platforms:
  - name: codex
    min_version: 0.51.0
    status: verified
    source: local-evidence
    evidence: docs/compat/codex-0.51.0.md
    validated: "2026-06-12"
```

The evidence path must exist. Concrete minimum versions without evidence are rejected by `skcr compatibility check`.

## Init with comma-separated platforms

```bash
./skcr init --target . --platform "gitlab-duo,codex,github-copilot"
```

## Init defaults

If neither `--platform` nor `--preset` is provided, `init` configures every known concrete platform from the platform model. Use `--platform` to intentionally narrow the bakefile, for example:

```bash
./skcr init --target . --platform "codex,gitlab-duo"
```

## GitLab Duo options

To control slash-command metadata in generated GitLab skills:

```yaml
targets:
  gitlab:
    gitlab_duo:
      slash_command: false
```

Default is `true`.
