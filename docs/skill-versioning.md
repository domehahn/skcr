# Skill Versioning Schema

## Why versioning matters

Agents, platforms, and reviewers load skills at runtime without access to the
repository history. Without machine-readable metadata, they cannot determine:

- which version of a skill is loaded,
- whether a skill is experimental, stable, or deprecated,
- which platform or agent version is required,
- whether a skill supersedes an older one,
- what changed between versions.

Every `SKILL.md` therefore carries a mandatory YAML frontmatter block and a
`## Changelog` section in the body.

---

## Frontmatter schema

```yaml
---
name: safe-implementer
description: Create or modify code, tests, and configuration safely.
version: "1.0.0"
since: "2025-01-01"
last_modified: "2026-06-10"
authors:
  - platform-engineering
stability: stable
min_platform_version:
  codex: "0.51.0"
  claude-code: "1.0.44"
  github-copilot: "1.300.0"
  gitlab-duo: "18.0"
  opencode: "0.6.0"
  openhands: "0.39.0"
  cursor: "1.0.0"
  roo-code: "3.20.0"
  kiro: "0.2.0"
  junie: "2025.2"
  gemini-cli: "0.1.12"
  windsurf: "1.9.0"
  ollama: "0.7.0"
deprecated_since:
replaces:
supersedes: []
changelog:
  - version: "1.0.0"
    date: "2025-01-01"
    change: "Initial release"
---
```

### Required fields

| Field | Description |
|---|---|
| `name` | Unique skill name. Must match the skill directory name. |
| `version` | Current version. SemVer without a leading `v` (e.g. `1.0.0`). |
| `since` | Date of first release in `YYYY-MM-DD` format. |
| `last_modified` | Date of the last material change in `YYYY-MM-DD` format. |
| `authors` | List of responsible teams or persons (at least one entry). |
| `stability` | One of `experimental`, `stable`, or `deprecated`. |
| `min_platform_version` | Minimum platform version per target. Built-in production skills must use concrete versions from `internal/platforms/compatibility.go`. |
| `changelog` | Machine-readable list of changes. Most recent entry first. |

### Optional fields

| Field | Description |
|---|---|
| `deprecated_since` | Date the skill was deprecated (`YYYY-MM-DD`). Required when `stability: deprecated`. |
| `replaces` | Name of the single older skill this skill directly succeeds. |
| `supersedes` | List of skills that are fachlich replaced by this skill. |

---

## `stability` values

| Value | Meaning |
|---|---|
| `experimental` | Skill is being validated. May change without notice. |
| `stable` | Skill is ready for production use. |
| `deprecated` | Skill is retired. Set `deprecated_since`. Point users to a replacement via `replaces` or `supersedes`. |

**Default for new skills:** `experimental`.  
**Promote to `stable`** once the skill has been validated in at least one production context.

---

## `min_platform_version`

List every target platform. Built-in production skills must use the central compatibility matrix in `internal/platforms/compatibility.go`; do not hand-edit per-skill values or use `"unknown"` for stable built-ins.

```yaml
min_platform_version:
  codex: "0.51.0"
  claude-code: "1.0.44"
  gitlab-duo: "18.0"
```

`unknown` is allowed only as an incomplete custom-skill marker. The validator reports it as a warning because production routing cannot treat it as verified compatibility.

---

## `replaces` and `supersedes`

Use `replaces` when this skill is the direct one-to-one successor of a single older skill:

```yaml
replaces: old-safe-implementer
```

Use `supersedes` when this skill fachlich absorbs multiple older skills:

```yaml
supersedes:
  - legacy-implementation-helper
  - old-code-reviewer
```

Leave both empty when not applicable:

```yaml
replaces:
supersedes: []
```

---

## Frontmatter `changelog`

The frontmatter `changelog` is machine-readable. The most recent entry always comes first and must match the `version` field:

```yaml
changelog:
  - version: "1.1.0"
    date: "2026-06-10"
    change: "Added versioning schema enforcement"
  - version: "1.0.0"
    date: "2025-01-01"
    change: "Initial release"
```

---

## Body `## Changelog` section

Every `SKILL.md` must also contain a human-readable `## Changelog` section below the frontmatter:

```markdown
## Changelog

### 1.1.0 - 2026-06-10

- Added mandatory YAML frontmatter version metadata.
- Added stability and platform compatibility metadata.

### 1.0.0 - 2025-01-01

- Initial release.
```

Add a new entry whenever skill instructions, the operating model, or guardrails change materially. The version in the most recent body entry must match the frontmatter `version` field.

---

## Versioning rules

- SemVer without a leading `v`: `1.0.0`, `0.1.0`, `2.1.0-beta.1`.
- New skills start at `0.1.0` (experimental) or `1.0.0` (immediately stable).
- Increment `MINOR` for backwards-compatible additions.
- Increment `MAJOR` for breaking changes to the skill's interface or behaviour.
- Update `last_modified` on every material change.
- Keep `since` fixed after first release.

---

## `universal-skill-creator` requirements

When the `$universal-skill-creator` skill generates a new skill, it MUST:

1. Include a complete YAML frontmatter block in `SKILL.md`.
2. Set `version` to `"0.1.0"` (experimental) or `"1.0.0"` (stable).
3. Set `stability` explicitly.
4. Set `since` and `last_modified` to the creation date (`YYYY-MM-DD`).
5. Set `authors` (at least one entry).
6. Set `min_platform_version` for every relevant platform using the central compatibility matrix; use `"unknown"` only for incomplete custom-skill drafts.
7. Include `replaces` and `supersedes` (empty if not applicable).
8. Include a machine-readable `changelog` in the frontmatter.
9. Include a `## Changelog` section in the body.
10. Ensure the frontmatter `version` matches the most recent changelog entry.

See `.agents/skills/universal-skill-creator/SKILL.md` for the full mandatory schema template.

---

## Validation

Run repository validation after editing skills:

```bash
skcr validate
```

The validator scans every configured platform skill directory and enforces:

| Rule | Description |
|---|---|
| Frontmatter present | `SKILL.md` starts with a YAML frontmatter block. |
| `name` | Present, non-empty, and matching the skill directory name. |
| `version` | Present and valid SemVer without a leading `v`. |
| `since` | Present and formatted as `YYYY-MM-DD`. |
| `last_modified` | Present, formatted as `YYYY-MM-DD`, and not older than the newest changelog entry. |
| `authors` | Present with at least one entry. |
| `stability` | Present and one of `experimental`, `stable`, or `deprecated`. |
| `deprecated_since` | Present and formatted as `YYYY-MM-DD` when `stability: deprecated`. |
| `min_platform_version` | Present with at least one platform entry; stable skills must not use `unknown`. |
| `replaces` | Present, even when empty. |
| `supersedes` | Present, using `[]` when empty. |
| Frontmatter `changelog` | Present with at least one entry; newest entry must match `version`. |
| Body `## Changelog` | Present with at least one version entry; newest entry must match `version`. |
