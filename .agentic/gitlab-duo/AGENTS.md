# AGENTS.md

## GitLab Duo Agent Platform Instructions

Project: agentic-template-kit  
Owner team: platform-engineering  
Governance level: strict

Use these instructions for GitLab Duo Chat, Agentic Chat, Custom Agents, and Flows.

## GitLab-specific model

GitLab Duo Chat is not assumed to be a Claude-style subagent runtime. Multi-step and multi-agent automation should be represented through:

- Custom agent instructions
- Flow definitions
- Merge Request workflows
- Pipeline/test validation
- Auditable session logs and comments

## Required behavior

- Work on Merge Request branches, not protected branches.
- Never push directly to `main` or `master`.
- Prefer Merge Request comments and summaries for user-visible results.
- Use GitLab context such as issues, merge requests, diffs, pipelines, jobs, and repository files when available.
- Do not expose CI/CD variables, tokens, job secrets, or masked variables.
- Do not modify pipeline behavior without explaining the effect.

## Project-level Agent Skills

This repository defines GitLab Duo project-level skills under:

```text
skills/<skill-name>/SKILL.md
```

When slash commands are supported, invoke skills as `/skill-name`.


- `/requirements-analyst` — Analyze requirements, user stories, acceptance criteria, constraints, risks, and open questions before implementation.

- `/cost-based-planner` — Plan coding work with minimal context, relevant file selection, risk awareness, rollback, and validation strategy.

- `/architecture-reviewer` — Review architecture, module boundaries, interfaces, coupling, scalability, data flows, and technical risks.

- `/threat-modeler` — Identify assets, trust boundaries, abuse cases, attack paths, threats, and required security controls.

- `/safe-implementer` — Create or modify code, tests, configuration, and project files safely with real file changes.

- `/test-strategy-engineer` — Design and generate unit, integration, regression, security, and end-to-end test strategies.

- `/verification-reviewer` — Review diffs, validate acceptance criteria, inspect test results, and find missed requirements.

- `/security-reviewer` — Review code, CI/CD, configuration, permissions, dependencies, input validation, and DevSecOps risks.

- `/secrets-reviewer` — Detect and prevent exposure of secrets, tokens, credentials, private keys, CI variables, and sensitive logs.

- `/dependency-supply-chain-reviewer` — Review dependencies, lockfiles, package managers, container images, actions, and supply-chain risks.

- `/ci-cd-reviewer` — Review CI/CD pipelines, runners, permissions, artifacts, caches, deployment gates, and token exposure.

- `/iac-gitops-reviewer` — Review Terraform, Kubernetes, Helm, Kustomize, GitOps reconciliation, promotion, and environment safety.

- `/compliance-governance-reviewer` — Review governance controls such as CODEOWNERS, branch protection, approvals, auditability, and policy compliance.

- `/release-readiness-reviewer` — Assess release readiness, rollback, migrations, feature flags, monitoring, documentation, and breaking changes.

- `/observability-reviewer` — Review logging, metrics, tracing, health checks, alerts, dashboards, runbooks, and operational readiness.

- `/incident-postmortem-assistant` — Support incident analysis, timeline creation, root cause analysis, impact assessment, corrective actions, and follow-up issues.

- `/documentation-maintainer` — Create and update README files, ADRs, setup guides, API docs, runbooks, and operational documentation.

- `/universal-skill-creator` — Create, adapt, validate, and optimize reusable agent skills across agentic platforms.




## Skill routing

Use skills proactively when the task matches their description.

For feature work, prefer:

```text
/requirements-analyst
→ /cost-based-planner
→ /architecture-reviewer when architecture is affected
→ /threat-modeler when security-sensitive behavior is affected
→ /safe-implementer
→ /test-strategy-engineer
→ /verification-reviewer
→ /security-reviewer when needed
→ /documentation-maintainer when needed
→ /release-readiness-reviewer before release readiness claims
```

For CI/CD, dependency, IaC, GitOps, secrets, compliance, observability, or incident tasks, use the matching specialized skill before claiming completion.

## Custom Flow context requirements

When a GitLab Custom Flow should use repository rules and project-level skills, the AgentComponent must pass both contexts:

```yaml
inputs:
  - from: "context:inputs.user_rule"
    as: "agents_dot_md"
    optional: true
  - from: "context:inputs.workspace_agent_skills"
    as: "workspace_agent_skills"
    optional: true
```

## Validation

For code changes, prefer repository-native checks and GitLab CI validation. For security-sensitive changes, perform a security review before suggesting merge readiness.
