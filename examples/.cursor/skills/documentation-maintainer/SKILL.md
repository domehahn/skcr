---
name: documentation-maintainer
description: Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.
version: "1.1.0"
since: "2026-06-10"
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
  - version: "1.1.0"
    date: "2026-06-10"
    change: "Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements"
  - version: "1.0.0"
    date: "2025-01-01"
    change: "Initial release"
---

# Documentation Maintainer

## Purpose

Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$documentation-maintainer`.

## Operating model

1. Clarify the goal and constraints.
2. Inspect the minimum relevant repository context.
3. Produce a concise execution plan for non-trivial work.
4. Execute with tools when implementation is requested.
5. Validate the result with repository-native checks.
6. Summarize changed files, validation results, and residual risks.

## DevSecOps guardrails

- Do not read secrets, `.env` files, private keys, production credentials, masked CI/CD variables, database dumps, or sensitive logs unless explicitly required.
- Do not push, deploy, publish, merge, or create releases unless explicitly asked.
- Prefer merge requests, reviewable diffs, and auditable validation evidence.
- Prefer least privilege, minimal changes, and explicit rollback notes.
- Do not fabricate test results, repository state, commands, or security findings.

## Skill-Specific Operating Model

- Identify the audience first: contributor, operator, reviewer, integrator, maintainer, or end user.
- Verify documentation against current repository behavior, commands, config, generated files, and governance rules.
- Prefer task-oriented docs with prerequisites, commands, expected outcomes, validation, and troubleshooting.

## Skill-Specific Checklist

- Purpose, repository structure, source-of-truth rules, and manual-edit boundaries are clear.
- Commands are current, copyable, and include required flags or paths.
- Security-sensitive instructions avoid exposing secrets or encouraging unsafe shortcuts.

## Decision Rules

- Do not document behavior as supported unless code, config, or tests show it.
- When desired source-of-truth differs from actual CLI behavior, document the actual behavior and the migration gap.
- Replace generic marketing language with concrete workflow guidance.

## Acceptance Criteria

- A reader can complete the documented workflow without hidden context.
- Canonical files, generated copies, and synchronized outputs are identified.
- Unknown platform or environment requirements are marked explicitly.

## Output Requirements

- Summarize documentation files changed and why.
- Call out verified commands or repository evidence.
- List remaining documentation gaps separately from completed updates.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
