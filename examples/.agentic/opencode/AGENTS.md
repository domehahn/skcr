# AGENTS.md

## OpenCode Instructions

Project: agentic-template-kit

Use this repository's DevSecOps SDLC capability model.


- `requirements-analyst` — Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.

- `cost-based-planner` — Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.

- `architecture-reviewer` — Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.

- `threat-modeler` — Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.

- `safe-implementer` — Create or modify code, tests, configuration, and project files safely with real file changes.

- `test-strategy-engineer` — Design and generate unit, integration, regression, security, and end-to-end test strategies.

- `verification-reviewer` — Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.

- `security-reviewer` — Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.

- `secrets-reviewer` — Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.

- `dependency-supply-chain-reviewer` — Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.

- `ci-cd-reviewer` — Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.

- `iac-gitops-reviewer` — Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.

- `compliance-governance-reviewer` — Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.

- `release-readiness-reviewer` — Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.

- `observability-reviewer` — Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.

- `incident-postmortem-assistant` — Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.

- `documentation-maintainer` — Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.

- `universal-skill-creator` — Create, adapt, validate, and optimize reusable agent skills across agentic platforms.


Prefer terminal-native, minimal, validated changes. Do not expose secrets.
