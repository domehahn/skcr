---
name: iac-gitops-reviewer
description: Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.
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

# IaC GitOps Reviewer

## Purpose

Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.

## When to use

Use this skill when the task matches the description above or when the central agent instructions route work to `$iac-gitops-reviewer`.

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

- Inventory environments, modules, state/backends, clusters, namespaces, charts, overlays, controllers, and promotion paths.
- Prioritize destructive actions, privilege boundaries, production changes, and drift-prone resources.
- Review desired state, reconciliation, manual break-glass paths, secrets handling, and rollback together.

## Skill-Specific Checklist

- Backends, workspaces, namespaces, and resource ownership are clear.
- Deletion, replacement, privilege escalation, and production promotion require gates.
- Secrets are referenced through approved systems and not committed in manifests, variables, plans, or rendered output.

## Decision Rules

- Block on committed secrets, public exposure, destructive production changes without approval, or broad IAM grants.
- Treat generated plans as evidence only when workspace, variables, command, and target environment are known.
- Do not recommend manual cluster changes as the primary fix for GitOps-managed resources.

## Acceptance Criteria

- Environment and promotion boundaries are named or marked unknown.
- Findings include affected resource/module/chart and unsafe behavior.
- Validation steps include plan, render, diff, policy, or reconciliation checks as appropriate.

## Output Requirements

- Group findings by environment or IaC subsystem.
- Include resource identifiers, file paths, expected controls, and validation evidence.
- Separate blockers from hardening recommendations.

## Changelog

### 1.1.0 - 2026-06-10

- Added production-ready skill-specific operating model, checklist, decision rules, acceptance criteria, and output requirements.

### 1.0.0 - 2025-01-01

- Initial release.
