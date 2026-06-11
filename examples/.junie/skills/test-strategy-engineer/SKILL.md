---
name: test-strategy-engineer
description: Design and generate unit, integration, regression, security, and end-to-end test strategies.
version: "1.1.0"
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
  - version: "1.1.0"
    date: "2026-06-10"
    change: "Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements"
  - version: "1.0.0"
    date: "2025-01-01"
    change: "Initial release"
---

# Test Strategy Engineer

## Purpose

Design and generate unit, integration, regression, security, and end-to-end test strategies.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$test-strategy-engineer`.

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

- Map behavior, risk, and change surface before selecting unit, integration, contract, end-to-end, smoke, or security tests.
- Prefer fast deterministic tests first, adding broader tests only where integration or workflow risk requires them.
- Convert bug reports, requirements, and risky boundaries into regression tests where possible.

## Skill-Specific Checklist

- Happy paths, edge cases, failure paths, permissions, compatibility, and migration behavior are covered as relevant.
- Fixtures, mocks, isolation, cleanup, determinism, and sensitive data handling are planned.
- Existing frameworks, naming patterns, CI runtime, and flake controls are respected.

## Decision Rules

- Do not add brittle end-to-end tests for logic that can be covered reliably at lower levels.
- Prioritize tests around changed behavior and historically risky paths.
- Do not use production secrets or live customer data in tests.

## Acceptance Criteria

- Each major risk maps to at least one validation method.
- Suggested or added tests fit repository tooling and conventions.
- Remaining coverage gaps and CI/runtime tradeoffs are explicit.

## Output Requirements

- Provide a concise test matrix by behavior and test level.
- Name specific files, commands, frameworks, fixtures, and mocks.
- Report tests added or recommended plus remaining gaps.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
