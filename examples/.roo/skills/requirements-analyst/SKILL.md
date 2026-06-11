---
name: requirements-analyst
description: Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.
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

# Requirements Analyst

## Purpose

Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$requirements-analyst`.

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

- Extract the actual user goal, affected actors, expected behavior, and non-goals from the request and repository context.
- Translate ambiguous asks into observable acceptance criteria before implementation starts.
- Identify constraints from governance, security, compatibility, data migration, localization, and platform support.

## Skill-Specific Checklist

- Actors, workflows, inputs, outputs, files, commands, APIs, or generated artifacts are named.
- Acceptance criteria cover normal, edge, and failure paths relevant to the request.
- Risks, assumptions, dependencies, and decisions that could invalidate the work are visible.

## Decision Rules

- Do not ask broad clarification questions when repository evidence can answer safely.
- Treat missing acceptance criteria as a requirements gap, not permission to invent behavior silently.
- Flag conflicts between requested behavior and existing governance or platform rules.

## Acceptance Criteria

- Requirements are specific enough for implementation and review.
- Open questions are limited to decisions that materially affect outcome or risk.
- Constraints and non-goals are explicit for the implementer.

## Output Requirements

- Summarize goal, scope, non-goals, assumptions, and constraints.
- Provide testable acceptance criteria.
- List risks and open questions separately, with recommended next action.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
