---
name: release-readiness-reviewer
description: Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.
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

# Release Readiness Reviewer

## Purpose

Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$release-readiness-reviewer`.

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

- Identify release scope across code, config, schema, infrastructure, dependencies, docs, and operations.
- Review rollout, rollback, compatibility, migration, feature flag, and monitoring plans before release notes polish.
- Treat irreversible migrations, public API changes, and security-sensitive changes as high risk.

## Skill-Specific Checklist

- Scope, version/tag target, included changes, excluded changes, and dependencies are clear.
- Tests, builds, scans, migrations, smoke checks, and release notes have evidence or documented gaps.
- Rollback, data recovery, feature flags, monitoring, communication, and support readiness are covered.

## Decision Rules

- Block on no rollback path for high-impact production changes unless risk is explicitly accepted.
- Passing tests are necessary but not sufficient for migration or operational releases.
- Prefer staged rollout or feature flags when impact or compatibility is uncertain.

## Acceptance Criteria

- Release readiness verdict is clear: ready, ready with conditions, or not ready.
- Breaking changes and migration risks are explicit and communicated.
- Known gaps are assigned owners or accepted as residual risk.

## Output Requirements

- Lead with readiness verdict, blockers, and required conditions.
- List evidence reviewed: tests, builds, scans, release notes, migrations, and deployment config.
- Include rollback and monitoring checklist items.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
