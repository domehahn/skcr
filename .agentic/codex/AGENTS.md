# AGENTS.md

## Codex Instructions

Project: agentic-template-kit  
Owner team: platform-engineering  
Governance level: strict

Use these instructions for Codex work in this repository.

## Required behavior

- Make real file changes when implementation is requested.
- Do not only explain how to do the work.
- Never push, deploy, publish, merge, or create pull requests unless explicitly asked.
- Do not read secrets, `.env` files, private keys, tokens, credentials, database dumps, or production logs unless explicitly required.
- Prefer minimal, targeted, reviewable diffs.
- Run repository-native validation when practical.
- Summarize changed files, checks run, and residual risks.

## Project Skills

Codex project skills are stored under:

```text
.agents/skills/<skill-name>/SKILL.md
```

Use skills explicitly with `$skill-name` when helpful.


- `$requirements-analyst` — Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.

- `$cost-based-planner` — Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.

- `$architecture-reviewer` — Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.

- `$threat-modeler` — Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.

- `$safe-implementer` — Create or modify code, tests, configuration, and project files safely with real file changes.

- `$test-strategy-engineer` — Design and generate unit, integration, regression, security, and end-to-end test strategies.

- `$verification-reviewer` — Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.

- `$security-reviewer` — Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.

- `$secrets-reviewer` — Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.

- `$dependency-supply-chain-reviewer` — Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.

- `$ci-cd-reviewer` — Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.

- `$iac-gitops-reviewer` — Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.

- `$compliance-governance-reviewer` — Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.

- `$release-readiness-reviewer` — Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.

- `$observability-reviewer` — Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.

- `$incident-postmortem-assistant` — Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.

- `$documentation-maintainer` — Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.

- `$universal-skill-creator` — Create, adapt, validate, and optimize reusable agent skills across agentic platforms.




## Skill routing

For feature work, prefer:

```text
$requirements-analyst
→ $cost-based-planner
→ $architecture-reviewer when architecture is affected
→ $threat-modeler when security-sensitive behavior is affected
→ $safe-implementer
→ $test-strategy-engineer
→ $verification-reviewer
→ $security-reviewer when needed
→ $documentation-maintainer when needed
→ $release-readiness-reviewer before release readiness claims
```

For CI/CD, dependency, IaC, GitOps, secrets, compliance, observability, or incident tasks, use the matching specialized skill before claiming completion.
