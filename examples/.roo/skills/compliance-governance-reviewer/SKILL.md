---
name: compliance-governance-reviewer
description: Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.
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

# Compliance Governance Reviewer

## Purpose

Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$compliance-governance-reviewer`.

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

- Review CODEOWNERS, branch protection, required checks, approval rules, release policy, auditability, and exception paths.
- Compare documented policy with enforceable repository, CI, and platform controls.
- Look for bypasses through generated files, bot accounts, admin overrides, unprotected branches, or direct tag pushes.

## Skill-Specific Checklist

- Critical paths have accountable owners and suitable required reviewers.
- Security, dependency, infrastructure, and release changes require review and auditable evidence.
- Policy text, branch protections, required checks, and generated output rules do not contradict each other.

## Decision Rules

- Treat unenforced policy text as advisory until backed by a repository, CI, or platform control.
- Do not invent compliance requirements; tie findings to repository policy, governance level, or common control expectations.
- Prefer explicit exception workflows over informal bypasses.

## Acceptance Criteria

- Present, missing, and locally unverifiable governance controls are named.
- Each finding links policy risk to a concrete control or documentation fix.
- No recommendation weakens existing security or review requirements.

## Output Requirements

- Summarize ownership, approvals, checks, branch/tag protection, and release gates.
- List findings with affected files/settings and expected control.
- Identify which items require platform-admin configuration.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
