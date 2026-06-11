---
name: observability-reviewer
description: Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.
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

# Observability Reviewer

## Purpose

Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$observability-reviewer`.

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

- Identify critical user journeys, services, dependencies, jobs, and failure modes before reviewing telemetry.
- Follow the signal chain from instrumentation to storage, dashboards, alerts, ownership, and runbooks.
- Prioritize actionable indicators: symptoms, saturation, errors, latency, freshness, and business impact.

## Skill-Specific Checklist

- Logs are structured, correlated, sampled where needed, and free of secrets or sensitive payloads.
- Metrics cover SLIs, error rates, latency, throughput, saturation, queue/freshness, and dependency health.
- Alerts are actionable, routed to owners, backed by runbooks, and tied to user impact.

## Decision Rules

- Block on logging secrets or sensitive customer data.
- Treat dashboard-only monitoring as insufficient for urgent user-impacting failures.
- Do not recommend alerting on every error; tie alerts to symptoms, budgets, or actionable thresholds.

## Acceptance Criteria

- Critical paths have enough telemetry to detect, diagnose, and verify recovery.
- Alert findings include owner, severity, threshold rationale, and response guidance.
- Privacy, retention, cardinality, and cost risks are considered where relevant.

## Output Requirements

- Summarize logs, metrics, traces, alerts, dashboards, and runbooks coverage.
- List findings with service/path references and operational impact.
- Include validation steps such as sample log review, synthetic checks, or alert dry runs.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
