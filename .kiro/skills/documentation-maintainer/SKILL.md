---
name: documentation-maintainer
description: Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.
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

## Output

Provide:

- Actions taken
- Files changed or reviewed
- Validation performed
- Findings or risks
- Recommended next step

## Changelog

### 1.0.0 - 2025-01-01

- Initial release.
