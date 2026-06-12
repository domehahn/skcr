---
name: incident-postmortem-assistant
description: Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.
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

# Incident Postmortem Assistant

## Purpose

Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$incident-postmortem-assistant`.

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

- Collect facts from incident notes, alerts, logs, deploy history, tickets, and communications without assigning blame.
- Build a timestamped timeline separating detection, impact, mitigation, recovery, and follow-up.
- Turn contributing factors into corrective actions with owner, priority, verification, and due date where available.

## Skill-Specific Checklist

- Impact includes affected users/systems, duration, severity, and customer or business effect.
- Timeline entries have timestamps, sources, and confidence levels.
- Causes distinguish direct trigger, contributing technical factors, and process/operational gaps.

## Decision Rules

- Use blameless language focused on system conditions and decision context.
- Do not claim root cause without evidence; mark hypotheses and missing data clearly.
- Avoid including secrets, customer identifiers, or sensitive logs in final output.

## Acceptance Criteria

- The postmortem explains what happened, why, how it was fixed, and how recurrence will be reduced.
- Corrective actions are specific, owned, and verifiable.
- Open questions are listed without blocking publication unnecessarily.

## Output Requirements

- Provide summary, impact, timeline, causes, response, corrective actions, and open questions.
- Mark assumptions, approximations, and evidence gaps.
- Produce follow-up items that can be copied into issues or tickets.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
