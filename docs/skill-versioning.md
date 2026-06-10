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
  codex: "unknown"
  claude-code: "unknown"
  github-copilot: "unknown"
  gitlab-duo: "unknown"
  opencode: "unknown"
  openhands: "unknown"
  cursor: "unknown"
  roo-code: "unknown"
  kiro: "unknown"
  junie: "unknown"
  gemini-cli: "unknown"
  windsurf: "unknown"
  ollama: "unknown"
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
| `min_platform_version` | Minimum platform version per target. Use `"unknown"` if not known. |
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

List every target platform. If the minimum version is not known, use `"unknown"` — do not omit the entry.

```yaml
min_platform_version:
  codex: "unknown"
  claude-code: "unknown"
  gitlab-duo: ">=18.0"
```

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
6. Set `min_platform_version` for every relevant platform (use `"unknown"` if needed).
7. Include `replaces` and `supersedes` (empty if not applicable).
8. Include a machine-readable `changelog` in the frontmatter.
9. Include a `## Changelog` section in the body.
10. Ensure the frontmatter `version` matches the most recent changelog entry.

See `.agents/skills/universal-skill-creator/SKILL.md` for the full mandatory schema template.

---

## Validation

`skpm validate --strict` enforces all required frontmatter fields.  
`skpm validate --publish` additionally promotes warnings to errors.

Validation codes related to versioning:

| Code | Description |
|---|---|
| `skill_md_no_frontmatter` | SKILL.md has no YAML frontmatter block. |
| `skill_md_missing_version` | `version` field is missing or empty. |
| `skill_md_missing_since` | `since` field is missing or empty. |
| `skill_md_missing_last_modified` | `last_modified` field is missing or empty. |
| `skill_md_missing_authors` | `authors` list is missing or empty. |
| `skill_md_missing_stability` | `stability` field is missing. |
| `skill_md_invalid_stability` | `stability` is not `experimental`, `stable`, or `deprecated`. |
| `skill_md_missing_min_platform_version` | `min_platform_version` map is missing or empty. |
| `skill_md_missing_changelog` | `changelog` list is missing or empty. |
| `skill_md_missing_changelog_section` | Body has no `## Changelog` section. |
| `skill_md_stability` | `stability: deprecated` without `deprecated_since`. |
| `skill_md_version` | Frontmatter `version` does not match newest changelog entry. |
