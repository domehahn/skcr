---
name: universal-skill-creator
description: Create, adapt, validate, and optimize reusable agent skills across agentic platforms.
version: "1.1.1"
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
  - version: "1.1.1"
    date: "2026-06-10"
    change: "Explicitly require applying the versioning schema while creating every new skill"
  - version: "1.1.0"
    date: "2026-06-10"
    change: "Mandatory versioning schema for generated skills: full YAML frontmatter, stability, min_platform_version, changelog"
  - version: "1.0.0"
    date: "2025-01-01"
    change: "Initial release"
---

# Universal Skill Creator

## Purpose

Create, adapt, validate, and optimize reusable agent skills across agentic platforms.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$universal-skill-creator`.

## Operating model

1. Clarify the goal and constraints.
2. Inspect the minimum relevant repository context.
3. When creating a new skill, apply the mandatory versioning schema before drafting the body content.
4. Produce a concise execution plan for non-trivial work.
5. Execute with tools when implementation is requested.
6. Validate the result with repository-native checks, including the versioning schema checklist below.
7. Summarize changed files, validation results, and residual risks.

## DevSecOps guardrails

- Do not read secrets, `.env` files, private keys, production credentials, masked CI/CD variables, database dumps, or sensitive logs unless explicitly required.
- Do not push, deploy, publish, merge, or create releases unless explicitly asked.
- Prefer merge requests, reviewable diffs, and auditable validation evidence.
- Prefer least privilege, minimal changes, and explicit rollback notes.
- Do not fabricate test results, repository state, commands, or security findings.

## Mandatory versioning schema for new skills

Every skill created by this skill MUST apply this versioning schema. Do not create, scaffold, or finalize a new skill unless its `SKILL.md` includes a complete YAML frontmatter block and a `## Changelog` body section.

### Required YAML frontmatter

```yaml
---
name: <skill-name>
description: <one-line description>
version: "0.1.0"
since: "YYYY-MM-DD"
last_modified: "YYYY-MM-DD"
authors:
  - <team-or-person>
stability: experimental
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
  - version: "0.1.0"
    date: "YYYY-MM-DD"
    change: "Initial release"
---
```

### Versioning rules

- Use SemVer without a leading `v` (e.g. `1.0.0`, `0.1.0`, `2.1.0-beta.1`).
- New skills start at `0.1.0` if experimental or `1.0.0` if immediately stable.
- Set `stability: experimental` for skills that are still being validated.
- Set `stability: stable` for skills that are ready for production use.
- Set `stability: deprecated` and fill `deprecated_since` when a skill is retired.
- Set `since` and `last_modified` to the current date in `YYYY-MM-DD` format.
- Set `min_platform_version` for every target platform supported by the repository. Use `"unknown"` if the minimum version is not known — do not omit configured platforms.
- Set `replaces` if this skill directly succeeds a previously named skill.
- Set `supersedes` if this skill fachlich replaces multiple older skills.
- The `changelog` in the frontmatter is machine-readable. Keep it in sync with the `## Changelog` body section.
- The most recent changelog entry always appears first.

### Required body section

Every skill MUST also have a `## Changelog` section in the markdown body:

```markdown
## Changelog

### 0.1.0 - YYYY-MM-DD

- Initial release.
```

Add a new entry whenever the skill's instructions, operating model, or guardrails change materially.

### Validation

Before finalising a new or updated skill, verify:

1. YAML frontmatter is present and parseable.
2. `name` matches the skill directory name.
3. `version` is valid SemVer without a leading `v`.
4. `since` and `last_modified` are in `YYYY-MM-DD` format.
5. `stability` is one of `experimental`, `stable`, `deprecated`.
6. `deprecated_since` is set when `stability: deprecated`.
7. `min_platform_version` is present and includes every configured target platform.
8. `changelog` in frontmatter has at least one entry matching `version`.
9. `## Changelog` section exists in the body.
10. Frontmatter `version` matches the most recent body changelog entry.

## Output

Provide:

- Actions taken
- Files changed or reviewed
- Validation performed
- Findings or risks
- Recommended next step

## Changelog

### 1.1.1 - 2026-06-10

- Made schema application an explicit required step in the skill creation workflow.
- Expanded the required frontmatter example to include all currently configured target platforms.
- Tightened validation so `min_platform_version` must cover every configured target platform.

### 1.1.0 - 2026-06-10

- Added mandatory versioning schema section for all newly created skills.
- Added explicit rules for `stability`, `since`, `last_modified`, `min_platform_version`, `replaces`, `supersedes`, and `changelog`.
- Added validation checklist to ensure generated skills are spec-compliant.

### 1.0.0 - 2025-01-01

- Initial release.
