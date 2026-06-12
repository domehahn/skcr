---
name: verification-reviewer
description: Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.
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

# Verification Reviewer

## Purpose

Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$verification-reviewer`.

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

- Start from acceptance criteria, then compare implementation, tests, docs, config, and generated output against them.
- Inspect diffs for behavioral regressions, missing edge cases, and unintended scope expansion.
- Treat test output as evidence only when command, environment, and result are known.

## Skill-Specific Checklist

- Every requested behavior and non-goal is accounted for.
- Changes are scoped, coherent, and free of unrelated churn.
- Tests, generated files, lockfiles, platform outputs, and validation evidence match the claimed result.

## Decision Rules

- Lead with findings; do not bury blockers under summary.
- Treat missing validation for high-risk changes as a finding.
- Avoid style findings unless they affect correctness, maintainability, or governance.

## Acceptance Criteria

- Each acceptance criterion is satisfied, unsatisfied, or unverifiable with evidence.
- Findings include severity, file/path reference, and remediation direction.
- Residual risks and test gaps are explicit.

## Output Requirements

- Findings first, ordered by severity with file references.
- Then list open questions, assumptions, validation commands/results, and acceptance status.
- End with a concise change summary only after issues are covered.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
