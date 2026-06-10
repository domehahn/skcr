# GitHub Copilot Repository Instructions

Project: agentic-template-kit  
Owner team: platform-engineering  
Governance level: strict

GitHub Copilot does not use `SKILL.md` project skills like Codex or GitLab Duo. This repository provides reusable prompt files under:

```text
.github/prompts/*.prompt.md
```

## Available prompt capabilities


- `requirements-analyst.prompt.md` — Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.

- `cost-based-planner.prompt.md` — Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.

- `architecture-reviewer.prompt.md` — Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.

- `threat-modeler.prompt.md` — Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.

- `safe-implementer.prompt.md` — Create or modify code, tests, configuration, and project files safely with real file changes.

- `test-strategy-engineer.prompt.md` — Design and generate unit, integration, regression, security, and end-to-end test strategies.

- `verification-reviewer.prompt.md` — Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.

- `security-reviewer.prompt.md` — Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.

- `secrets-reviewer.prompt.md` — Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.

- `dependency-supply-chain-reviewer.prompt.md` — Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.

- `ci-cd-reviewer.prompt.md` — Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.

- `iac-gitops-reviewer.prompt.md` — Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.

- `compliance-governance-reviewer.prompt.md` — Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.

- `release-readiness-reviewer.prompt.md` — Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.

- `observability-reviewer.prompt.md` — Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.

- `incident-postmortem-assistant.prompt.md` — Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.

- `documentation-maintainer.prompt.md` — Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.

- `universal-skill-creator.prompt.md` — Create, adapt, validate, and optimize reusable agent skills across agentic platforms.


## Required behavior

- Prefer minimal, reviewable changes.
- Do not expose secrets or credentials.
- Do not change CI/CD or security posture without explaining the impact.
- Run or suggest repository-native validation.
- Summarize changed files, validation, and risks.
