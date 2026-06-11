---
name: cost-based-planner
description: Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.
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

# Cost Based Planner

## Purpose

Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$cost-based-planner`.

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

- Start from the user goal and identify the smallest useful set of files, commands, and decisions needed to reduce uncertainty.
- Budget context deliberately: read high-signal entry points, tests, configs, and recent diffs before broad exploration.
- Split work into reversible steps with validation after risky changes.

## Skill-Specific Checklist

- Scope, non-goals, constraints, and affected surfaces are explicit.
- Relevant files are selected by references, tests, ownership, execution path, and risk.
- Rollback and validation are matched to data, security, compatibility, and user-facing impact.

## Decision Rules

- Ask clarification only when repository context cannot support a safe assumption.
- Read more before editing shared contracts, auth, persistence, release, or generated output.
- Do not minimize context so aggressively that a high-risk dependency is missed.

## Acceptance Criteria

- The plan states what will change, what will not change, and how success will be checked.
- High-risk areas have mitigation, validation, or explicit out-of-scope treatment.
- The plan can be converted directly into implementation steps.

## Output Requirements

- Provide concise ordered steps with status for ongoing work.
- Name anchor files/directories and validation commands.
- Include assumptions, blockers, and rollback notes.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
